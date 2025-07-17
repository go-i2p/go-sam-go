package raw

import (
	"net"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewRawSession creates a new raw session
func NewRawSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating new RawSession")

	// Create the base session using the common package
	session, err := sam.NewGenericSession("RAW", id, keys, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create raw session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	rs := &RawSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created RawSession")
	return rs, nil
}

// NewReader creates a RawReader for receiving raw datagrams
func (s *RawSession) NewReader() *RawReader {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil
	}

	return &RawReader{
		session:   s,
		recvChan:  make(chan *RawDatagram, 10), // Buffer for incoming datagrams
		errorChan: make(chan error, 1),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
		closed:    false,
		mu:        sync.RWMutex{},
	}
}

// ...existing code...

// NewWriter creates a RawWriter for sending raw datagrams
func (s *RawSession) NewWriter() *RawWriter {
	return &RawWriter{
		session: s,
		timeout: 30, // Default timeout in seconds
	}
}

// PacketConn returns a net.PacketConn interface for this session
func (s *RawSession) PacketConn() net.PacketConn {
	return &RawConn{
		session: s,
		reader:  s.NewReader(),
		writer:  s.NewWriter(),
	}
}

// SendDatagram sends a raw datagram to the specified destination
func (s *RawSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	return s.NewWriter().SendDatagram(data, dest)
}

// ReceiveDatagram receives a raw datagram from any source
func (s *RawSession) ReceiveDatagram() (*RawDatagram, error) {
	reader := s.NewReader()
	if reader == nil {
		return nil, oops.Errorf("session is closed")
	}

	// Start the receive loop for this reader
	go reader.receiveLoop()

	// Get one datagram and then close the reader to clean up the goroutine
	datagram, err := reader.ReceiveDatagram()
	reader.Close()

	return datagram, err
}

// Close closes the raw session and all associated resources
func (s *RawSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Closing RawSession")

	s.closed = true

	// Close the base session
	if err := s.BaseSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close base session")
		return oops.Errorf("failed to close raw session: %w", err)
	}

	logger.Debug("Successfully closed RawSession")
	return nil
}

// Addr returns the I2P address of this session
func (s *RawSession) Addr() i2pkeys.I2PAddr {
	return s.Keys().Addr()
}

// Network returns the network type
func (a *RawAddr) Network() string {
	return "i2p-raw"
}

// String returns the string representation of the address
func (a *RawAddr) String() string {
	return a.addr.Base32()
}
