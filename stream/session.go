package stream

import (
	"context"
	"runtime"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// cleanupStreamListener is called by AddCleanup to ensure the listener is closed and the goroutine is cleaned up
// This prevents goroutine leaks if the user forgets to call Close()
func cleanupStreamListener(l *StreamListener) {
	log.Warn("StreamListener garbage collected without being closed, closing now to prevent goroutine leak")
	l.Close()
}

// NewStreamSession creates a new streaming session for TCP-like I2P connections.
// It initializes the session with the provided SAM connection, session ID, cryptographic keys,
// and configuration options. The session provides both client and server capabilities for
// establishing reliable streaming connections over the I2P network.
// Example usage: session, err := NewStreamSession(sam, "my-session", keys, []string{"inbound.length=1"})
func NewStreamSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating new StreamSession")

	// Create the base session using the common package
	session, err := sam.NewGenericSession("STREAM", id, keys, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create stream session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ss := &StreamSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created StreamSession")
	return ss, nil
}

// NewStreamSessionFromSubsession creates a StreamSession for a subsession that has already been
// registered with a PRIMARY session using SESSION ADD. This constructor skips the session
// creation step since the subsession is already registered with the SAM bridge.
//
// This function is specifically designed for use with SAMv3.3 PRIMARY sessions where
// subsessions are created using SESSION ADD rather than SESSION CREATE commands.
//
// Parameters:
//   - sam: SAM connection for data operations (separate from the primary session's control connection)
//   - id: The subsession ID that was already registered with SESSION ADD
//   - keys: The I2P keys from the primary session (shared across all subsessions)
//   - options: Configuration options for the subsession
//
// Returns a StreamSession ready for use without attempting to create a new SAM session.
func NewStreamSessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating StreamSession from existing subsession")

	// Create a BaseSession manually since the session is already registered
	// We need a way to create BaseSession from the common package
	baseSession, err := common.NewBaseSessionFromSubsession(sam, id, keys)
	if err != nil {
		logger.WithError(err).Error("Failed to create base session from subsession")
		return nil, oops.Errorf("failed to create stream session from subsession: %w", err)
	}

	ss := &StreamSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created StreamSession from subsession")
	return ss, nil
}

// NewStreamSessionWithSignature creates a new streaming session with a custom signature type for TCP-like I2P connections.
// This is the package-level function version that allows specifying cryptographic signature algorithms.
// It initializes the session with the provided SAM connection, session ID, cryptographic keys,
// configuration options, and signature type. The session provides both client and server capabilities for
// establishing reliable streaming connections over the I2P network with custom cryptographic settings.
// Example usage: session, err := NewStreamSessionWithSignature(sam, "my-session", keys, []string{"inbound.length=1"}, "EdDSA_SHA512_Ed25519")
func NewStreamSessionWithSignature(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new StreamSession with signature")

	// Create the base session using the common package with signature
	session, err := sam.NewGenericSessionWithSignature("STREAM", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature")
		return nil, oops.Errorf("failed to create stream session with signature: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ss := &StreamSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created StreamSession with signature")
	return ss, nil
}

// NewStreamSessionWithSignatureAndPorts creates a new stream session with custom signature type and port configuration.
// This function provides advanced control over both cryptographic parameters and port mapping for stream sessions.
// The 'from' parameter specifies the local port binding, while 'to' specifies the target port for connections.
// Port specifications can be single ports ("80") or ranges ("8080-8090") depending on I2P router configuration.
//
// This method enables complex port forwarding scenarios and integration with existing network infrastructure
// that expects specific port mappings. It's particularly useful for applications that need to maintain
// consistent port assignments or work with legacy systems expecting fixed port numbers.
//
// Example usage:
//
//	session, err := NewStreamSessionWithSignatureAndPorts(sam, "http-proxy", "8080", "80", keys,
//	                   []string{"inbound.length=2"}, "EdDSA_SHA512_Ed25519")
func NewStreamSessionWithSignatureAndPorts(sam *common.SAM, id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"from":    from,
		"to":      to,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new StreamSession with signature and ports")

	// Create the base session using the common package with signature and port configuration
	session, err := sam.NewGenericSessionWithSignatureAndPorts("STREAM", id, from, to, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature and ports")
		return nil, oops.Errorf("failed to create stream session with signature and ports: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ss := &StreamSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created StreamSession with signature and ports")
	return ss, nil
}

// Listen creates a StreamListener that accepts incoming connections from remote I2P destinations.
// It initializes a listener with buffered channels for connection handling and starts an internal
// accept loop to manage incoming connections asynchronously. The listener provides thread-safe
// operations and properly handles session closure and resource cleanup.
// A finalizer is set on the listener to ensure that the accept loop is terminated
// if the listener is garbage collected without being closed.
// Example usage: listener, err := session.Listen(); conn, err := listener.Accept()
func (s *StreamSession) Listen() (*StreamListener, error) {
	// Check closed state with read lock, then release immediately to avoid deadlock
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	s.mu.RUnlock()

	logger := log.WithField("id", s.ID())
	logger.Debug("Creating StreamListener")

	ctx, cancel := context.WithCancel(context.Background())
	listener := &StreamListener{
		session:    s,
		acceptChan: make(chan *StreamConn, 10), // Buffer for incoming connections
		errorChan:  make(chan error, 1),
		closeChan:  make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Set up cleanup to ensure the listener is closed and the goroutine is cleaned up
	// This prevents goroutine leaks if the user forgets to call Close()
	listener.cleanup = runtime.AddCleanup(&listener.cleanup, cleanupStreamListener, listener)

	// Start accepting connections in a goroutine
	go listener.acceptLoop()

	// Register the listener with the session (using separate write lock)
	s.registerListener(listener)

	logger.Debug("Successfully created StreamListener")
	return listener, nil
}

// Accept creates a listener and accepts the next incoming connection from remote I2P destinations.
// This is a convenience method that automatically creates a listener and calls Accept() on it.
// It provides a simpler API for applications that only need to accept a single connection
// or want to handle each connection acceptance individually.
//
// For applications that need to accept multiple connections or want more control over
// listener lifecycle, use Listen() to get a StreamListener and call Accept() on it directly.
//
// Each call to Accept creates a new internal listener, so applications accepting multiple
// connections should use Listen() once and then call Accept() multiple times on the listener
// for better performance and resource management.
//
// Returns a StreamConn for the accepted connection, or an error if the acceptance fails.
// The error may be due to session closure, network issues, or I2P tunnel problems.
//
// Example usage: conn, err := session.Accept() // Simple single connection acceptance
func (s *StreamSession) Accept() (*StreamConn, error) {
	logger := log.WithField("id", s.ID())
	logger.Debug("Accepting connection via session")

	// Create a listener for this acceptance
	listener, err := s.Listen()
	if err != nil {
		logger.WithError(err).Error("Failed to create listener for Accept")
		return nil, oops.Errorf("failed to create listener for session accept: %w", err)
	}

	// Accept a connection and then close the listener
	defer func() {
		if closeErr := listener.Close(); closeErr != nil {
			logger.WithError(closeErr).Warn("Failed to close listener after Accept")
		}
	}()

	conn, err := listener.AcceptStream()
	if err != nil {
		logger.WithError(err).Error("Failed to accept connection")
		return nil, oops.Errorf("failed to accept connection: %w", err)
	}

	logger.Debug("Successfully accepted connection via session")
	return conn, nil
}

// NewDialer creates a StreamDialer for establishing outbound connections to I2P destinations.
// It initializes a dialer with a default timeout of 30 seconds, which can be customized using
// the SetTimeout method. The dialer supports both string destinations and native I2P addresses.
// Example usage: dialer := session.NewDialer().SetTimeout(60*time.Second)
func (s *StreamSession) NewDialer() *StreamDialer {
	return &StreamDialer{
		session: s,
		timeout: 30 * time.Second, // Default timeout
	}
}

// SetTimeout sets the default timeout duration for dial operations.
// This method allows customization of the connection timeout and returns the dialer
// for method chaining. The timeout applies to all subsequent dial operations.
// Example usage: dialer.SetTimeout(60*time.Second)
func (d *StreamDialer) SetTimeout(timeout time.Duration) *StreamDialer {
	d.timeout = timeout
	return d
}

// Dial establishes a connection to the specified I2P destination using the default timeout.
// This is a convenience method that creates a new dialer and establishes a connection
// to the specified destination string. For custom timeout or multiple connections,
// use NewDialer() for better performance.
// Example usage: conn, err := session.Dial("destination.b32.i2p")
func (s *StreamSession) Dial(destination string) (*StreamConn, error) {
	return s.NewDialer().Dial(destination)
}

// DialI2P establishes a connection to the specified I2P address using native addressing.
// This is a convenience method that creates a new dialer and establishes a connection
// to the specified I2P address using the i2pkeys.I2PAddr type. The method uses the
// session's default timeout settings.
// Example usage: conn, err := session.DialI2P(addr)
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error) {
	return s.NewDialer().DialI2P(addr)
}

// DialContext establishes a connection with context support for cancellation and timeout.
// This is a convenience method that creates a new dialer and establishes a connection
// to the specified destination with context-based cancellation support. The context
// can be used to cancel the connection attempt or apply custom timeouts.
// Example usage: conn, err := session.DialContext(ctx, "destination.b32.i2p")
func (s *StreamSession) DialContext(ctx context.Context, destination string) (*StreamConn, error) {
	return s.NewDialer().DialContext(ctx, destination)
}

// Close closes the streaming session and all associated resources.
// This method is safe to call multiple times and will only perform cleanup once.
// All listeners and connections created from this session will become invalid after closing.
// The method properly handles concurrent access and resource cleanup.
// Example usage: defer session.Close()
func (s *StreamSession) Close() error {
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		return nil
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Closing StreamSession")

	s.closed = true

	// Close all listeners first to stop their accept loops
	listeners := s.copyAndClearListeners()

	// CRITICAL FIX: Release the write lock BEFORE calling BaseSession.Close()
	// This prevents deadlock when Listen() operations are waiting for registerListener()
	// which needs the write lock, while BaseSession.Close() can block on network I/O
	s.mu.Unlock()

	// CRITICAL FIX: Close listeners BEFORE closing the base session
	// This ensures SetReadDeadline() in closeWithoutUnregister() unblocks any
	// pending reads in accept loops before BaseSession.Close() marks session as closed
	for _, listener := range listeners {
		listener.closeWithoutUnregister()
	}

	// Brief delay to allow read deadlines to take effect and unblock pending reads
	time.Sleep(150 * time.Millisecond)

	// Close the base session after listeners are shut down
	if err := s.BaseSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close base session")
	}

	logger.Debug("Successfully closed StreamSession")
	return nil
}

// Addr returns the I2P address of this session for identification purposes.
// This address can be used by other I2P nodes to connect to this session.
// The address is derived from the session's cryptographic keys and remains constant
// for the lifetime of the session.
// Example usage: addr := session.Addr()
func (s *StreamSession) Addr() i2pkeys.I2PAddr {
	return s.Keys().Addr()
}

// registerListener adds a listener to the session's listener list
func (s *StreamSession) registerListener(listener *StreamListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
}

// unregisterListener removes a listener from the session's listener list
func (s *StreamSession) unregisterListener(listener *StreamListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, l := range s.listeners {
		if l == listener {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			break
		}
	}
}

// copyAndClearListeners returns a copy of listeners and clears the list (must be called with mutex held)
func (s *StreamSession) copyAndClearListeners() []*StreamListener {
	listeners := make([]*StreamListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.listeners = nil
	return listeners
}
