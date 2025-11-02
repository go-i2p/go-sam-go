package primary

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/go-i2p/go-sam-go/datagram3"
	"github.com/go-i2p/go-sam-go/raw"
	"github.com/go-i2p/go-sam-go/stream"
	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/logger"
	"github.com/samber/oops"
)

// PrimarySession manages multiple sub-sessions sharing the same I2P identity and tunnels.
type PrimarySession struct {
	*common.BaseSession
	sam      *common.SAM
	options  []string
	registry *SubSessionRegistry
	mu       sync.RWMutex
	closed   bool
	// usedPorts tracks allocated ports for stream subsessions to prevent duplicates
	usedPorts map[int]bool
	// nextAutoPort tracks the next available port for auto-assignment
	nextAutoPort int
	// subSessionPorts tracks which auto-assigned port belongs to which subsession
	subSessionPorts map[string]int
}

// NewPrimarySession creates a new primary session for managing multiple sub-sessions.
// It initializes the session with the provided SAM connection, session ID, cryptographic keys,
// and configuration options. The primary session allows creating multiple sub-sessions of
// different types (stream, datagram, raw) while sharing the same I2P identity and tunnels.
// Example usage: session, err := NewPrimarySession(sam, "my-primary", keys, []string{"inbound.length=2"})
func NewPrimarySession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	logger := log.WithFields(logger.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating new PrimarySession")

	// Create the base session using the common package with PRIMARY style
	// The PRIMARY session type allows multiple sub-sessions with shared identity
	session, err := sam.NewGenericSession("PRIMARY", id, keys, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic primary session")
		return nil, oops.Errorf("failed to create primary session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ps := &PrimarySession{
		BaseSession:     baseSession,
		sam:             sam,
		options:         options,
		registry:        NewSubSessionRegistry(),
		usedPorts:       make(map[int]bool),
		nextAutoPort:    49152, // Start from dynamic port range
		subSessionPorts: make(map[string]int),
	}

	logger.Debug("Successfully created PrimarySession")
	return ps, nil
}

// NewPrimarySessionWithSignature creates a new primary session with the specified signature type.
// This method allows specifying custom cryptographic parameters for enhanced security or
// compatibility with specific I2P network configurations.
// Example usage: session, err := NewPrimarySessionWithSignature(sam, "secure-primary", keys, options, "EdDSA_SHA512_Ed25519")
func NewPrimarySessionWithSignature(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error) {
	logger := log.WithFields(logger.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new PrimarySession with signature")

	// Create the base session using the common package with custom signature
	// This enables advanced cryptographic configuration for enhanced security
	session, err := sam.NewGenericSessionWithSignature("PRIMARY", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic primary session with signature")
		return nil, oops.Errorf("failed to create primary session: %w", err)
	}

	// Ensure the session is of the correct type for primary session operations
	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	// Initialize the primary session with the base session and configuration
	ps := &PrimarySession{
		BaseSession:     baseSession,
		sam:             sam,
		options:         options,
		registry:        NewSubSessionRegistry(),
		usedPorts:       make(map[int]bool),
		nextAutoPort:    49152, // Start from dynamic port range
		subSessionPorts: make(map[string]int),
	}

	logger.Debug("Successfully created PrimarySession with signature")
	return ps, nil
}

// NewStreamSubSession creates a new stream sub-session within this primary session.
// The sub-session shares the primary session's I2P identity and tunnel infrastructure
// while providing full StreamSession functionality for TCP-like reliable connections.
// Each sub-session must have a unique identifier within the primary session scope.
//
// If no port options (LISTEN_PORT, FROM_PORT) are provided, the method will automatically
// assign a unique port to prevent "Duplicate protocol" errors from the SAM bridge.
// This ensures multiple stream sub-sessions can coexist within the same primary session.
//
// Example usage: streamSub, err := primary.NewStreamSubSession("tcp-handler", []string{"FROM_PORT=8080"})
func (p *PrimarySession) NewStreamSubSession(id string, options []string) (*StreamSubSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	log.WithFields(logger.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
	}).Debug("Creating stream sub-session")

	// Ensure unique port assignment to prevent duplicate protocol errors
	finalOptions, assignedPort, err := p.ensureUniqueStreamPort(options)
	if err != nil {
		return nil, err
	}

	// Add and setup the stream subsession
	subSAM, err := p.addAndSetupStreamSubsession(id, finalOptions)
	if err != nil {
		// If we auto-assigned a port, release it on failure
		if assignedPort > 0 {
			p.releasePort(assignedPort)
		}
		return nil, err
	}

	// Create the stream session from the subsession
	streamSession, err := p.createStreamSessionFromSubsession(subSAM, id, finalOptions)
	if err != nil {
		subSAM.Close()
		p.sam.RemoveSubSession(id)
		// If we auto-assigned a port, release it on failure
		if assignedPort > 0 {
			p.releasePort(assignedPort)
		}
		return nil, err
	}

	// Register and return the stream sub-session
	subSession, err := p.registerStreamSubSession(id, streamSession)
	if err != nil {
		streamSession.Close()
		p.sam.RemoveSubSession(id)
		// If we auto-assigned a port, release it on failure
		if assignedPort > 0 {
			p.releasePort(assignedPort)
		}
		return nil, err
	}

	// Track the port assignment for cleanup when subsession is closed
	if assignedPort > 0 {
		p.subSessionPorts[id] = assignedPort
	}

	log.WithField("assigned_port", assignedPort).Debug("Successfully created stream sub-session")
	return subSession, nil
}

// NewStreamSubSessionWithPort creates a new stream sub-session with explicit port configuration.
// This method allows precise control over port assignments for advanced use cases where
// specific port numbers are required for the stream subsession.
//
// Parameters:
//   - id: Unique subsession identifier within the primary session scope
//   - options: Additional SAM protocol options (port options will be added/overridden)
//   - fromPort: The FROM_PORT parameter for outbound connections (0 = any port)
//   - toPort: The TO_PORT parameter for inbound connections (0 = any port)
//
// The method validates that the specified ports are available and reserves them
// to prevent conflicts with other subsessions. Port 0 means "any available port".
//
// Example usage:
//
//	streamSub, err := primary.NewStreamSubSessionWithPort("tcp-server", []string{}, 8080, 8080)
//	streamSub2, err := primary.NewStreamSubSessionWithPort("tcp-client", []string{}, 0, 8081)
func (p *PrimarySession) NewStreamSubSessionWithPort(id string, options []string, fromPort, toPort int) (*StreamSubSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	log.WithFields(logger.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
		"from_port":  fromPort,
		"to_port":    toPort,
	}).Debug("Creating stream sub-session with explicit ports")

	// Build options with explicit port configuration
	finalOptions, reservedPorts, err := p.buildOptionsWithPorts(options, fromPort, toPort)
	if err != nil {
		return nil, err
	}

	// Add and setup the stream subsession
	subSAM, err := p.addAndSetupStreamSubsession(id, finalOptions)
	if err != nil {
		// Release any reserved ports on failure
		for _, port := range reservedPorts {
			if port > 0 {
				p.releasePort(port)
			}
		}
		return nil, err
	}

	// Create the stream session from the subsession
	streamSession, err := p.createStreamSessionFromSubsession(subSAM, id, finalOptions)
	if err != nil {
		subSAM.Close()
		p.sam.RemoveSubSession(id)
		// Release any reserved ports on failure
		for _, port := range reservedPorts {
			if port > 0 {
				p.releasePort(port)
			}
		}
		return nil, err
	}

	// Register and return the stream sub-session
	subSession, err := p.registerStreamSubSession(id, streamSession)
	if err != nil {
		streamSession.Close()
		p.sam.RemoveSubSession(id)
		// Release any reserved ports on failure
		for _, port := range reservedPorts {
			if port > 0 {
				p.releasePort(port)
			}
		}
		return nil, err
	}

	// Track the primary port assignment for cleanup (use fromPort as primary)
	primaryPort := fromPort
	if primaryPort == 0 && toPort > 0 {
		primaryPort = toPort
	}
	if primaryPort > 0 {
		p.subSessionPorts[id] = primaryPort
	}

	log.WithFields(logger.Fields{
		"from_port": fromPort,
		"to_port":   toPort,
	}).Debug("Successfully created stream sub-session with explicit ports")
	return subSession, nil
}

// addAndSetupStreamSubsession registers the subsession with SAM and creates its connection.
// This helper registers a stream subsession using SESSION ADD and creates a dedicated
// SAM connection for stream data operations (CONNECT/ACCEPT).
func (p *PrimarySession) addAndSetupStreamSubsession(id string, options []string) (*common.SAM, error) {
	// Add the subsession to the primary session using SESSION ADD
	if err := p.sam.AddSubSession("STREAM", id, options); err != nil {
		log.WithError(err).Error("Failed to add stream subsession")
		return nil, oops.Errorf("failed to create stream sub-session: %w", err)
	}

	// Create a new SAM connection for the sub-session data operations
	subSAM, err := p.createSubSAMConnection()
	if err != nil {
		log.WithError(err).Error("Failed to create sub-SAM connection")
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	return subSAM, nil
}

// createStreamSessionFromSubsession wraps the registered subsession in a StreamSession.
// This helper creates a stream session wrapper for a subsession that has already been
// registered via SESSION ADD, avoiding duplicate session creation.
func (p *PrimarySession) createStreamSessionFromSubsession(subSAM *common.SAM, id string, options []string) (*stream.StreamSession, error) {
	streamSession, err := stream.NewStreamSessionFromSubsession(subSAM, id, p.Keys(), options)
	if err != nil {
		log.WithError(err).Error("Failed to create stream session wrapper")
		return nil, oops.Errorf("failed to create stream sub-session: %w", err)
	}
	return streamSession, nil
}

// registerStreamSubSession wraps and registers a stream session as a subsession.
// This helper creates the StreamSubSession adapter and registers it with the primary
// session's subsession registry for lifecycle management.
func (p *PrimarySession) registerStreamSubSession(id string, streamSession *stream.StreamSession) (*StreamSubSession, error) {
	// Wrap the stream session in a sub-session adapter
	subSession := NewStreamSubSession(id, streamSession)

	// Register the sub-session with the primary session registry
	if err := p.registry.Register(id, subSession); err != nil {
		log.WithError(err).Error("Failed to register stream sub-session")
		return nil, oops.Errorf("failed to register stream sub-session: %w", err)
	}

	return subSession, nil
}

// NewUniqueStreamSubSession creates a new unique stream sub-session within this primary session.
func (p *PrimarySession) NewUniqueStreamSubSession(s string) (*StreamSubSession, error) {
	// random number between 1000 and 9999
	randomId := s + "-" + strconv.FormatInt(rand.Int63n(8999)+1000, 10)
	return p.NewStreamSubSession(randomId, nil)
}

// setupUDPListenerForDatagram creates and binds a UDP listener for DATAGRAM forwarding.
// Returns the UDP connection and assigned port number. This helper isolates the network
// setup logic from the main session creation flow, improving testability and code clarity.
func setupUDPListenerForDatagram() (*net.UDPConn, int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0") // Port 0 = let OS choose
	if err != nil {
		return nil, 0, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, 0, oops.Errorf("failed to create UDP listener: %w", err)
	}

	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	return udpConn, udpPort, nil
}

// registerDatagramSubsession registers a DATAGRAM subsession with the SAM bridge and creates
// a new sub-SAM connection for data operations. Returns the sub-SAM connection or error.
// Closes the UDP connection and performs cleanup on failure.
func (p *PrimarySession) registerDatagramSubsession(id string, options []string, udpConn *net.UDPConn, logger *logger.Entry) (*common.SAM, error) {
	if err := p.sam.AddSubSession("DATAGRAM", id, options); err != nil {
		logger.WithError(err).Error("Failed to add datagram subsession")
		udpConn.Close()
		return nil, oops.Errorf("failed to create datagram sub-session: %w", err)
	}

	subSAM, err := p.createSubSAMConnection()
	if err != nil {
		logger.WithError(err).Error("Failed to create sub-SAM connection")
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	return subSAM, nil
}

// createAndRegisterDatagramWrapper creates a datagram session wrapper from the subsession components
// and registers it with the primary session registry. Returns the configured sub-session or error.
// Performs complete cleanup of all resources on failure.
func (p *PrimarySession) createAndRegisterDatagramWrapper(id string, options []string, subSAM *common.SAM, udpConn *net.UDPConn, logger *logger.Entry) (*DatagramSubSession, error) {
	datagramSession, err := datagram.NewDatagramSessionFromSubsession(subSAM, id, p.Keys(), options, udpConn)
	if err != nil {
		logger.WithError(err).Error("Failed to create datagram session wrapper")
		subSAM.Close()
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create datagram sub-session: %w", err)
	}

	subSession := NewDatagramSubSession(id, datagramSession)

	if err := p.registry.Register(id, subSession); err != nil {
		logger.WithError(err).Error("Failed to register datagram sub-session")
		datagramSession.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to register datagram sub-session: %w", err)
	}

	return subSession, nil
}

// NewDatagramSubSession creates a new datagram sub-session within this primary session.
// The sub-session shares the primary session's I2P identity and tunnel infrastructure
// while providing full DatagramSession functionality for UDP-like authenticated messaging.
// Each sub-session must have a unique identifier within the primary session scope.
//
// This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
// the subsession with the primary session's SAM connection, ensuring compliance
// with the I2P SAM protocol specification for PRIMARY session management.
//
// Per SAMv3.3 specification, DATAGRAM subsessions REQUIRE a PORT parameter.
// If PORT is not included in the options, PORT=0 (any port) will be added automatically.
//
// Example usage:
//
//	datagramSub, err := primary.NewDatagramSubSession("udp-handler", []string{"PORT=8080", "FROM_PORT=8080"})
//	writer := datagramSub.NewWriter()
//	reader := datagramSub.NewReader()
func (p *PrimarySession) NewDatagramSubSession(id string, options []string) (*DatagramSubSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	logger := log.WithFields(logger.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
	})
	logger.Debug("Creating datagram sub-session with UDP forwarding")

	udpConn, udpPort, err := setupUDPListenerForDatagram()
	if err != nil {
		return nil, err
	}

	logger.WithField("udp_port", udpPort).Debug("Created UDP listener for datagram forwarding")

	finalOptions := ensureDatagramForwardingParameters(options, udpPort)

	subSAM, err := p.registerDatagramSubsession(id, finalOptions, udpConn, logger)
	if err != nil {
		return nil, err
	}

	subSession, err := p.createAndRegisterDatagramWrapper(id, options, subSAM, udpConn, logger)
	if err != nil {
		return nil, err
	}

	logger.WithField("udp_port", udpPort).Debug("Successfully created datagram sub-session with UDP forwarding")
	return subSession, nil
}

// NewRawSubSession creates a new raw sub-session within this primary session using SAMv3 UDP forwarding.
// The sub-session shares the primary session's I2P identity and tunnel infrastructure
// while providing full RawSession functionality for unrepliable datagram communication.
// Each sub-session must have a unique identifier within the primary session scope.
//
// This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
// the subsession with the primary session's SAM connection, ensuring compliance
// with the I2P SAM protocol specification for PRIMARY session management.
//
// Per SAMv3.3 specification, RAW subsessions REQUIRE UDP forwarding for proper operation.
// V1/V2 TCP control socket reading is no longer supported.
//
// Example usage:
//
//	rawSub, err := primary.NewRawSubSession("raw-sender", []string{"FROM_PORT=8080"})
//	writer := rawSub.NewWriter()
//	reader := rawSub.NewReader()
func (p *PrimarySession) NewRawSubSession(id string, options []string) (*RawSubSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	log.WithFields(logger.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
	}).Debug("Creating raw sub-session with UDP forwarding")

	// Setup UDP listener and add subsession
	udpConn, udpPort, finalOptions, err := p.setupRawUDPForwarding(options)
	if err != nil {
		return nil, err
	}

	// Add subsession with automatic cleanup on error
	if err = p.addRawSubSession(id, finalOptions); err != nil {
		udpConn.Close()
		return nil, err
	}

	// Create and register the raw session wrapper
	subSession, err := p.createAndRegisterRawSession(id, options, udpConn, udpPort)
	if err != nil {
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, err
	}

	return subSession, nil
}

// createAndRegisterRawSession creates the raw session wrapper and registers it.
// This helper consolidates SAM connection creation, session wrapping, and registration
// with unified error handling and cleanup logic.
func (p *PrimarySession) createAndRegisterRawSession(id string, options []string, udpConn *net.UDPConn, udpPort int) (*RawSubSession, error) {
	// Create SAM connection for the subsession
	subSAM, err := p.createSubSAMConnection()
	if err != nil {
		log.WithError(err).Error("Failed to create sub-SAM connection")
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	// Create raw session wrapper
	rawSession, err := raw.NewRawSessionFromSubsession(subSAM, id, p.Keys(), options, udpConn)
	if err != nil {
		log.WithError(err).Error("Failed to create raw session wrapper")
		subSAM.Close()
		return nil, oops.Errorf("failed to create raw sub-session: %w", err)
	}

	// Register the subsession
	subSession, err := p.registerRawSubSession(id, rawSession, udpPort)
	if err != nil {
		rawSession.Close()
		subSAM.Close()
		return nil, err
	}

	return subSession, nil
}

// setupRawUDPForwarding creates and configures a UDP listener for raw datagram forwarding.
// Returns the UDP connection, assigned port number, and finalized options with forwarding parameters.
// This helper isolates network setup and option configuration for improved testability.
func (p *PrimarySession) setupRawUDPForwarding(options []string) (*net.UDPConn, int, []string, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0") // Port 0 = let OS choose
	if err != nil {
		return nil, 0, nil, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, 0, nil, oops.Errorf("failed to create UDP listener: %w", err)
	}

	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	log.WithField("udp_port", udpPort).Debug("Created UDP listener for raw datagram forwarding")

	finalOptions := ensureRawForwardingParameters(options, udpPort)
	return udpConn, udpPort, finalOptions, nil
}

// addRawSubSession adds a raw subsession to the primary session using SESSION ADD command.
// Handles SAM protocol communication and error logging for subsession registration.
func (p *PrimarySession) addRawSubSession(id string, finalOptions []string) error {
	if err := p.sam.AddSubSession("RAW", id, finalOptions); err != nil {
		log.WithError(err).Error("Failed to add raw subsession")
		return oops.Errorf("failed to create raw sub-session: %w", err)
	}
	return nil
}

// registerRawSubSession wraps and registers a raw session with the primary session registry.
// Returns the completed sub-session adapter ready for use.
func (p *PrimarySession) registerRawSubSession(id string, rawSession *raw.RawSession, udpPort int) (*RawSubSession, error) {
	subSession := NewRawSubSession(id, rawSession)

	if err := p.registry.Register(id, subSession); err != nil {
		log.WithError(err).Error("Failed to register raw sub-session")
		rawSession.Close()
		return nil, oops.Errorf("failed to register raw sub-session: %w", err)
	}

	log.WithField("udp_port", udpPort).Debug("Successfully created raw sub-session with UDP forwarding")
	return subSession, nil
}

// setupUDPListenerForDatagram3 creates and binds a UDP listener for DATAGRAM3 forwarding.
// Returns the UDP connection and assigned port number. This helper isolates the network
// setup logic from the main session creation flow, improving testability and code clarity.
func setupUDPListenerForDatagram3() (*net.UDPConn, int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0") // Port 0 = let OS choose
	if err != nil {
		return nil, 0, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, 0, oops.Errorf("failed to create UDP listener: %w", err)
	}

	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	return udpConn, udpPort, nil
}

// registerDatagram3Subsession registers a DATAGRAM3 subsession with the SAM bridge and creates
// a new sub-SAM connection for data operations. Returns the sub-SAM connection or error.
// Closes the UDP connection and performs cleanup on failure.
func (p *PrimarySession) registerDatagram3Subsession(id string, options []string, udpConn *net.UDPConn, logger *logger.Entry) (*common.SAM, error) {
	if err := p.sam.AddSubSession("DATAGRAM3", id, options); err != nil {
		logger.WithError(err).Error("Failed to add datagram3 subsession")
		udpConn.Close()
		return nil, oops.Errorf("failed to create datagram3 sub-session: %w", err)
	}

	subSAM, err := p.createSubSAMConnection()
	if err != nil {
		logger.WithError(err).Error("Failed to create sub-SAM connection")
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	return subSAM, nil
}

// createAndRegisterDatagram3Wrapper creates a datagram3 session wrapper from the subsession components
// and registers it with the primary session registry. Returns the configured sub-session or error.
// Performs complete cleanup of all resources on failure.
func (p *PrimarySession) createAndRegisterDatagram3Wrapper(id string, options []string, subSAM *common.SAM, udpConn *net.UDPConn, logger *logger.Entry) (*Datagram3SubSession, error) {
	datagram3Session, err := datagram3.NewDatagram3SessionFromSubsession(subSAM, id, p.Keys(), options, udpConn)
	if err != nil {
		logger.WithError(err).Error("Failed to create datagram3 session wrapper")
		subSAM.Close()
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create datagram3 sub-session: %w", err)
	}

	subSession := NewDatagram3SubSession(id, datagram3Session)

	if err := p.registry.Register(id, subSession); err != nil {
		logger.WithError(err).Error("Failed to register datagram3 sub-session")
		datagram3Session.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to register datagram3 sub-session: %w", err)
	}

	return subSession, nil
}

// NewDatagram3SubSession creates a new datagram3 sub-session within this primary session using SAMv3 UDP forwarding.
// The sub-session shares the primary session's I2P identity and tunnel infrastructure
// while providing full Datagram3Session functionality for repliable but UNAUTHENTICATED datagram communication.
// Each sub-session must have a unique identifier within the primary session scope.
//
// This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
// the subsession with the primary session's SAM connection, ensuring compliance
// with the I2P SAM protocol specification for PRIMARY session management.
//
// Per SAMv3.3 specification, DATAGRAM3 subsessions REQUIRE UDP forwarding for proper operation.
// Received datagrams contain a 32-byte hash instead of full authenticated destination.
// Use the session's hash resolver to convert hashes to destinations for replies.
//
// Example usage:
//
//	datagram3Sub, err := primary.NewDatagram3SubSession("udp3-handler", []string{"FROM_PORT=8080"})
//	reader := datagram3Sub.NewReader()
//	writer := datagram3Sub.NewWriter()
//	// Receive datagram with UNAUTHENTICATED source hash
//	dg, err := reader.ReceiveDatagram()
//	// Resolve hash to reply (cached by session)
//	err = dg.ResolveSource(datagram3Sub)
//	err = writer.SendDatagram([]byte("reply"), dg.Source)
func (p *PrimarySession) NewDatagram3SubSession(id string, options []string) (*Datagram3SubSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	logger := log.WithFields(logger.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
	})
	logger.Warn("Creating DATAGRAM3 sub-session - sources are UNAUTHENTICATED and can be spoofed!")

	udpConn, udpPort, err := setupUDPListenerForDatagram3()
	if err != nil {
		return nil, err
	}

	logger.WithField("udp_port", udpPort).Debug("Created UDP listener for datagram3 forwarding")

	finalOptions := ensureDatagram3ForwardingParameters(options, udpPort)

	subSAM, err := p.registerDatagram3Subsession(id, finalOptions, udpConn, logger)
	if err != nil {
		return nil, err
	}

	subSession, err := p.createAndRegisterDatagram3Wrapper(id, options, subSAM, udpConn, logger)
	if err != nil {
		return nil, err
	}

	logger.WithField("udp_port", udpPort).Warn("Successfully created datagram3 sub-session - remember sources are UNAUTHENTICATED!")
	return subSession, nil
}

// GetSubSession retrieves a sub-session by its unique identifier.
// Returns the sub-session instance if found, or an error if the sub-session
// does not exist or the primary session is closed. This method provides
// safe access to registered sub-sessions for management and operation.
//
// Example usage:
//
//	subSession, err := primary.GetSubSession("stream-1")
//	if streamSub, ok := subSession.(*StreamSubSession); ok {
//	    conn, err := streamSub.Dial("destination.b32.i2p")
//	}
func (p *PrimarySession) GetSubSession(id string) (SubSession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	subSession, exists := p.registry.Get(id)
	if !exists {
		return nil, oops.Errorf("sub-session with ID '%s' not found", id)
	}

	return subSession, nil
}

// ListSubSessions returns a list of all currently active sub-sessions.
// This method provides a snapshot of all registered sub-sessions that can be
// safely iterated without holding locks. The returned list includes sub-sessions
// of all types (stream, datagram, raw) currently managed by this primary session.
//
// Example usage:
//
//	subSessions := primary.ListSubSessions()
//	for _, sub := range subSessions {
//	    log.Printf("Sub-session %s (type: %s) is active: %v", sub.ID(), sub.Type(), sub.Active())
//	}
func (p *PrimarySession) ListSubSessions() []SubSession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil
	}

	return p.registry.List()
}

// CloseSubSession closes and unregisters a specific sub-session by its ID.
// This method provides selective termination of sub-sessions without affecting
// the primary session or other sub-sessions. The sub-session is properly cleaned
// up and removed from the registry after closure.
//
// Example usage:
//
//	err := primary.CloseSubSession("stream-1")
//	if err != nil {
//	    log.Printf("Failed to close sub-session: %v", err)
//	}
func (p *PrimarySession) CloseSubSession(id string) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return oops.Errorf("primary session is closed")
	}
	p.mu.RUnlock()

	logger := log.WithFields(logger.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
	})
	logger.Debug("Closing sub-session")

	// Get the sub-session from registry
	subSession, exists := p.registry.Get(id)
	if !exists {
		return oops.Errorf("sub-session with ID '%s' not found", id)
	}

	// Close the sub-session
	if err := subSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close sub-session")
		return oops.Errorf("failed to close sub-session: %w", err)
	}

	// Unregister from the registry
	if err := p.registry.Unregister(id); err != nil {
		logger.WithError(err).Error("Failed to unregister sub-session")
		return oops.Errorf("failed to unregister sub-session: %w", err)
	}

	// Release auto-assigned port if this was a stream subsession with auto-port
	p.mu.Lock()
	if port, hasAutoPort := p.subSessionPorts[id]; hasAutoPort {
		p.releasePort(port)
		delete(p.subSessionPorts, id)
		logger.WithField("released_port", port).Debug("Released auto-assigned port for closed sub-session")
	}
	p.mu.Unlock()

	logger.Debug("Successfully closed sub-session")
	return nil
}

// Close closes the primary session and all associated sub-sessions.
// This method performs a complete cleanup cascade, ensuring that all resources
// are properly released and all sub-sessions are terminated before closing
// the primary session itself. It's safe to call multiple times.
//
// The method first closes all registered sub-sessions, then closes the primary
// session's registry and base session. This prevents resource leaks and ensures
// proper cleanup of the entire session hierarchy.
//
// Example usage:
//
//	defer primary.Close()
func (p *PrimarySession) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	logger := log.WithField("id", p.ID())
	logger.Debug("Closing PrimarySession")

	p.closed = true

	// Close the sub-session registry first, which will close all sub-sessions
	if err := p.registry.Close(); err != nil {
		logger.WithError(err).Error("Failed to close sub-session registry")
		// Continue with base session closure even if registry close fails
	}

	// Clean up all auto-assigned ports
	for id, port := range p.subSessionPorts {
		p.releasePort(port)
		log.WithField("sub_id", id).WithField("port", port).Debug("Released auto-assigned port during primary session close")
	}
	p.subSessionPorts = make(map[string]int) // Clear the mapping
	p.usedPorts = make(map[int]bool)         // Clear all port allocations

	// Close the underlying base session
	if err := p.BaseSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close base session")
		return oops.Errorf("failed to close primary session: %w", err)
	}

	logger.Debug("Successfully closed PrimarySession")
	return nil
}

// Addr returns the I2P address of this primary session.
// This address represents the session's identity on the I2P network and is
// shared by all sub-sessions created from this primary session. The address
// is derived from the primary session's cryptographic keys and remains constant.
//
// Example usage:
//
//	addr := primary.Addr()
//	fmt.Printf("Primary session address: %s", addr.Base32())
func (p *PrimarySession) Addr() i2pkeys.I2PAddr {
	return p.Keys().Addr()
}

// SubSessionCount returns the number of currently active sub-sessions.
// This method provides a quick way to check how many sub-sessions are
// currently managed by this primary session across all types.
//
// Example usage:
//
//	count := primary.SubSessionCount()
//	log.Printf("Primary session managing %d sub-sessions", count)
func (p *PrimarySession) SubSessionCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return 0
	}

	return p.registry.Count()
}

// createSubSAMConnection creates a new SAM connection for sub-sessions.
// This helper method establishes a separate SAM connection that sub-sessions
// can use while maintaining the same configuration and connection parameters
// as the primary session. Each sub-session requires its own SAM connection
// for proper protocol isolation and resource management.
func (p *PrimarySession) createSubSAMConnection() (*common.SAM, error) {
	// Create a new SAM connection using the same host:port as the primary session
	// Extract the address from the existing SAM connection configuration
	address := p.sam.SAMEmit.I2PConfig.SamHost + ":" + strconv.Itoa(p.sam.SAMEmit.I2PConfig.SamPort)

	sam, err := common.NewSAM(address)
	if err != nil {
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	return sam, nil
}

// ensurePortParameter checks if the PORT parameter is present in the options slice.
// If not found, it adds PORT=0 to comply with SAMv3.3 specification which requires
// PORT parameter for DATAGRAM and RAW subsessions. PORT=0 means "any port" which
// allows I2P router to route traffic based on protocol alone.
//
// According to SAMv3.3 spec: "PORT=$port # Required for DATAGRAM* and RAW, invalid for STREAM"
func ensurePortParameter(options []string) []string {
	// Check if PORT parameter already exists in options
	for _, opt := range options {
		if len(opt) >= 5 && (opt[:5] == "PORT=" || opt[:5] == "port=") {
			// PORT parameter already present, return options unchanged
			return options
		}
	}

	// PORT parameter not found, add PORT=0 (default/any port)
	result := make([]string, len(options)+1)
	copy(result, options)
	result[len(options)] = "PORT=0"
	return result
}

// ensureDatagramForwardingParameters ensures PORT and HOST parameters for UDP forwarding.
func ensureDatagramForwardingParameters(options []string, udpPort int) []string {
	hasPort := false
	hasHost := false

	// Check what parameters are already present
	for _, opt := range options {
		if len(opt) >= 5 && (opt[:5] == "PORT=" || opt[:5] == "port=") {
			hasPort = true
		}
		if len(opt) >= 5 && (opt[:5] == "HOST=" || opt[:5] == "host=") {
			hasHost = true
		}
	}

	// Build result with necessary parameters
	result := make([]string, len(options), len(options)+2)
	copy(result, options)

	// Add PORT/HOST to tell SAM bridge where to forward datagrams TO (our UDP listener)
	// Do NOT set sam.udp.port/sam.udp.host - those configure SAM bridge's own UDP port (default 7655)
	if !hasPort {
		result = append(result, fmt.Sprintf("PORT=%d", udpPort)) // Forward to our UDP port
	}
	if !hasHost {
		result = append(result, "HOST=127.0.0.1") // Forward to localhost
	}

	return result
}

// ensureRawForwardingParameters ensures PORT and HOST parameters for UDP forwarding.
func ensureRawForwardingParameters(options []string, udpPort int) []string {
	hasPort := false
	hasHost := false

	// Check what parameters are already present
	for _, opt := range options {
		if len(opt) >= 5 && (opt[:5] == "PORT=" || opt[:5] == "port=") {
			hasPort = true
		}
		if len(opt) >= 5 && (opt[:5] == "HOST=" || opt[:5] == "host=") {
			hasHost = true
		}
	}

	// Build result with necessary parameters
	result := make([]string, len(options), len(options)+2)
	copy(result, options)

	// Add PORT/HOST to tell SAM bridge where to forward datagrams TO (our UDP listener)
	// Do NOT set sam.udp.port/sam.udp.host - those configure SAM bridge's own UDP port (default 7655)
	if !hasPort {
		result = append(result, fmt.Sprintf("PORT=%d", udpPort)) // Forward to our UDP port
	}
	if !hasHost {
		result = append(result, "HOST=127.0.0.1") // Forward to localhost
	}

	return result
}

// ensureDatagram3ForwardingParameters ensures PORT and HOST parameters for UDP forwarding.
func ensureDatagram3ForwardingParameters(options []string, udpPort int) []string {
	hasPort := false
	hasHost := false

	// Check what parameters are already present
	for _, opt := range options {
		if len(opt) >= 5 && (opt[:5] == "PORT=" || opt[:5] == "port=") {
			hasPort = true
		}
		if len(opt) >= 5 && (opt[:5] == "HOST=" || opt[:5] == "host=") {
			hasHost = true
		}
	}

	// Build result with necessary parameters
	result := make([]string, len(options), len(options)+2)
	copy(result, options)

	// Add PORT/HOST to tell SAM bridge where to forward datagrams TO (our UDP listener)
	// Do NOT set sam.udp.port/sam.udp.host - those configure SAM bridge's own UDP port (default 7655)
	if !hasPort {
		result = append(result, fmt.Sprintf("PORT=%d", udpPort)) // Forward to our UDP port
	}
	if !hasHost {
		result = append(result, "HOST=127.0.0.1") // Forward to localhost
	}

	return result
}
