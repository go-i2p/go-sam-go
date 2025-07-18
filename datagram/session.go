package datagram

import (
	"net"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewDatagramSession creates a new datagram session for UDP-like I2P messaging.
// This function establishes a new datagram session with the provided SAM connection,
// session ID, cryptographic keys, and configuration options. It returns a DatagramSession
// instance that can be used for sending and receiving datagrams over the I2P network.
// Example usage: session, err := NewDatagramSession(sam, "my-session", keys, []string{"inbound.length=1"})
func NewDatagramSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	// Log session creation with detailed parameters for debugging
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating new DatagramSession")

	// Create the base session using the common package for session management
	// This handles the underlying SAM protocol communication and session establishment
	session, err := sam.NewGenericSession("DATAGRAM", id, keys, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create datagram session: %w", err)
	}

	// Ensure the session is of the correct type for datagram operations
	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	// Initialize the datagram session with the base session and configuration
	ds := &DatagramSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created DatagramSession")
	return ds, nil
}

// NewReader creates a DatagramReader for receiving datagrams from any source.
// This method initializes a new reader with buffered channels for asynchronous datagram
// reception. The reader must be started manually with receiveLoop() for continuous operation.
// Example usage: reader := session.NewReader(); go reader.receiveLoop(); datagram, err := reader.ReceiveDatagram()
func (s *DatagramSession) NewReader() *DatagramReader {
	// Create reader with buffered channels for non-blocking operation
	// The buffer size of 10 prevents blocking when multiple datagrams arrive rapidly
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

// NewWriter creates a DatagramWriter for sending datagrams to specific destinations.
// This method initializes a new writer with a default timeout of 30 seconds for send operations.
// The timeout can be customized using the SetTimeout method on the returned writer.
// Example usage: writer := session.NewWriter().SetTimeout(60*time.Second); err := writer.SendDatagram(data, dest)
func (s *DatagramSession) NewWriter() *DatagramWriter {
	// Initialize writer with default timeout for send operations
	// The timeout prevents indefinite blocking on send operations
	return &DatagramWriter{
		session: s,
		timeout: 30, // Default timeout in seconds
	}
}

// PacketConn returns a net.PacketConn interface for this session.
// This method provides compatibility with standard Go networking code by wrapping
// the datagram session in a connection that implements the PacketConn interface.
// Example usage: conn := session.PacketConn(); n, addr, err := conn.ReadFrom(buffer)
func (s *DatagramSession) PacketConn() net.PacketConn {
	// Create a PacketConn wrapper with integrated reader and writer
	// This provides standard Go networking interface compliance
	return &DatagramConn{
		session: s,
		reader:  s.NewReader(),
		writer:  s.NewWriter(),
	}
}

// SendDatagram sends a datagram to the specified destination address.
// This is a convenience method that creates a temporary writer and sends the datagram
// immediately. For multiple sends, it's more efficient to create a writer once and reuse it.
// Example usage: err := session.SendDatagram([]byte("hello"), destinationAddr)
func (s *DatagramSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	// Use a temporary writer for one-time send operations
	// This simplifies the API for simple send operations
	return s.NewWriter().SendDatagram(data, dest)
}

// ReceiveDatagram receives a single datagram from any source.
// This is a convenience method that creates a temporary reader, starts the receive loop,
// gets one datagram, and cleans up resources automatically. For continuous reception,
// use NewReader() and manage the reader lifecycle manually.
// Example usage: datagram, err := session.ReceiveDatagram()
func (s *DatagramSession) ReceiveDatagram() (*Datagram, error) {
	// Create temporary reader for one-time receive operations
	reader := s.NewReader()
	// Start the receive loop for datagram processing
	go reader.receiveLoop()
	return reader.ReceiveDatagram()
}

// Close closes the datagram session and all associated resources.
// This method safely terminates the session, closes the underlying connection,
// and cleans up any background goroutines. It's safe to call multiple times.
// Example usage: defer session.Close()
func (s *DatagramSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	// Log session closure for debugging and monitoring
	logger := log.WithField("id", s.ID())
	logger.Debug("Closing DatagramSession")

	s.closed = true

	// Close the underlying base session to terminate SAM communication
	// This ensures proper cleanup of the I2P connection
	if err := s.BaseSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close base session")
		return oops.Errorf("failed to close datagram session: %w", err)
	}

	logger.Debug("Successfully closed DatagramSession")
	return nil
}

// Addr returns the I2P address of this session.
// This address represents the session's identity on the I2P network and can be
// used by other nodes to send datagrams to this session. The address is derived
// from the session's cryptographic keys.
// Example usage: myAddr := session.Addr(); fmt.Println("My I2P address:", myAddr.Base32())
func (s *DatagramSession) Addr() i2pkeys.I2PAddr {
	// Return the I2P address derived from the session's cryptographic keys
	return s.Keys().Addr()
}

// Network returns the network type for this address.
// This method implements the net.Addr interface and always returns "i2p-datagram"
// to identify this as an I2P datagram address type for networking compatibility.
// Example usage: network := addr.Network() // returns "i2p-datagram"
func (a *DatagramAddr) Network() string {
	// Return the network type identifier for I2P datagram addresses
	return "i2p-datagram"
}

// String returns the string representation of the address.
// This method implements the net.Addr interface and returns the Base32 encoded
// representation of the I2P address for human-readable display and logging.
// Example usage: addrStr := addr.String() // returns "abcd1234...xyz.b32.i2p"
func (a *DatagramAddr) String() string {
	// Return the Base32 encoded I2P address for human-readable representation
	return a.addr.Base32()
}
