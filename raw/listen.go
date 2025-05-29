package raw

import (
	"github.com/samber/oops"
)

func (s *RawSession) Listen() (*RawListener, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, oops.Errorf("session is closed")
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Creating RawListener")

	listener := &RawListener{
		session:    s,
		reader:     s.NewReader(),
		acceptChan: make(chan *RawConn, 10), // Buffer for incoming connections
		errorChan:  make(chan error, 1),
		closeChan:  make(chan struct{}),
	}

	// Start accepting raw connections in a goroutine
	go listener.acceptLoop()

	logger.Debug("Successfully created RawListener")
	return listener, nil
}
