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

	// Close the base session first to unblock any pending reads
	if err := s.BaseSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close base session")
		// Continue with listener cleanup even if base session close fails
	}

	for _, listener := range listeners {
		listener.closeWithoutUnregister()
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
