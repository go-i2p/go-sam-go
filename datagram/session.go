package datagram

import (
	"bufio"
	"encoding/base64"
	"net"
	"strings"
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

// NewDatagramSessionFromSubsession creates a DatagramSession for a subsession that has already been
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
// Returns a DatagramSession ready for use without attempting to create a new SAM session.
func NewDatagramSessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating DatagramSession from existing subsession")

	// Create a BaseSession manually since the session is already registered
	baseSession, err := common.NewBaseSessionFromSubsession(sam, id, keys)
	if err != nil {
		logger.WithError(err).Error("Failed to create base session from subsession")
		return nil, oops.Errorf("failed to create datagram session from subsession: %w", err)
	}

	ds := &DatagramSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created DatagramSession from subsession")
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
		session:     s,
		recvChan:    make(chan *Datagram, 10), // Buffer for incoming datagrams
		errorChan:   make(chan error, 1),
		closeChan:   make(chan struct{}),
		doneChan:    make(chan struct{}, 1),
		closed:      false,
		loopStarted: false,
		mu:          sync.RWMutex{},
		closeOnce:   sync.Once{},
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

// ReceiveDatagram receives a single datagram from the I2P network.
// This method is a convenience wrapper that performs a direct single read operation
// without starting a continuous receive loop. For continuous reception,
// use NewReader() and manage the reader lifecycle manually.
// Example usage: datagram, err := session.ReceiveDatagram()
func (s *DatagramSession) ReceiveDatagram() (*Datagram, error) {
	// ARCHITECTURAL FIX: Perform direct read for one-shot operations instead of
	// creating a reader with receive loop, which causes deadlocks due to RWMutex contention
	return s.readSingleDatagram()
}

// readSingleDatagram performs a direct read from the SAM connection for one-shot datagram operations.
// This method bypasses the reader infrastructure to avoid deadlocks when only one datagram is needed.
func (s *DatagramSession) readSingleDatagram() (*Datagram, error) {
	// Use the session's direct connection for immediate read
	conn := s.Conn()
	if conn == nil {
		return nil, oops.Errorf("session connection is not available")
	}

	// Read directly from SAM connection
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, oops.Errorf("failed to read from SAM connection: %w", err)
	}

	response := string(buffer[:n])
	log.WithField("response", response).Debug("Received SAM response")

	// Validate response format
	if !strings.Contains(response, "DATAGRAM RECEIVED") {
		return nil, oops.Errorf("unexpected response format: %s", response)
	}

	// Parse the response using the same logic as DatagramReader
	source, data, err := s.parseDatagramResponse(response)
	if err != nil {
		return nil, err
	}

	return s.createDatagram(source, data)
}

// parseDatagramResponse parses the DATAGRAM RECEIVED response to extract source and data.
func (s *DatagramSession) parseDatagramResponse(response string) (string, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(response))
	scanner.Split(bufio.ScanWords)

	var source, data string
	for scanner.Scan() {
		word := scanner.Text()
		switch {
		case word == "DATAGRAM" || word == "RECEIVED":
			// Skip protocol tokens
			continue
		case strings.HasPrefix(word, "DESTINATION="):
			source = word[12:]
		case strings.HasPrefix(word, "SIZE="):
			// Skip size, we'll get actual data size from payload
			continue
		default:
			// Remaining data is the base64-encoded payload
			if data == "" {
				data = word
			} else {
				data = data + " " + word
			}
		}
	}

	if source == "" {
		return "", "", oops.Errorf("no source in datagram")
	}
	if data == "" {
		return "", "", oops.Errorf("no data in datagram")
	}

	return source, data, nil
}

// createDatagram constructs the final Datagram from parsed source and data.
func (s *DatagramSession) createDatagram(source, data string) (*Datagram, error) {
	sourceAddr, err := i2pkeys.NewI2PAddrFromString(source)
	if err != nil {
		return nil, oops.Errorf("failed to parse source address: %w", err)
	}

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, oops.Errorf("failed to decode datagram data: %w", err)
	}

	datagram := &Datagram{
		Data:   decodedData,
		Source: sourceAddr,
		Local:  s.Addr(),
	}

	return datagram, nil
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
