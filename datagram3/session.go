package datagram3

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

// NewDatagram3Session creates a new datagram3 session with hash-based source identification.
// It initializes the session with the provided SAM connection, session ID, cryptographic keys,
// and configuration options. The session automatically creates a UDP listener for receiving
// forwarded datagrams per SAMv3 requirements and initializes a hash resolver for source lookups.
// Note: DATAGRAM3 sources are not with full destinations; use datagram2 if authentication is required.
// Example usage: session, err := NewDatagram3Session(sam, "my-session", keys, []string{"inbound.length=1"})
func NewDatagram3Session(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*Datagram3Session, error) {
	// Log session creation with SECURITY WARNING
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"style":   "DATAGRAM3",
		"options": options,
	})
	
	logger.Debug("Creating new Datagram3Session with SAMv3 UDP forwarding")

	// Create UDP listener for receiving forwarded datagrams (SAMv3 requirement)
	// The SAM bridge will forward incoming DATAGRAM3 messages to this local UDP port
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		logger.WithError(err).Error("Failed to resolve UDP address")
		return nil, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.WithError(err).Error("Failed to create UDP listener")
		return nil, oops.Errorf("failed to create UDP listener: %w", err)
	}

	// Get the actual port assigned by the OS
	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	logger.WithField("udp_port", udpPort).Debug("Created UDP listener for datagram3 forwarding")

	// Inject UDP forwarding parameters into session options (SAMv3 requirement)
	// PORT/HOST tell the SAM bridge where to forward received datagrams
	options = ensureUDPForwardingParameters(options, udpPort)

	// CRITICAL: Use STYLE=DATAGRAM3 (not DATAGRAM or DATAGRAM2)
	// Create the base session using the common package for session management
	// This handles the underlying SAM protocol communication and session establishment
	session, err := sam.NewGenericSession("DATAGRAM3", id, keys, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session")
		udpConn.Close() // Clean up UDP listener on error
		return nil, oops.Errorf("failed to create datagram3 session: %w", err)
	}

	// Ensure the session is of the correct type for datagram3 operations
	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		udpConn.Close() // Clean up UDP listener on error
		return nil, oops.Errorf("invalid session type")
	}

	// Initialize the datagram3 session with UDP forwarding enabled and hash resolver
	ds := &Datagram3Session{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,                 // Always true for SAMv3
		resolver:    NewHashResolver(sam), // Initialize hash-to-destination cache
	}

	logger.Debug("Successfully created Datagram3Session with UDP forwarding and hash resolver")
	return ds, nil
}

// ensureUDPForwardingParameters injects UDP forwarding parameters into session options if not already present.
// This ensures SAMv3 UDP forwarding is configured with PORT and HOST parameters.
// PORT/HOST specify where the SAM bridge should forward datagrams TO (the client's UDP listener).
// sam.udp.port/sam.udp.host are NOT set here - they configure the SAM bridge's own UDP port (default 7655).
// This is required for all datagram3 sessions in v3-only mode.
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

// NewDatagram3SessionFromSubsession creates a Datagram3Session for a subsession that has already been
// registered with a PRIMARY session using SESSION ADD. This constructor skips the session
// creation step since the subsession is already registered with the SAM bridge.
//
// For PRIMARY datagram3 subsessions, UDP forwarding is mandatory (SAMv3 requirement).
// The UDP connection must be provided for proper datagram reception.
// Note: Sources are not with full destinations; use NewDatagramSubSession if authentication is required.
//
// Example usage: sub, err := NewDatagram3SessionFromSubsession(sam, "sub1", keys, options, udpConn)
func NewDatagram3SessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, udpConn *net.UDPConn) (*Datagram3Session, error) {
	logger := log.WithFields(logrus.Fields{
		"id":          id,
		"style":       "DATAGRAM3",
		"options":     options,
		"udp_enabled": udpConn != nil,
	})
	
	logger.Debug("Creating Datagram3Session from existing subsession with SAMv3 UDP forwarding")

	// Validate UDP connection is provided (mandatory for SAMv3 datagram3 subsessions)
	if udpConn == nil {
		logger.Error("UDP connection is required for SAMv3 datagram3 subsessions")
		return nil, oops.Errorf("udp connection is required for datagram3 subsessions (v3 only)")
	}

	// Create a BaseSession manually since the subsession is already registered via SESSION ADD
	// This bypasses SESSION CREATE and uses the existing registration
	baseSession, err := common.NewBaseSessionFromSubsession(sam, id, keys)
	if err != nil {
		logger.WithError(err).Error("Failed to create base session from subsession")
		return nil, oops.Errorf("failed to create datagram3 session from subsession: %w", err)
	}

	ds := &Datagram3Session{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,                 // Always true for SAMv3
		resolver:    NewHashResolver(sam), // Initialize hash-to-destination cache
	}

	logger.Debug("Successfully created Datagram3Session from subsession with UDP forwarding and hash resolver")
	return ds, nil
}

// NewReader creates a Datagram3Reader for receiving datagrams with hash-based sources.
// This method initializes a new reader with buffered channels for asynchronous datagram
// reception. The reader must be started manually with receiveLoop() for continuous operation.
// Received datagrams contain 32-byte hashes; call ResolveSource() to obtain full destinations for replies.
// Example usage: reader := session.NewReader(); go reader.receiveLoop(); datagram, err := reader.ReceiveDatagram()
func (s *Datagram3Session) NewReader() *Datagram3Reader {
	// Create reader with buffered channels for non-blocking operation
	// The buffer size of 10 prevents blocking when multiple datagrams arrive rapidly
	return &Datagram3Reader{
		session:     s,
		recvChan:    make(chan *Datagram3, 10), // Buffer for incoming datagrams
		errorChan:   make(chan error, 1),
		closeChan:   make(chan struct{}),
		doneChan:    make(chan struct{}, 1),
		closed:      false,
		loopStarted: false,
		mu:          sync.RWMutex{},
		closeOnce:   sync.Once{},
	}
}

// NewWriter creates a Datagram3Writer for sending datagrams to I2P destinations.
// This method initializes a new writer with a default timeout of 30 seconds for send operations.
// The timeout can be customized using the SetTimeout method on the returned writer.
// Example usage: writer := session.NewWriter().SetTimeout(60*time.Second); err := writer.SendDatagram(data, dest)
func (s *Datagram3Session) NewWriter() *Datagram3Writer {
	// Initialize writer with default timeout for send operations
	// The timeout prevents indefinite blocking on send operations
	return &Datagram3Writer{
		session: s,
		timeout: 30, // Default timeout in seconds
	}
}

// Close terminates the datagram3 session and cleans up all resources.
// This method ensures proper cleanup of the UDP connection and I2P tunnels.
// After calling Close(), the session cannot be reused.
// Example usage: defer session.Close()
func (s *Datagram3Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	log.WithField("session", s.ID()).Debug("Closing Datagram3Session")

	// Close UDP connection if present
	if s.udpConn != nil {
		if err := s.udpConn.Close(); err != nil {
			log.WithError(err).Warn("Error closing UDP connection")
		}
	}

	// Clear hash resolver cache to free memory
	if s.resolver != nil {
		s.resolver.Clear()
	}

	// Close base session (closes I2P tunnels)
	if err := s.BaseSession.Close(); err != nil {
		return oops.Errorf("failed to close base session: %w", err)
	}

	s.closed = true
	log.WithField("session", s.ID()).Debug("Datagram3Session closed successfully")
	return nil
}

// Addr returns the local I2P address of this datagram3 session.
// This is the destination address that other I2P nodes can use to send datagrams to this session.
func (s *Datagram3Session) Addr() i2pkeys.I2PAddr {
	return s.Keys().Addr()
}
