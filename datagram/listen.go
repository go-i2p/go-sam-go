package datagram

import (
	"net"

	"github.com/samber/oops"
)

func (s *DatagramSession) Listen() (*DatagramListener, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, oops.Errorf("session is closed")
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Creating PacketListener")

	listener := &DatagramListener{
		session:    s,
		reader:     s.NewReader(),
		acceptChan: make(chan net.Conn, 10), // Buffer for incoming connections
		errorChan:  make(chan error, 1),
		closeChan:  make(chan struct{}),
	}

	// Start accepting packet connections in a goroutine
	go listener.acceptLoop()

	logger.Debug("Successfully created PacketListener")
	return listener, nil
}
