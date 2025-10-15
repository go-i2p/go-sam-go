package raw

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

// ensureRawUDPForwardingParameters injects UDP forwarding parameters into session options if not already present.
// This ensures SAMv3 UDP forwarding is configured with PORT, HOST, sam.udp.port, and sam.udp.host parameters.
// This is required for all raw sessions in v3-only mode.
func ensureRawUDPForwardingParameters(options []string, udpPort int) []string {
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

	// Add PORT/HOST to tell SAM bridge where to forward datagrams TO (our UDP listener)
	// Do NOT set sam.udp.port/sam.udp.host - those configure SAM bridge's own UDP port (default 7655)
	if !hasHost {
		updatedOptions = append(updatedOptions, "HOST=127.0.0.1")
	}
	if !hasPort {
		updatedOptions = append(updatedOptions, "PORT="+strconv.Itoa(udpPort))
	}

	return updatedOptions
}

// NewRawSession creates a new raw session for sending and receiving raw datagrams using SAMv3 UDP forwarding.
// It initializes the session with the provided SAM connection, session ID, cryptographic keys,
// and configuration options. It automatically creates a UDP listener for receiving forwarded datagrams
// (SAMv3 requirement) and configures the session with PORT/HOST parameters.
// V1/V2 compatibility (reading from TCP control socket) is no longer supported.
// Returns a RawSession instance that uses UDP forwarding for all raw datagram reception.
// Example usage: session, err := NewRawSession(sam, "my-session", keys, []string{"inbound.length=1"})
func NewRawSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error) {
	log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	}).Debug("Creating new RawSession with SAMv3 UDP forwarding")

	// Create UDP listener and inject forwarding parameters
	udpConn, options, err := setupRawUDPForwarding(options)
	if err != nil {
		return nil, err
	}

	// Create the base session for raw datagrams
	baseSession, err := createGenericRawSession(sam, id, keys, options)
	if err != nil {
		udpConn.Close()
		return nil, err
	}

	rs := &RawSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,
	}

	log.Debug("Successfully created RawSession with UDP forwarding")
	return rs, nil
}

// setupRawUDPForwarding creates a UDP listener and injects forwarding parameters.
// This helper creates a UDP listener for SAMv3 raw datagram forwarding and automatically
// configures the session options to include the HOST and PORT parameters required by the SAM bridge.
// Returns the UDP connection, updated options slice, and any error encountered.
func setupRawUDPForwarding(options []string) (*net.UDPConn, []string, error) {
	// Create UDP listener for receiving forwarded raw datagrams
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
	log.WithField("udp_port", udpPort).Debug("Created UDP listener for raw datagram forwarding")

	// Inject UDP forwarding parameters into session options
	options = ensureRawUDPForwardingParameters(options, udpPort)

	return udpConn, options, nil
}

// createGenericRawSession creates a validated BaseSession for raw datagrams.
// This helper creates the generic session through the SAM bridge using STYLE=RAW,
// validates the session type, and ensures proper cleanup on error.
func createGenericRawSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*common.BaseSession, error) {
	// Create the base session using RAW style
	session, err := sam.NewGenericSession("RAW", id, keys, options)
	if err != nil {
		log.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create raw session: %w", err)
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

// NewRawSessionFromSubsession creates a RawSession for a subsession that has already been
// registered with a PRIMARY session using SESSION ADD. This constructor skips the session
// creation step since the subsession is already registered with the SAM bridge.
//
// This function is specifically designed for use with SAMv3.3 PRIMARY sessions where
// subsessions are created using SESSION ADD rather than SESSION CREATE commands.
//
// For PRIMARY raw subsessions, UDP forwarding is mandatory (SAMv3 requirement).
// The UDP connection must be provided for proper raw datagram reception via UDP forwarding.
//
// Parameters:
//   - sam: SAM connection for data operations (separate from the primary session's control connection)
//   - id: The subsession ID that was already registered with SESSION ADD
//   - keys: The I2P keys from the primary session (shared across all subsessions)
//   - options: Configuration options for the subsession
//   - udpConn: UDP connection for receiving forwarded raw datagrams (required, not nil)
//
// Returns a RawSession ready for use without attempting to create a new SAM session.
func NewRawSessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, udpConn *net.UDPConn) (*RawSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":          id,
		"options":     options,
		"udp_enabled": udpConn != nil,
	})
	logger.Debug("Creating RawSession from existing subsession with SAMv3 UDP forwarding")

	if udpConn == nil {
		logger.Error("UDP connection is required for SAMv3 raw subsessions")
		return nil, oops.Errorf("udp connection is required for raw subsessions (v3 only)")
	}

	// Create a BaseSession manually since the session is already registered
	baseSession, err := common.NewBaseSessionFromSubsession(sam, id, keys)
	if err != nil {
		logger.WithError(err).Error("Failed to create base session from subsession")
		return nil, oops.Errorf("failed to create raw session from subsession: %w", err)
	}

	rs := &RawSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,
	}

	logger.Debug("Successfully created RawSession from subsession with UDP forwarding")
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

// ReceiveDatagram receives a single raw datagram from any source using SAMv3 UDP forwarding.
// This method performs a direct UDP read without creating a reader or receive loop.
// V1/V2 TCP control socket reading is no longer supported.
// Example usage: datagram, err := session.ReceiveDatagram()
func (s *RawSession) ReceiveDatagram() (*RawDatagram, error) {
	s.mu.RLock()
	udpConn := s.udpConn
	s.mu.RUnlock()

	// V3-only: Always read from UDP connection
	if udpConn == nil {
		return nil, oops.Errorf("UDP connection not available (v3 UDP forwarding required)")
	}

	return s.readRawFromUDP(udpConn)
}

// readRawFromUDP reads a forwarded raw datagram from the UDP connection.
// This is used for SAMv3 UDP forwarding where raw datagrams are forwarded as-is.
// Format per SAMv3.md: Raw datagrams are forwarded as-is to the specified host:port without a prefix.
// For RAW sessions with HEADER=true option, datagrams are prepended with a line containing
// PROTOCOL=nnn FROM_PORT=nnnn TO_PORT=nnnn.
func (s *RawSession) readRawFromUDP(udpConn *net.UDPConn) (*RawDatagram, error) {
	buffer := make([]byte, 65536) // Large buffer for UDP datagrams
	n, remoteAddr, err := udpConn.ReadFromUDP(buffer)
	if err != nil {
		return nil, oops.Errorf("failed to read from UDP connection: %w", err)
	}

	log.WithFields(logrus.Fields{
		"bytes_read":  n,
		"remote_addr": remoteAddr,
	}).Debug("Received UDP raw datagram")

	// Raw datagrams are forwarded as-is (no prefix in standard mode)
	// TODO: Handle HEADER=true mode if needed in the future
	data := buffer[:n]

	// For raw datagrams without header, we don't have source destination information
	// This is by design for RAW datagrams - they are anonymous
	// Create empty I2P address for anonymous source
	emptyKeys := i2pkeys.I2PKeys{}
	datagram := &RawDatagram{
		Data:   data,
		Source: emptyKeys.Addr(), // Empty source for anonymous raw datagrams
		Local:  s.Addr(),
	}

	return datagram, nil
}

// Close closes the raw session and all associated resources.
// This method safely terminates the session, closes the UDP listener and underlying connection,
// and cleans up any background goroutines. It's safe to call multiple times.
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

	// Close the UDP listener for v3 forwarding
	if s.udpConn != nil {
		if err := s.udpConn.Close(); err != nil {
			logger.WithError(err).Warn("Failed to close UDP listener")
			// Continue with base session closure even if UDP close fails
		}
	}

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
