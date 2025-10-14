package sam3

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-i2p/i2pkeys"
)

// TestListener manages a local I2P listener for testing purposes.
// It provides a stable, local destination that can replace external sites in tests.
type TestListener struct {
	sam      *SAM
	session  *StreamSession
	listener *StreamListener
	addr     i2pkeys.I2PAddr
	closed   bool
	mu       sync.RWMutex
}

// TestListenerConfig holds configuration for creating test listeners.
type TestListenerConfig struct {
	SessionID    string
	HTTPResponse string // Optional custom HTTP response content
	Timeout      time.Duration
}

// DefaultTestListenerConfig returns a default configuration for test listeners.
func DefaultTestListenerConfig(sessionID string) *TestListenerConfig {
	return &TestListenerConfig{
		SessionID:    sessionID,
		HTTPResponse: "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html><body><h1>Test I2P Site</h1><p>This is a test response from a local I2P listener.</p></body></html>",
		Timeout:      5 * time.Minute, // I2P tunnels can take time to establish
	}
}

// SetupTestListener creates and starts a local I2P listener that can serve as a test destination.
// This replaces the need for external sites like i2p-projekt.i2p or idk.i2p in tests.
// The listener will respond to HTTP GET requests with basic HTML content.
func SetupTestListener(t *testing.T, config *TestListenerConfig) *TestListener {
	t.Helper()

	if config == nil {
		config = DefaultTestListenerConfig("test_listener")
	}

	sam := createSAMConnection(t)
	keys := generateListenerKeys(t, sam)
	session := createStreamSession(t, sam, config.SessionID, keys)
	listener := createListener(t, session, sam)

	testListener := initializeTestListener(sam, session, listener, keys)
	go testListener.serve(t, config.HTTPResponse)

	waitForListenerReady(t, testListener, config.Timeout)

	t.Logf("Test listener ready at %s", testListener.addr.Base32())
	return testListener
}

// createSAMConnection establishes a SAM connection for the test listener.
func createSAMConnection(t *testing.T) *SAM {
	sam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		t.Fatalf("Failed to create SAM connection for test listener: %v", err)
	}
	return sam
}

// generateListenerKeys creates cryptographic keys for the test listener identity.
func generateListenerKeys(t *testing.T, sam *SAM) i2pkeys.I2PKeys {
	keys, err := sam.NewKeys()
	if err != nil {
		sam.Close()
		t.Fatalf("Failed to generate keys for test listener: %v", err)
	}
	return keys
}

// createStreamSession establishes a minimal 1-hop I2P session for faster testing.
func createStreamSession(t *testing.T, sam *SAM, sessionID string, keys i2pkeys.I2PKeys) *StreamSession {
	session, err := sam.NewStreamSession(sessionID, keys, []string{
		"inbound.length=1",
		"outbound.length=1",
		"inbound.lengthVariance=0",
		"outbound.lengthVariance=0",
		"inbound.quantity=1",
		"outbound.quantity=1",
	})
	if err != nil {
		sam.Close()
		t.Fatalf("Failed to create stream session for test listener: %v", err)
	}
	return session
}

// createListener establishes an I2P listener on the session for accepting connections.
func createListener(t *testing.T, session *StreamSession, sam *SAM) *StreamListener {
	listener, err := session.Listen()
	if err != nil {
		session.Close()
		sam.Close()
		t.Fatalf("Failed to create listener for test listener: %v", err)
	}
	return listener
}

// initializeTestListener constructs the TestListener structure with all required components.
func initializeTestListener(sam *SAM, session *StreamSession, listener *StreamListener, keys i2pkeys.I2PKeys) *TestListener {
	return &TestListener{
		sam:      sam,
		session:  session,
		listener: listener,
		addr:     keys.Addr(),
	}
}

// waitForListenerReady blocks until the listener is accepting connections or times out.
func waitForListenerReady(t *testing.T, testListener *TestListener, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := testListener.waitForReady(ctx, t); err != nil {
		testListener.Close()
		t.Fatalf("Test listener failed to become ready: %v", err)
	}
}

// Addr returns the I2P address of the test listener.
func (tl *TestListener) Addr() i2pkeys.I2PAddr {
	return tl.addr
}

// AddrString returns the Base32 address string of the test listener.
func (tl *TestListener) AddrString() string {
	return tl.addr.Base32()
}

// serve handles incoming connections to the test listener.
func (tl *TestListener) serve(t *testing.T, httpResponse string) {
	for {
		tl.mu.RLock()
		if tl.closed {
			tl.mu.RUnlock()
			return
		}
		tl.mu.RUnlock()

		conn, err := tl.listener.Accept()
		if err != nil {
			tl.mu.RLock()
			closed := tl.closed
			tl.mu.RUnlock()
			if !closed {
				t.Logf("Test listener accept error: %v", err)
			}
			return
		}

		// Handle connection in goroutine to support multiple concurrent requests
		go tl.handleConnection(conn, httpResponse, t)
	}
}

// handleConnection processes a single connection to the test listener.
func (tl *TestListener) handleConnection(conn net.Conn, httpResponse string, t *testing.T) {
	defer conn.Close()

	// Read the request (we expect HTTP GET)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Logf("Test listener read error: %v", err)
		return
	}

	request := string(buf[:n])
	t.Logf("Test listener received request: %s", strings.ReplaceAll(request, "\n", "\\n"))

	// Send the configured HTTP response
	_, err = conn.Write([]byte(httpResponse))
	if err != nil {
		t.Logf("Test listener write error: %v", err)
	}
}

// waitForReady waits for the test listener to be available for connections.
// This implements proper I2P timing considerations where tunnel establishment can take time.
func (tl *TestListener) waitForReady(ctx context.Context, t *testing.T) error {
	clientSession, cleanup, err := tl.createTestClient()
	if err != nil {
		return err
	}
	defer cleanup()

	return tl.retryConnection(ctx, t, clientSession)
}

// createTestClient creates a test client SAM connection and session for verifying listener readiness.
// Returns the session, a cleanup function, and any error encountered.
func (tl *TestListener) createTestClient() (*StreamSession, func(), error) {
	clientSAM, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test client SAM: %w", err)
	}

	clientKeys, err := clientSAM.NewKeys()
	if err != nil {
		clientSAM.Close()
		return nil, nil, fmt.Errorf("failed to generate test client keys: %w", err)
	}

	clientSession, err := clientSAM.NewStreamSession("test_client_"+tl.session.ID(), clientKeys, []string{
		"inbound.length=1",
		"outbound.length=1",
		"inbound.lengthVariance=0",
		"outbound.lengthVariance=0",
		"inbound.quantity=1",
		"outbound.quantity=1",
	})
	if err != nil {
		clientSAM.Close()
		return nil, nil, fmt.Errorf("failed to create test client session: %w", err)
	}

	cleanup := func() {
		clientSession.Close()
		clientSAM.Close()
	}

	return clientSession, cleanup, nil
}

// retryConnection attempts to connect to the test listener with exponential backoff.
// Returns nil on successful connection, or an error if the context times out.
func (tl *TestListener) retryConnection(ctx context.Context, t *testing.T, clientSession *StreamSession) error {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		if err := checkContextDone(ctx); err != nil {
			return err
		}

		if err := tl.attemptConnection(ctx, t, clientSession, &backoff, maxBackoff); err != nil {
			if err == errConnectionFailed {
				continue
			}
			return err
		}

		t.Logf("Test listener is ready")
		return nil
	}
}

// errConnectionFailed is a sentinel error indicating a connection attempt failed but should be retried.
var errConnectionFailed = fmt.Errorf("connection attempt failed")

// attemptConnection tries to establish a connection to the test listener.
// Returns nil on success, errConnectionFailed if the attempt should be retried,
// or a context error if the context is done.
func (tl *TestListener) attemptConnection(ctx context.Context, t *testing.T, clientSession *StreamSession, backoff *time.Duration, maxBackoff time.Duration) error {
	t.Logf("Attempting to connect to test listener...")
	conn, err := clientSession.DialI2P(tl.addr)
	if err != nil {
		t.Logf("Test listener not ready yet: %v (retrying in %v)", err, *backoff)

		if err := waitWithBackoff(ctx, backoff, maxBackoff); err != nil {
			return err
		}
		return errConnectionFailed
	}

	conn.Close()
	return nil
}

// checkContextDone checks if the context is done and returns an appropriate error.
func checkContextDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for test listener to be ready: %w", ctx.Err())
	default:
		return nil
	}
}

// waitWithBackoff waits for the current backoff duration and increases it exponentially.
// Returns an error if the context is cancelled during the wait.
func waitWithBackoff(ctx context.Context, backoff *time.Duration, maxBackoff time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for test listener to be ready: %w", ctx.Err())
	case <-time.After(*backoff):
	}

	*backoff = *backoff * 2
	if *backoff > maxBackoff {
		*backoff = maxBackoff
	}
	return nil
}

// Close shuts down the test listener and cleans up resources.
func (tl *TestListener) Close() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.closed {
		return nil
	}
	tl.closed = true

	errs := tl.collectCloseErrors()
	return formatCloseErrors(errs)
}

// collectCloseErrors attempts to close all resources and collects any errors encountered.
func (tl *TestListener) collectCloseErrors() []error {
	var errs []error

	if err := tl.closeListener(); err != nil {
		errs = append(errs, err)
	}

	if err := tl.closeSession(); err != nil {
		errs = append(errs, err)
	}

	if err := tl.closeSAM(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// closeListener closes the stream listener if it exists.
func (tl *TestListener) closeListener() error {
	if tl.listener != nil {
		if err := tl.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}
	return nil
}

// closeSession closes the stream session if it exists.
func (tl *TestListener) closeSession() error {
	if tl.session != nil {
		if err := tl.session.Close(); err != nil {
			return fmt.Errorf("failed to close session: %w", err)
		}
	}
	return nil
}

// closeSAM closes the SAM connection if it exists.
func (tl *TestListener) closeSAM() error {
	if tl.sam != nil {
		if err := tl.sam.Close(); err != nil {
			return fmt.Errorf("failed to close SAM: %w", err)
		}
	}
	return nil
}

// formatCloseErrors formats multiple close errors into a single error or returns nil.
func formatCloseErrors(errs []error) error {
	if len(errs) > 0 {
		return fmt.Errorf("multiple close errors: %v", errs)
	}
	return nil
}

// SetupTestListenerWithHTTP creates a test listener that provides HTTP-like responses
// suitable for replacing external web sites in tests.
func SetupTestListenerWithHTTP(t *testing.T, sessionID string) *TestListener {
	config := &TestListenerConfig{
		SessionID: sessionID,
		HTTPResponse: "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/html\r\n" +
			"Content-Length: 120\r\n" +
			"\r\n" +
			"<html><head><title>Test I2P Site</title></head>" +
			"<body><h1>Hello from I2P!</h1><p>This is a test response.</p></body></html>",
		Timeout: 5 * time.Minute,
	}
	return SetupTestListener(t, config)
}

// generateUniqueSessionID creates a unique session ID to prevent conflicts during concurrent test execution.
func generateUniqueSessionID(testName string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d", testName, timestamp)
}
