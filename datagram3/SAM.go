package datagram3

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SAM wraps common.SAM to provide datagram3-specific functionality for I2P messaging.
// This type extends the base SAM functionality with methods specifically designed for
// DATAGRAM3 communication, providing repliable but datagrams with hash-based
// source identification.
//
// DATAGRAM3 uses 32-byte hashes instead of full destinations for source identification,
// reducing overhead at the cost of full destination verification. Applications requiring source
// authentication MUST implement their own authentication layer.
//
// Example usage: sam := &SAM{SAM: baseSAM}; session, err := sam.NewDatagram3Session(id, keys, options)
type SAM struct {
	*common.SAM
}

// NewDatagram3Session creates a new repliable but hash-based datagram3 session.
// This method establishes a new DATAGRAM3 session for UDP-like messaging over I2P with
// hash-based source identification. Session creation can take 2-5 minutes due to I2P tunnel
// establishment, so generous timeouts are recommended.
//
// DATAGRAM3 provides repliable datagrams with minimal overhead by using hash-based source
// identification instead of full with full destinations destinations. Received datagrams contain a
// 32-byte hash that must be resolved via NAMING LOOKUP to reply. The session maintains
// a cache to avoid repeated lookups.
//
// Key differences from DATAGRAM and DATAGRAM2:
//   - Repliable: Can reply to sender (like DATAGRAM/DATAGRAM2)
//   - Unwith full destinations: Source uses hash-based identification (unlike DATAGRAM/DATAGRAM2)
//   - Hash-based: Source is 32-byte hash, NOT full destination
//   - Lower overhead: Hash-based identification required
//   - Reply requires NAMING LOOKUP: Hash must be resolved to full destination
//
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	session, err := sam.NewDatagram3Session("my-session", keys, []string{"inbound.length=1"})
func (s *SAM) NewDatagram3Session(id string, keys i2pkeys.I2PKeys, options []string) (*Datagram3Session, error) {
	// Delegate to the package-level function for session creation
	// This provides consistency with the package API design
	return NewDatagram3Session(s.SAM, id, keys, options)
}

// NewDatagram3SessionWithSignature creates a new datagram3 session with custom signature type.
// This method allows specifying a custom cryptographic signature type for the session,
// enabling advanced security configurations beyond the default Ed25519 algorithm.
// DATAGRAM3 supports offline signatures, allowing pre-signed destinations for enhanced
// privacy and key management flexibility.
//
// Different signature types provide various security levels for the local destination:
//   - Ed25519 (type 7) - Recommended for most applications
//   - ECDSA (types 1-3) - Legacy compatibility
//   - RedDSA (type 11) - Advanced privacy features
//
// Example usage:
//
//	session, err := sam.NewDatagram3SessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
func (s *SAM) NewDatagram3SessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*Datagram3Session, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})

	logger.Debug("Creating new Datagram3Session with signature")

	// Create the base session using the common package with custom signature
	// CRITICAL: Use STYLE=DATAGRAM3 (not DATAGRAM or DATAGRAM2)
	session, err := s.SAM.NewGenericSessionWithSignature("DATAGRAM3", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature")
		return nil, oops.Errorf("failed to create datagram3 session: %w", err)
	}

	// Ensure the session is of the correct type for datagram3 operations
	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	// Initialize the datagram3 session with the base session and hash resolver
	ds := &Datagram3Session{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
		resolver:    NewHashResolver(s.SAM),
	}

	logger.Debug("Successfully created Datagram3Session with signature")
	return ds, nil
}

// NewDatagram3SessionWithPorts creates a new datagram3 session with port specifications.
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
//	session, err := sam.NewDatagram3SessionWithPorts(id, "8080", "8081", keys, options)
func (s *SAM) NewDatagram3SessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*Datagram3Session, error) {
	logger := log.WithFields(logrus.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	})

	logger.Debug("Creating new Datagram3Session with ports")

	// Create UDP listener and get assigned port
	udpConn, udpPort, err := createUDPListener()
	if err != nil {
		return nil, err
	}

	// Inject UDP forwarding parameters into session options
	options = ensureUDPForwardingParameters(options, udpPort)

	// Create the generic session with port configuration
	session, err := createGenericDatagram3Session(s.SAM, id, fromPort, toPort, keys, options)
	if err != nil {
		udpConn.Close() // Clean up UDP listener on error
		return nil, err
	}

	// Wrap session as Datagram3Session with UDP forwarding enabled
	ds, err := wrapAsDatagram3Session(session, s.SAM, options, udpConn, true)
	if err != nil {
		udpConn.Close() // Clean up UDP listener on error
		return nil, err
	}

	logger.Debug("Successfully created Datagram3Session with ports and UDP forwarding")
	return ds, nil
}

// createUDPListener establishes a UDP listener for SAMv3 datagram forwarding.
// This function creates a UDP socket bound to localhost on an OS-assigned port,
// which the SAM bridge will use to forward incoming DATAGRAM3 messages.
// Returns the UDP connection, the assigned port number, and any error encountered.
func createUDPListener() (*net.UDPConn, int, error) {
	// Resolve UDP address on localhost with OS-assigned port
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.WithError(err).Error("Failed to resolve UDP address")
		return nil, 0, oops.Errorf("failed to resolve UDP address: %w", err)
	}

	// Create UDP listener for receiving forwarded datagrams
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.WithError(err).Error("Failed to create UDP listener")
		return nil, 0, oops.Errorf("failed to create UDP listener: %w", err)
	}

	// Extract the actual port assigned by the OS
	udpPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	log.WithField("udp_port", udpPort).Debug("Created UDP listener for datagram3 forwarding")

	return udpConn, udpPort, nil
}

// createGenericDatagram3Session creates and validates a DATAGRAM3 session with port configuration.
// This function establishes a new generic session using the DATAGRAM3 style with specified
// I2CP port ranges for protocol-level communication. The function handles session creation
// and basic error validation, returning the generic session interface for further processing.
func createGenericDatagram3Session(sam *common.SAM, id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (common.Session, error) {
	// Create the base session using DATAGRAM3 style with port configuration
	// CRITICAL: Use STYLE=DATAGRAM3 (not DATAGRAM or DATAGRAM2)
	session, err := sam.NewGenericSessionWithSignatureAndPorts("DATAGRAM3", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		log.WithError(err).Error("Failed to create generic session with ports")
		return nil, oops.Errorf("failed to create datagram3 session: %w", err)
	}

	return session, nil
}

// wrapAsDatagram3Session validates and wraps a generic session as a Datagram3Session.
// This function performs type assertion to ensure the session is a BaseSession,
// then initializes a Datagram3Session with the provided configuration including
// UDP forwarding support and hash resolver for source identification.
func wrapAsDatagram3Session(session common.Session, sam *common.SAM, options []string, udpConn *net.UDPConn, udpEnabled bool) (*Datagram3Session, error) {
	// Validate session type for datagram3 operations
	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		log.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	// Initialize the datagram3 session with all components
	ds := &Datagram3Session{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  udpEnabled,
		resolver:    NewHashResolver(sam),
	}

	return ds, nil
}
