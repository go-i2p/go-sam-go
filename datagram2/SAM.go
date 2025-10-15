package datagram2

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SAM wraps common.SAM to provide datagram2-specific functionality for I2P messaging.
// This type extends the base SAM functionality with methods specifically designed for
// DATAGRAM2 communication, providing authenticated datagrams with replay protection.
// DATAGRAM2 is the recommended format for new applications, offering enhanced security
// over legacy DATAGRAM sessions through replay attack prevention and offline signature support.
// Example usage: sam := &SAM{SAM: baseSAM}; session, err := sam.NewDatagram2Session(id, keys, options)
type SAM struct {
	*common.SAM
}

// NewDatagram2Session creates a new datagram2 session with the SAM bridge using default settings.
// This method establishes a new DATAGRAM2 session for authenticated UDP-like messaging over I2P
// with replay protection. It uses default signature settings (Ed25519) and automatically configures
// UDP forwarding for SAMv3 compatibility. Session creation can take 2-5 minutes due to I2P tunnel
// establishment, so generous timeouts are recommended.
//
// DATAGRAM2 provides enhanced security compared to legacy DATAGRAM:
//   - Replay protection prevents replay attacks
//   - Offline signature support for advanced key management
//   - Identical SAM API for easy migration
//
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	session, err := sam.NewDatagram2Session("my-session", keys, []string{"inbound.length=1"})
func (s *SAM) NewDatagram2Session(id string, keys i2pkeys.I2PKeys, options []string) (*Datagram2Session, error) {
	// Delegate to the package-level function for session creation
	// This provides consistency with the package API design
	return NewDatagram2Session(s.SAM, id, keys, options)
}

// NewDatagram2SessionWithSignature creates a new datagram2 session with custom signature type.
// This method allows specifying a custom cryptographic signature type for the session,
// enabling advanced security configurations beyond the default Ed25519 algorithm.
// DATAGRAM2 supports offline signatures, allowing pre-signed destinations for enhanced
// privacy and key management flexibility.
//
// Different signature types provide various security levels and compatibility options:
//   - Ed25519 (type 7) - Recommended for most applications
//   - ECDSA (types 1-3) - Legacy compatibility
//   - RedDSA (type 11) - Advanced privacy features
//
// Example usage:
//
//	session, err := sam.NewDatagram2SessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
func (s *SAM) NewDatagram2SessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*Datagram2Session, error) {
	// Log session creation with signature type for debugging
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new Datagram2Session with signature")

	// Create the base session using the common package with custom signature
	// CRITICAL: Use STYLE=DATAGRAM2 (not DATAGRAM) for replay protection
	session, err := s.SAM.NewGenericSessionWithSignature("DATAGRAM2", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature")
		return nil, oops.Errorf("failed to create datagram2 session: %w", err)
	}

	// Ensure the session is of the correct type for datagram2 operations
	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	// Initialize the datagram2 session with the base session and configuration
	ds := &Datagram2Session{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created Datagram2Session with signature")
	return ds, nil
}

// NewDatagram2SessionWithPorts creates a new datagram2 session with port specifications.
// This method allows configuring specific I2CP port ranges for the session, enabling fine-grained
// control over network communication ports for advanced routing scenarios. Port configuration
// is useful for applications requiring specific port mappings or PRIMARY session subsessions.
// This function automatically creates a UDP listener for SAMv3 UDP forwarding (required for v3 mode).
//
// The FROM_PORT and TO_PORT parameters specify I2CP ports for protocol-level communication,
// distinct from the UDP forwarding port which is auto-assigned by the OS.
//
// Example usage:
//
//	session, err := sam.NewDatagram2SessionWithPorts(id, "8080", "8081", keys, options)
func (s *SAM) NewDatagram2SessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*Datagram2Session, error) {
	log.WithFields(logrus.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	}).Debug("Creating new Datagram2Session with ports")

	// Create UDP listener and inject forwarding parameters
	udpConn, options, err := setupUDPListenerWithForwarding(options)
	if err != nil {
		return nil, err
	}

	// Create the base session with port configuration
	baseSession, err := createGenericDatagram2SessionWithPorts(s.SAM, id, fromPort, toPort, keys, options)
	if err != nil {
		udpConn.Close()
		return nil, err
	}

	// Initialize the datagram2 session with UDP forwarding enabled
	ds := &Datagram2Session{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,
	}

	log.Debug("Successfully created Datagram2Session with ports and UDP forwarding")
	return ds, nil
}

// setupUDPListenerWithForwarding creates a UDP listener and injects forwarding parameters.
// This helper creates a UDP listener for SAMv3 datagram forwarding and automatically configures
// the session options to include the HOST and PORT parameters required by the SAM bridge.
// Returns the UDP connection, updated options slice, and any error encountered.
func setupUDPListenerWithForwarding(options []string) (*net.UDPConn, []string, error) {
	// Create UDP listener for receiving forwarded datagrams (SAMv3 requirement)
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

// createGenericDatagram2SessionWithPorts creates a validated BaseSession for datagram2 with ports.
// This helper creates the generic session through the SAM bridge, validates the session type,
// and ensures proper cleanup on error. It uses STYLE=DATAGRAM2 for replay protection.
func createGenericDatagram2SessionWithPorts(sam *common.SAM, id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*common.BaseSession, error) {
	// Create the base session using DATAGRAM2 style for replay protection
	session, err := sam.NewGenericSessionWithSignatureAndPorts("DATAGRAM2", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		log.WithError(err).Error("Failed to create generic session with ports")
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
