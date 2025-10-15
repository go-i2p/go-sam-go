package datagram2

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

// NewDatagram2Session creates a new datagram2 session with replay protection for UDP-like I2P messaging.
// It initializes the session with the provided SAM connection, session ID, cryptographic keys,
// and configuration options. The session automatically creates a UDP listener for receiving
// forwarded datagrams per SAMv3 requirements.
// Example usage: session, err := NewDatagram2Session(sam, "my-session", keys, []string{"inbound.length=1"})
func NewDatagram2Session(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*Datagram2Session, error) {
	log.WithFields(logrus.Fields{
		"id":      id,
		"style":   "DATAGRAM2",
		"options": options,
	}).Debug("Creating new Datagram2Session with SAMv3 UDP forwarding")

	// Create UDP listener and inject forwarding parameters
	udpConn, options, err := setupUDPForwardingListener(options)
	if err != nil {
		return nil, err
	}

	// Create the base session for datagram2
	baseSession, err := createGenericDatagram2Session(sam, id, keys, options)
	if err != nil {
		udpConn.Close()
		return nil, err
	}

	// Initialize the datagram2 session with UDP forwarding enabled
	ds := &Datagram2Session{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true, // Always true for SAMv3
	}

	log.Debug("Successfully created Datagram2Session with UDP forwarding and replay protection")
	return ds, nil
}

// setupUDPForwardingListener creates a UDP listener and injects forwarding parameters.
// This helper creates a UDP listener for SAMv3 datagram forwarding and automatically configures
// the session options to include the HOST and PORT parameters required by the SAM bridge.
// Returns the UDP connection, updated options slice, and any error encountered.
func setupUDPForwardingListener(options []string) (*net.UDPConn, []string, error) {
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
	log.WithField("udp_port", udpPort).Debug("Created UDP listener for datagram2 forwarding")

	// Inject UDP forwarding parameters into session options
	options = ensureUDPForwardingParameters(options, udpPort)

	return udpConn, options, nil
}

// createGenericDatagram2Session creates a validated BaseSession for datagram2.
// This helper creates the generic session through the SAM bridge using STYLE=DATAGRAM2,
// validates the session type, and ensures proper cleanup on error. It uses DATAGRAM2
// for replay protection.
func createGenericDatagram2Session(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*common.BaseSession, error) {
	// Create the base session using DATAGRAM2 style for replay protection
	session, err := sam.NewGenericSession("DATAGRAM2", id, keys, options)
	if err != nil {
		log.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create datagram2 session: %w", err)
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
// This is required for all datagram2 sessions in v3-only mode.
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

// NewDatagram2SessionFromSubsession creates a Datagram2Session for a subsession that has already been
// registered with a PRIMARY session using SESSION ADD. This constructor skips the session
// creation step since the subsession is already registered with the SAM bridge.
//
// For PRIMARY datagram2 subsessions, UDP forwarding is mandatory (SAMv3 requirement).
// The UDP connection must be provided for proper datagram reception.
//
// Example usage: session, err := NewDatagram2SessionFromSubsession(sam, "sub1", keys, options, udpConn)
func NewDatagram2SessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, udpConn *net.UDPConn) (*Datagram2Session, error) {
	logger := log.WithFields(logrus.Fields{
		"id":          id,
		"style":       "DATAGRAM2",
		"options":     options,
		"udp_enabled": udpConn != nil,
	})
	logger.Debug("Creating Datagram2Session from existing subsession with SAMv3 UDP forwarding")

	// Validate UDP connection is provided (mandatory for SAMv3 datagram2 subsessions)
	if udpConn == nil {
		logger.Error("UDP connection is required for SAMv3 datagram2 subsessions")
		return nil, oops.Errorf("udp connection is required for datagram2 subsessions (v3 only)")
	}

	// Create a BaseSession manually since the subsession is already registered via SESSION ADD
	// This bypasses SESSION CREATE and uses the existing registration
	baseSession, err := common.NewBaseSessionFromSubsession(sam, id, keys)
	if err != nil {
		logger.WithError(err).Error("Failed to create base session from subsession")
		return nil, oops.Errorf("failed to create datagram2 session from subsession: %w", err)
	}

	ds := &Datagram2Session{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true, // Always true for SAMv3
	}

	logger.Debug("Successfully created Datagram2Session from subsession with UDP forwarding and replay protection")
	return ds, nil
}

// NewReader creates a Datagram2Reader for receiving authenticated datagrams with replay protection.
// This method initializes a new reader with buffered channels for asynchronous datagram
// reception. The reader must be started manually with receiveLoop() for continuous operation.
// Example usage: reader := session.NewReader(); go reader.receiveLoop(); datagram, err := reader.ReceiveDatagram()
func (s *Datagram2Session) NewReader() *Datagram2Reader {
	// Create reader with buffered channels for non-blocking operation
	// The buffer size of 10 prevents blocking when multiple datagrams arrive rapidly
	return &Datagram2Reader{
		session:     s,
		recvChan:    make(chan *Datagram2, 10), // Buffer for incoming datagrams
		errorChan:   make(chan error, 1),
		closeChan:   make(chan struct{}),
		doneChan:    make(chan struct{}, 1),
		closed:      false,
		loopStarted: false,
		mu:          sync.RWMutex{},
		closeOnce:   sync.Once{},
	}
}

// NewWriter creates a Datagram2Writer for sending authenticated datagrams with replay protection.
// This method initializes a new writer with a default timeout of 30 seconds for send operations.
// The timeout can be customized using the SetTimeout method on the returned writer.
// Example usage: writer := session.NewWriter().SetTimeout(60*time.Second); err := writer.SendDatagram(data, dest)
func (s *Datagram2Session) NewWriter() *Datagram2Writer {
	// Initialize writer with default timeout for send operations
	// The timeout prevents indefinite blocking on send operations
	return &Datagram2Writer{
		session: s,
		timeout: 30, // Default timeout in seconds
	}
}

// PacketConn returns a net.PacketConn interface for this datagram2 session.
// This method provides compatibility with standard Go networking code by wrapping
// the datagram2 session in a connection that implements the PacketConn interface.
// The connection provides authenticated datagrams with replay protection.
//
// Example usage:
//
//	conn := session.PacketConn()
//	n, addr, err := conn.ReadFrom(buffer)
//	n, err = conn.WriteTo(data, destination)
func (s *Datagram2Session) PacketConn() net.PacketConn {
	// Create a PacketConn wrapper with integrated reader and writer
	// This provides standard Go networking interface compliance
	return &Datagram2Conn{
		session: s,
		reader:  s.NewReader(),
		writer:  s.NewWriter(),
	}
}

// SendDatagram sends an authenticated datagram with replay protection to the specified destination.
// This is a convenience method that creates a temporary writer and sends the datagram
// immediately. For multiple sends, it's more efficient to create a writer once and reuse it.
// Example usage: err := session.SendDatagram([]byte("hello"), destinationAddr)
func (s *Datagram2Session) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	// Use a temporary writer for one-time send operations
	// This simplifies the API for simple send operations
	return s.NewWriter().SendDatagram(data, dest)
}

// ReceiveDatagram receives a single authenticated datagram from the I2P network.
// This method is a convenience wrapper that performs a direct single read operation
// without starting a continuous receive loop. For continuous reception,
// use NewReader() and manage the reader lifecycle manually.
// Example usage: datagram, err := session.ReceiveDatagram()
func (s *Datagram2Session) ReceiveDatagram() (*Datagram2, error) {
	// ARCHITECTURAL FIX: Perform direct read for one-shot operations instead of
	// creating a reader with receive loop, which causes deadlocks due to RWMutex contention
	return s.readSingleDatagram()
}

// readSingleDatagram performs a direct read from the UDP connection for one-shot datagram operations.
// This method bypasses the reader infrastructure to avoid deadlocks when only one datagram is needed.
// SAMv3 UDP forwarding mode only - reads from UDP connection where SAM bridge forwards datagrams.
// V1/V2 TCP control socket reading is no longer supported.
func (s *Datagram2Session) readSingleDatagram() (*Datagram2, error) {
	s.mu.RLock()
	udpConn := s.udpConn
	s.mu.RUnlock()

	// V3-only: Always read from UDP connection
	if udpConn == nil {
		return nil, oops.Errorf("UDP connection not available (v3 UDP forwarding required)")
	}

	return s.readDatagramFromUDP(udpConn)
}

// readDatagramFromUDP reads a forwarded datagram2 message from the UDP connection.
// Format per SAMv3.md: destination line, port lines, empty line, payload.
func (s *Datagram2Session) readDatagramFromUDP(udpConn *net.UDPConn) (*Datagram2, error) {
	buffer := make([]byte, 65536) // Large buffer for UDP datagrams (I2P maximum)
	n, _, err := udpConn.ReadFromUDP(buffer)
	if err != nil {
		return nil, oops.Errorf("failed to read from UDP connection: %w", err)
	}

	log.WithFields(logrus.Fields{
		"bytes_read": n,
		"style":      "DATAGRAM2",
	}).Debug("Received UDP datagram2 message")

	// Parse the UDP datagram format per SAMv3.md
	response := string(buffer[:n])

	// Find the first newline - that's the end of the header line
	firstNewline := strings.Index(response, "\n")
	if firstNewline == -1 {
		return nil, oops.Errorf("invalid UDP datagram2 format: no newline found")
	}

	// Line 1: Source destination (base64, authenticated) followed by optional FROM_PORT=nnn TO_PORT=nnn
	headerLine := strings.TrimSpace(response[:firstNewline])

	if headerLine == "" {
		return nil, oops.Errorf("empty header line in UDP datagram2")
	}

	// Parse the header line to extract the authenticated source destination
	// Format: "$destination FROM_PORT=nnn TO_PORT=nnn"
	// We need to split on space and take the first part as the destination
	parts := strings.Fields(headerLine)
	if len(parts) == 0 {
		return nil, oops.Errorf("empty header line in UDP datagram2")
	}

	source := parts[0] // First field is the authenticated source destination
	// Remaining parts are FROM_PORT and TO_PORT which we ignore for now

	// Everything after the first newline is the payload
	data := response[firstNewline+1:]

	if data == "" {
		return nil, oops.Errorf("no data in UDP datagram2")
	}

	return s.createDatagram(source, data)
}

// createDatagram constructs the final Datagram2 from parsed authenticated source and data.
func (s *Datagram2Session) createDatagram(source, data string) (*Datagram2, error) {
	sourceAddr, err := i2pkeys.NewI2PAddrFromString(source)
	if err != nil {
		return nil, oops.Errorf("failed to parse authenticated source address: %w", err)
	}

	// Data is already raw bytes, not base64 encoded
	datagram := &Datagram2{
		Data:   []byte(data),
		Source: sourceAddr, // Authenticated by I2P router with replay protection
		Local:  s.Addr(),
	}

	return datagram, nil
}

// Close closes the datagram2 session and all associated resources.
// This method safely terminates the session, closes the UDP listener and underlying connection,
// and cleans up any background goroutines. It's safe to call multiple times.
// Example usage: defer session.Close()
func (s *Datagram2Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	// Log session closure for debugging and monitoring
	logger := log.WithFields(logrus.Fields{
		"id":    s.ID(),
		"style": "DATAGRAM2",
	})
	logger.Debug("Closing Datagram2Session")

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
		return oops.Errorf("failed to close datagram2 session: %w", err)
	}

	logger.Debug("Successfully closed Datagram2Session")
	return nil
}

// Addr returns the I2P address of this datagram2 session.
// This address represents the session's identity on the I2P network and can be
// used by other nodes to send authenticated datagrams with replay protection to this session.
// The address is derived from the session's cryptographic keys.
// Example usage: myAddr := session.Addr(); fmt.Println("My I2P address:", myAddr.Base32())
func (s *Datagram2Session) Addr() i2pkeys.I2PAddr {
	// Return the I2P address derived from the session's cryptographic keys
	return s.Keys().Addr()
}
