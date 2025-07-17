package raw

import (
	"github.com/samber/oops"
)

// Listen creates a RawListener for accepting incoming raw connections.
// This method initializes the listener with buffered channels for incoming connections
// and starts the accept loop in a background goroutine to handle incoming datagrams.
// Example usage: listener, err := session.Listen()
func (s *RawSession) Listen() (*RawListener, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the session is already closed before creating listener
	if s.closed {
		return nil, oops.Errorf("session is closed")
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Creating RawListener")

	// Initialize listener with buffered channels for connection management
	// The acceptChan buffers incoming connections to prevent blocking
	listener := &RawListener{
		session:    s,
		reader:     s.NewReader(),
		acceptChan: make(chan *RawConn, 10), // Buffer for incoming connections
		errorChan:  make(chan error, 1),
		closeChan:  make(chan struct{}),
	}

	// Start accepting raw connections in a goroutine
	// This allows the listener to handle multiple concurrent connections
	go listener.acceptLoop()

	logger.Debug("Successfully created RawListener")
	return listener, nil
}
