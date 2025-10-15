package datagram

import (
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewDatagramSession creates a new datagram session for UDP-like I2P messaging using SAMv3 UDP forwarding.
// This function establishes a new datagram session with the provided SAM connection,
// session ID, cryptographic keys, and configuration options. It automatically creates a UDP listener
// for receiving forwarded datagrams (SAMv3 requirement) and configures the session with PORT/HOST parameters.
// V1/V2 compatibility (reading from TCP control socket) is no longer supported.
// Returns a DatagramSession instance that uses UDP forwarding for all datagram reception.
// Example usage: session, err := NewDatagramSession(sam, "my-session", keys, []string{"inbound.length=1"})
func NewDatagramSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new DatagramSession with SAMv3 UDP forwarding")

	// Create UDP listener and inject forwarding parameters
	udpConn, options, err := setupDatagramUDPForwarding(options)
	if err != nil {
		return nil, err
	}

	// Create the base session for datagram
	baseSession, err := createGenericDatagramSession(sam, id, keys, options)
	if err != nil {
		udpConn.Close()
		return nil, err
	}

	// Initialize the datagram session with UDP forwarding enabled
	ds := &DatagramSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,
	}

	log.Debug("Successfully created DatagramSession with UDP forwarding")
	return ds, nil
}

// setupDatagramUDPForwarding creates a UDP listener and injects forwarding parameters.
// This helper creates a UDP listener for SAMv3 datagram forwarding and automatically configures
// the session options to include the HOST and PORT parameters required by the SAM bridge.
// Returns the UDP connection, updated options slice, and any error encountered.
func setupDatagramUDPForwarding(options []string) (*net.UDPConn, []string, error) {
	// Create UDP listener for receiving forwarded datagrams
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.WithError(err).Error("Failed to resolve UDP address")
		return nil, nil, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.WithError(err).Error("Failed to create UDP listener")
		return nil, nil, oops.Errorf("failed to create UDP listener: %w", err)
	}

	// Get the actual port assigned by the OS
	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	log.WithField("udp_port", udpPort).Debug("Created UDP listener for datagram forwarding")

	// Inject UDP forwarding parameters into session options
	options = ensureUDPForwardingParameters(options, udpPort)

	return udpConn, options, nil
}

// createGenericDatagramSession creates a validated BaseSession for datagram.
// This helper creates the generic session through the SAM bridge using STYLE=DATAGRAM,
// validates the session type, and ensures proper cleanup on error.
func createGenericDatagramSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*common.BaseSession, error) {
	// Create the base session using DATAGRAM style
	session, err := sam.NewGenericSession("DATAGRAM", id, keys, options)
	if err != nil {
		log.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create datagram session: %w", err)
	}

	// Validate session type
	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		log.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	return baseSession, nil
}

// ensureUDPForwardingParameters injects UDP forwarding parameters into session options if not already present.
// This ensures SAMv3 UDP forwarding is configured with PORT and HOST parameters.
// PORT/HOST specify where the SAM bridge should forward datagrams TO (the client's UDP listener).
// sam.udp.port/sam.udp.host are NOT set here - they configure the SAM bridge's own UDP port (default 7655).
// This is required for all datagram sessions in v3-only mode.
func ensureUDPForwardingParameters(options []string, udpPort int) []string {
	updatedOptions := make([]string, 0, len(options)+2)

	hasPort := false
	hasHost := false

	// Check existing options
	for _, opt := range options {
		if strings.HasPrefix(opt, "PORT=") {
			hasPort = true
		} else if strings.HasPrefix(opt, "HOST=") {
			hasHost = true
		}
		updatedOptions = append(updatedOptions, opt)
	}

	// Inject missing UDP forwarding parameters
	// PORT/HOST tell SAM bridge where to forward datagrams TO (our UDP listener)
	if !hasHost {
		updatedOptions = append(updatedOptions, "HOST=127.0.0.1")
	}
	if !hasPort {
		updatedOptions = append(updatedOptions, "PORT="+strconv.Itoa(udpPort))
	}

	return updatedOptions
}

// NewDatagramSessionFromSubsession creates a DatagramSession for a subsession that has already been
// registered with a PRIMARY session using SESSION ADD. This constructor skips the session
// creation step since the subsession is already registered with the SAM bridge.
//
// This function is specifically designed for use with SAMv3.3 PRIMARY sessions where
// subsessions are created using SESSION ADD rather than SESSION CREATE commands.
//
// For PRIMARY datagram subsessions, UDP forwarding is mandatory (SAMv3 requirement).
// The UDP connection must be provided for proper datagram reception via UDP forwarding.
//
// Parameters:
//   - sam: SAM connection for data operations (separate from the primary session's control connection)
//   - id: The subsession ID that was already registered with SESSION ADD
//   - keys: The I2P keys from the primary session (shared across all subsessions)
//   - options: Configuration options for the subsession
//   - udpConn: UDP connection for receiving forwarded datagrams (required, not nil)
//
// Returns a DatagramSession ready for use without attempting to create a new SAM session.
func NewDatagramSessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, udpConn *net.UDPConn) (*DatagramSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":          id,
		"options":     options,
		"udp_enabled": udpConn != nil,
	})
	logger.Debug("Creating DatagramSession from existing subsession with SAMv3 UDP forwarding")

	if udpConn == nil {
		logger.Error("UDP connection is required for SAMv3 datagram subsessions")
		return nil, oops.Errorf("udp connection is required for datagram subsessions (v3 only)")
	}

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
		udpConn:     udpConn,
		udpEnabled:  true,
	}

	logger.Debug("Successfully created DatagramSession from subsession with UDP forwarding")
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

// readSingleDatagram performs a direct read from the UDP connection for one-shot datagram operations.
// This method bypasses the reader infrastructure to avoid deadlocks when only one datagram is needed.
// SAMv3 UDP forwarding mode only - reads from UDP connection where SAM bridge forwards datagrams.
// V1/V2 TCP control socket reading is no longer supported.
func (s *DatagramSession) readSingleDatagram() (*Datagram, error) {
	s.mu.RLock()
	udpConn := s.udpConn
	s.mu.RUnlock()

	// V3-only: Always read from UDP connection
	if udpConn == nil {
		return nil, oops.Errorf("UDP connection not available (v3 UDP forwarding required)")
	}

	return s.readDatagramFromUDP(udpConn)
}

// readDatagramFromUDP reads a forwarded datagram from the UDP connection.
// This is used for PRIMARY subsessions where datagrams are forwarded via UDP by the SAM bridge.
// Format per SAMv3.md:
//
//	Line 1: $destination (base64 I2P destination)
//	Line 2+: FROM_PORT=nnn TO_PORT=nnn (SAMv3.2+, may be on one or two lines)
//	Then: \n (empty line separator)
//	Remaining: $datagram_payload (raw data)
func (s *DatagramSession) readDatagramFromUDP(udpConn *net.UDPConn) (*Datagram, error) {
	buffer := make([]byte, 65536) // Large buffer for UDP datagrams
	n, _, err := udpConn.ReadFromUDP(buffer)
	if err != nil {
		return nil, oops.Errorf("failed to read from UDP connection: %w", err)
	}

	log.WithField("bytes_read", n).Debug("Received UDP datagram")

	// Parse the UDP datagram format per SAMv3.md
	response := string(buffer[:n])

	// Find the first newline - that's the end of the header line
	firstNewline := strings.Index(response, "\n")
	if firstNewline == -1 {
		return nil, oops.Errorf("invalid UDP datagram format: no newline found")
	}

	// Line 1: Source destination (base64) followed by optional FROM_PORT=nnn TO_PORT=nnn
	headerLine := strings.TrimSpace(response[:firstNewline])

	if headerLine == "" {
		return nil, oops.Errorf("empty header line in UDP datagram")
	}

	// Parse the header line to extract the destination
	// Format: "$destination FROM_PORT=nnn TO_PORT=nnn"
	// We need to split on space and take the first part as the destination
	parts := strings.Fields(headerLine)
	if len(parts) == 0 {
		return nil, oops.Errorf("empty header line in UDP datagram")
	}

	source := parts[0] // First field is the destination
	// Remaining parts are FROM_PORT and TO_PORT which we ignore for now

	// Everything after the first newline is the payload
	data := response[firstNewline+1:]

	if data == "" {
		return nil, oops.Errorf("no data in UDP datagram")
	}

	return s.createDatagram(source, data)
}

// createDatagram constructs the final Datagram from parsed source and data.
func (s *DatagramSession) createDatagram(source, data string) (*Datagram, error) {
	sourceAddr, err := i2pkeys.NewI2PAddrFromString(source)
	if err != nil {
		return nil, oops.Errorf("failed to parse source address: %w", err)
	}

	// Data is already raw bytes, not base64 encoded
	datagram := &Datagram{
		Data:   []byte(data),
		Source: sourceAddr,
		Local:  s.Addr(),
	}

	return datagram, nil
}

// Close closes the datagram session and all associated resources.
// This method safely terminates the session, closes the UDP listener and underlying connection,
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

	// Close the UDP listener for v3 forwarding
	if s.udpConn != nil {
		if err := s.udpConn.Close(); err != nil {
			logger.WithError(err).Warn("Failed to close UDP listener")
			// Continue with base session closure even if UDP close fails
		}
	}

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
