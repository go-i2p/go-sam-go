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
// DATAGRAM3 communication, providing repliable but UNAUTHENTICATED datagrams with hash-based
// source identification.
//
// ⚠️  SECURITY WARNING: DATAGRAM3 sources are NOT authenticated and can be spoofed!
// ⚠️  Do not trust source addresses without additional application-level authentication.
// ⚠️  If you need authenticated sources, use DATAGRAM2 instead.
//
// DATAGRAM3 uses 32-byte hashes instead of full destinations for source identification,
// reducing overhead at the cost of source verification. Applications requiring source
// authentication MUST implement their own authentication layer.
//
// Example usage: sam := &SAM{SAM: baseSAM}; session, err := sam.NewDatagram3Session(id, keys, options)
type SAM struct {
	*common.SAM
}

// NewDatagram3Session creates a new repliable but UNAUTHENTICATED datagram3 session.
// This method establishes a new DATAGRAM3 session for UDP-like messaging over I2P with
// hash-based source identification. Session creation can take 2-5 minutes due to I2P tunnel
// establishment, so generous timeouts are recommended.
//
// ⚠️  SECURITY WARNING: DATAGRAM3 sources are NOT authenticated and can be spoofed!
// ⚠️  Applications requiring source authentication should use DATAGRAM2 instead.
//
// DATAGRAM3 provides repliable datagrams with minimal overhead by using hash-based source
// identification instead of full authenticated destinations. Received datagrams contain a
// 32-byte hash that must be resolved via NAMING LOOKUP to reply. The session maintains
// a cache to avoid repeated lookups.
//
// Key differences from DATAGRAM and DATAGRAM2:
//   - Repliable: Can reply to sender (like DATAGRAM/DATAGRAM2)
//   - Unauthenticated: Source is NOT verified (unlike DATAGRAM/DATAGRAM2)
//   - Hash-based: Source is 32-byte hash, NOT full destination
//   - Lower overhead: No signature verification required
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
// ⚠️  SECURITY WARNING: Custom signature types do NOT add source authentication to DATAGRAM3!
// ⚠️  Sources remain unauthenticated regardless of signature configuration.
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
	// Log session creation with security warning
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Warn("Creating DATAGRAM3 session: sources are UNAUTHENTICATED and can be spoofed")
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
// ⚠️  SECURITY WARNING: Port configuration does NOT add source authentication to DATAGRAM3!
// ⚠️  Sources remain unauthenticated regardless of port settings.
//
// The FROM_PORT and TO_PORT parameters specify I2CP ports for protocol-level communication,
// distinct from the UDP forwarding port which is auto-assigned by the OS.
//
// Example usage:
//
//	session, err := sam.NewDatagram3SessionWithPorts(id, "8080", "8081", keys, options)
func (s *SAM) NewDatagram3SessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*Datagram3Session, error) {
	// Log session creation with security warning
	logger := log.WithFields(logrus.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	})
	logger.Warn("Creating DATAGRAM3 session with ports: sources are UNAUTHENTICATED")
	logger.Debug("Creating new Datagram3Session with ports")

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
	// HOST and PORT tell the SAM bridge where to forward received datagrams
	options = ensureUDPForwardingParameters(options, udpPort)

	// Create the base session using the common package with port configuration
	// CRITICAL: Use STYLE=DATAGRAM3 (not DATAGRAM or DATAGRAM2)
	session, err := s.SAM.NewGenericSessionWithSignatureAndPorts("DATAGRAM3", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with ports")
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
		sam:         s.SAM,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,
		resolver:    NewHashResolver(s.SAM),
	}

	logger.Debug("Successfully created Datagram3Session with ports and UDP forwarding")
	return ds, nil
}
