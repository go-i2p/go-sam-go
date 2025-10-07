package primary

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/go-i2p/go-sam-go/raw"
	"github.com/go-i2p/go-sam-go/stream"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// PrimarySession provides master session capabilities for managing multiple sub-sessions
// of different types (stream, datagram, raw) within a single I2P session context.
// It enables complex applications with multiple communication patterns while sharing
// the same I2P identity and tunnel infrastructure for enhanced efficiency and anonymity.
//
// The primary session manages the lifecycle of all sub-sessions, ensures proper cleanup
// cascading when the primary session is closed, and provides thread-safe operations
// for creating, managing, and terminating sub-sessions across different protocols.
type PrimarySession struct {
	*common.BaseSession
	sam      *common.SAM
	options  []string
	registry *SubSessionRegistry
	mu       sync.RWMutex
	closed   bool
}

// NewPrimarySession creates a new primary session with the provided SAM connection,
// session ID, cryptographic keys, and configuration options. The primary session
// acts as a master container that can create and manage multiple sub-sessions of
// different types while sharing the same I2P identity and tunnel infrastructure.
//
// The session uses PRIMARY session type in the SAM protocol, which allows multiple
// sub-sessions to be created using the same underlying I2P destination and keys.
// This provides better resource efficiency and maintains consistent identity across
// different communication patterns within the same application.
//
// Example usage:
//
//	session, err := NewPrimarySession(sam, "my-primary", keys, []string{"inbound.length=2"})
//	streamSub, err := session.NewStreamSubSession("stream-1", streamOptions)
//	datagramSub, err := session.NewDatagramSubSession("datagram-1", datagramOptions)
func NewPrimarySession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	logger := log.WithFields(logrus.Fields{
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
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		registry:    NewSubSessionRegistry(),
	}

	logger.Debug("Successfully created PrimarySession")
	return ps, nil
}

// NewPrimarySessionWithSignature creates a new primary session with the specified signature type.
// This is a package-level function that provides direct access to signature-aware session creation
// without requiring wrapper types. It delegates to the common package for session creation while
// maintaining the same primary session functionality and sub-session management capabilities.
//
// The signature type allows specifying custom cryptographic parameters for enhanced security
// or compatibility with specific I2P network configurations. Different signature types provide
// various security levels, performance characteristics, and compatibility options.
//
// Example usage:
//
//	session, err := NewPrimarySessionWithSignature(sam, "secure-primary", keys, options, "EdDSA_SHA512_Ed25519")
//	streamSub, err := session.NewStreamSubSession("stream-1", streamOptions)
func NewPrimarySessionWithSignature(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error) {
	logger := log.WithFields(logrus.Fields{
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
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		registry:    NewSubSessionRegistry(),
	}

	logger.Debug("Successfully created PrimarySession with signature")
	return ps, nil
}

// NewStreamSubSession creates a new stream sub-session within this primary session.
// The sub-session shares the primary session's I2P identity and tunnel infrastructure
// while providing full StreamSession functionality for TCP-like reliable connections.
// Each sub-session must have a unique identifier within the primary session scope.
//
// This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
// the subsession with the primary session's SAM connection, ensuring compliance
// with the I2P SAM protocol specification for PRIMARY session management.
//
// Example usage:
//
//	streamSub, err := primary.NewStreamSubSession("tcp-handler", []string{"FROM_PORT=8080"})
//	listener, err := streamSub.Listen()
//	conn, err := streamSub.Dial("destination.b32.i2p")
func (p *PrimarySession) NewStreamSubSession(id string, options []string) (*StreamSubSession, error) {
	// Use write lock to ensure atomic sub-session creation and prevent race conditions
	// during concurrent session creation operations in I2P SAM protocol
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	logger := log.WithFields(logrus.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
	})
	logger.Debug("Creating stream sub-session")

	// Add the subsession to the primary session using SESSION ADD
	if err := p.sam.AddSubSession("STREAM", id, options); err != nil {
		logger.WithError(err).Error("Failed to add stream subsession")
		return nil, oops.Errorf("failed to create stream sub-session: %w", err)
	}

	// Create a new SAM connection for the sub-session data operations
	// This connection will be used for STREAM CONNECT, STREAM ACCEPT, etc.
	subSAM, err := p.createSubSAMConnection()
	if err != nil {
		logger.WithError(err).Error("Failed to create sub-SAM connection")
		// Clean up the subsession registration
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	// Create the stream session using the new subsession constructor
	// This avoids creating a duplicate session since it's already registered via SESSION ADD
	streamSession, err := stream.NewStreamSessionFromSubsession(subSAM, id, p.Keys(), options)
	if err != nil {
		logger.WithError(err).Error("Failed to create stream session wrapper")
		subSAM.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create stream sub-session: %w", err)
	}

	// Wrap the stream session in a sub-session adapter
	subSession := NewStreamSubSession(id, streamSession)

	// Register the sub-session with the primary session registry
	if err := p.registry.Register(id, subSession); err != nil {
		logger.WithError(err).Error("Failed to register stream sub-session")
		streamSession.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to register stream sub-session: %w", err)
	}

	logger.Debug("Successfully created stream sub-session")
	return subSession, nil
}

// NewUniqueStreamSubSession creates a new unique stream sub-session within this primary session.
func (p *PrimarySession) NewUniqueStreamSubSession(s string) (*StreamSubSession, error) {
	// random number between 1000 and 9999
	randomId := s + "-" + strconv.FormatInt(rand.Int63n(8999)+1000, 10)
	return p.NewStreamSubSession(randomId, nil)
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
	// Use write lock to ensure atomic sub-session creation and prevent race conditions
	// during concurrent session creation operations in I2P SAM protocol
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	logger := log.WithFields(logrus.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
	})
	logger.Debug("Creating datagram sub-session with UDP forwarding")

	// PRIMARY datagram subsessions MUST use UDP forwarding because the control socket
	// is already used by the PRIMARY session. Per SAMv3.md: "If $port is not set,
	// datagrams will NOT be forwarded, they will be received on the control socket"
	// Setup UDP listener for receiving forwarded datagrams
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0") // Port 0 = let OS choose
	if err != nil {
		return nil, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, oops.Errorf("failed to create UDP listener: %w", err)
	}

	// Get the actual port assigned by the OS
	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	logger.WithField("udp_port", udpPort).Debug("Created UDP listener for datagram forwarding")

	// Ensure PORT parameter is present and add UDP forwarding parameters
	// This tells SAM bridge to forward datagrams to our UDP port instead of control socket
	finalOptions := ensureDatagramForwardingParameters(options, udpPort)

	// Add the subsession to the primary session using SESSION ADD
	if err := p.sam.AddSubSession("DATAGRAM", id, finalOptions); err != nil {
		logger.WithError(err).Error("Failed to add datagram subsession")
		udpConn.Close()
		return nil, oops.Errorf("failed to create datagram sub-session: %w", err)
	}

	// Create a new SAM connection for the sub-session data operations (for sending)
	subSAM, err := p.createSubSAMConnection()
	if err != nil {
		logger.WithError(err).Error("Failed to create sub-SAM connection")
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	// Create the datagram session with UDP connection for receiving forwarded datagrams
	datagramSession, err := datagram.NewDatagramSessionFromSubsession(subSAM, id, p.Keys(), options, udpConn)
	if err != nil {
		logger.WithError(err).Error("Failed to create datagram session wrapper")
		subSAM.Close()
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create datagram sub-session: %w", err)
	}

	// Wrap the datagram session in a sub-session adapter
	subSession := NewDatagramSubSession(id, datagramSession)

	// Register the sub-session with the primary session registry
	if err := p.registry.Register(id, subSession); err != nil {
		logger.WithError(err).Error("Failed to register datagram sub-session")
		datagramSession.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to register datagram sub-session: %w", err)
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
	// Use write lock to ensure atomic sub-session creation and prevent race conditions
	// during concurrent session creation operations in I2P SAM protocol
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, oops.Errorf("primary session is closed")
	}

	logger := log.WithFields(logrus.Fields{
		"primary_id": p.ID(),
		"sub_id":     id,
		"options":    options,
	})
	logger.Debug("Creating raw sub-session with UDP forwarding")

	// PRIMARY raw subsessions MUST use UDP forwarding because the control socket
	// is already used by the PRIMARY session. Setup UDP listener for receiving forwarded raw datagrams
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0") // Port 0 = let OS choose
	if err != nil {
		return nil, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, oops.Errorf("failed to create UDP listener: %w", err)
	}

	// Get the actual port assigned by the OS
	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	logger.WithField("udp_port", udpPort).Debug("Created UDP listener for raw datagram forwarding")

	// Ensure PORT parameter and UDP forwarding parameters are present
	finalOptions := ensureRawForwardingParameters(options, udpPort)

	// Add the subsession to the primary session using SESSION ADD
	if err := p.sam.AddSubSession("RAW", id, finalOptions); err != nil {
		logger.WithError(err).Error("Failed to add raw subsession")
		udpConn.Close()
		return nil, oops.Errorf("failed to create raw sub-session: %w", err)
	}

	// Create a new SAM connection for the sub-session data operations (for sending)
	subSAM, err := p.createSubSAMConnection()
	if err != nil {
		logger.WithError(err).Error("Failed to create sub-SAM connection")
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create sub-SAM connection: %w", err)
	}

	// Create the raw session with UDP connection for receiving forwarded raw datagrams
	rawSession, err := raw.NewRawSessionFromSubsession(subSAM, id, p.Keys(), options, udpConn)
	if err != nil {
		logger.WithError(err).Error("Failed to create raw session wrapper")
		subSAM.Close()
		udpConn.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to create raw sub-session: %w", err)
	}

	// Wrap the raw session in a sub-session adapter
	subSession := NewRawSubSession(id, rawSession)

	// Register the sub-session with the primary session registry
	if err := p.registry.Register(id, subSession); err != nil {
		logger.WithError(err).Error("Failed to register raw sub-session")
		rawSession.Close()
		p.sam.RemoveSubSession(id)
		return nil, oops.Errorf("failed to register raw sub-session: %w", err)
	}

	logger.WithField("udp_port", udpPort).Debug("Successfully created raw sub-session with UDP forwarding")
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

	logger := log.WithFields(logrus.Fields{
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

// ensureDatagramForwardingParameters ensures proper UDP forwarding parameters for PRIMARY datagram subsessions.
// Per SAMv3.md, PRIMARY datagram subsessions MUST use UDP forwarding because the control socket is already
// used by the PRIMARY session. This function:
// 1. Ensures PORT parameter is present (required for DATAGRAM subsessions)
// 2. Adds sam.udp.host and sam.udp.port to enable UDP forwarding by SAM bridge
// 3. If sam.udp.port is already present in options, it is preserved
//
// Without these parameters, datagrams would try to use the control socket which is not possible for subsessions.
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

// ensureRawForwardingParameters ensures proper UDP forwarding parameters for PRIMARY raw subsessions.
// Per SAMv3.md, PRIMARY raw subsessions MUST use UDP forwarding because the control socket is already
// used by the PRIMARY session. PORT/HOST tell SAM where to forward datagrams TO (our UDP listener).
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
