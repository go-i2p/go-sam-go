package raw

import (
	"net"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewRawSession creates a new raw session for sending and receiving raw datagrams.
// It initializes the session with the provided SAM connection, session ID, cryptographic keys,
// and configuration options, returning a RawSession instance or an error if creation fails.
// Example usage: session, err := NewRawSession(sam, "my-session", keys, []string{"inbound.length=1"})
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

// NewReader creates a RawReader for receiving raw datagrams from any source.
// It initializes buffered channels for incoming datagrams and errors, returning nil if the session is closed.
// The caller must start the receive loop manually by calling receiveLoop() in a goroutine.
// Example usage: reader := session.NewReader(); go reader.receiveLoop()
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

// NewWriter creates a RawWriter for sending raw datagrams to specific destinations.
// It initializes the writer with a default timeout of 30 seconds for send operations.
// The timeout can be customized using the SetTimeout method on the returned writer.
// Example usage: writer := session.NewWriter().SetTimeout(60*time.Second)
func (s *RawSession) NewWriter() *RawWriter {
	return &RawWriter{
		session: s,
		timeout: 30, // Default timeout in seconds
	}
}

// PacketConn returns a net.PacketConn interface for this session.
// This provides compatibility with standard Go networking code by wrapping the session
// in a RawConn that implements the PacketConn interface for datagram operations.
// Example usage: conn := session.PacketConn(); n, addr, err := conn.ReadFrom(buf)
func (s *RawSession) PacketConn() net.PacketConn {
	conn := &RawConn{
		session: s,
		reader:  s.NewReader(),
		writer:  s.NewWriter(),
	}
	
	// Set up cleanup to prevent resource leaks if Close() is not called
	conn.addCleanup()
	
	return conn
}

// SendDatagram sends a raw datagram to the specified destination address.
// This is a convenience method that creates a temporary writer and sends the datagram immediately.
// For multiple sends, it's more efficient to create a writer once and reuse it.
// Example usage: err := session.SendDatagram(data, destAddr)
func (s *RawSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	return s.NewWriter().SendDatagram(data, dest)
}

// ReceiveDatagram receives a single raw datagram from any source.
// This is a convenience method that creates a temporary reader, starts the receive loop,
// gets one datagram, and cleans up the resources automatically.
// Example usage: datagram, err := session.ReceiveDatagram()
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

// Close closes the raw session and all associated resources.
// This method is safe to call multiple times and will only perform cleanup once.
// All readers and writers created from this session will become invalid after closing.
// Example usage: defer session.Close()
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

// Addr returns the I2P address of this session.
// This address can be used by other I2P nodes to send datagrams to this session.
// The address is derived from the session's cryptographic keys.
// Example usage: addr := session.Addr()
func (s *RawSession) Addr() i2pkeys.I2PAddr {
	return s.Keys().Addr()
}

// Network returns the network type for this address.
// This method implements the net.Addr interface and always returns "i2p-raw"
// to identify this as an I2P raw datagram address type.
// Example usage: network := addr.Network() // returns "i2p-raw"
func (a *RawAddr) Network() string {
	return "i2p-raw"
}

// String returns the string representation of the address.
// This method implements the net.Addr interface and returns the Base32 encoded
// representation of the I2P address for human-readable display.
// Example usage: addrStr := addr.String() // returns "abcd1234...xyz.b32.i2p"
func (a *RawAddr) String() string {
	return a.addr.Base32()
}
