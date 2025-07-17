package datagram

import (
	"net"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewDatagramSession creates a new datagram session
func NewDatagramSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating new DatagramSession")

	// Create the base session using the common package
	session, err := sam.NewGenericSession("DATAGRAM", id, keys, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create datagram session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ds := &DatagramSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created DatagramSession")
	return ds, nil
}

// NewReader creates a DatagramReader for receiving datagrams
func (s *DatagramSession) NewReader() *DatagramReader {
	return &DatagramReader{
		session:   s,
		recvChan:  make(chan *Datagram, 10), // Buffer for incoming datagrams
		errorChan: make(chan error, 1),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}, 1),
		closed:    false,
		mu:        sync.RWMutex{},
		closeOnce: sync.Once{},
	}
}

// NewWriter creates a DatagramWriter for sending datagrams
func (s *DatagramSession) NewWriter() *DatagramWriter {
	return &DatagramWriter{
		session: s,
		timeout: 30, // Default timeout in seconds
	}
}

// PacketConn returns a net.PacketConn interface for this session
func (s *DatagramSession) PacketConn() net.PacketConn {
	return &DatagramConn{
		session: s,
		reader:  s.NewReader(),
		writer:  s.NewWriter(),
	}
}

// SendDatagram sends a datagram to the specified destination
func (s *DatagramSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	return s.NewWriter().SendDatagram(data, dest)
}

// ReceiveDatagram receives a datagram from any source
func (s *DatagramSession) ReceiveDatagram() (*Datagram, error) {
	reader := s.NewReader()
	// Start the receive loop
	go reader.receiveLoop()
	return reader.ReceiveDatagram()
}

// Close closes the datagram session and all associated resources
func (s *DatagramSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Closing DatagramSession")

	s.closed = true

	// Close the base session
	if err := s.BaseSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close base session")
		return oops.Errorf("failed to close datagram session: %w", err)
	}

	logger.Debug("Successfully closed DatagramSession")
	return nil
}

// Addr returns the I2P address of this session
func (s *DatagramSession) Addr() i2pkeys.I2PAddr {
	return s.Keys().Addr()
}

// Network returns the network type
func (a *DatagramAddr) Network() string {
	return "i2p-datagram"
}

// String returns the string representation of the address
func (a *DatagramAddr) String() string {
	return a.addr.Base32()
}
