package datagram

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SAM wraps common.SAM to provide datagram-specific functionality for I2P messaging.
// This type extends the base SAM functionality with methods specifically designed for
// datagram communication, including session creation with various configuration options
// and signature types for enhanced security and routing control.
// Example usage: sam := &SAM{SAM: baseSAM}; session, err := sam.NewDatagramSession(id, keys, options)
type SAM struct {
	*common.SAM
}

// NewDatagramSession creates a new datagram session with the SAM bridge using default settings.
// This method establishes a new datagram session for UDP-like messaging over I2P with the specified
// session ID, cryptographic keys, and configuration options. It uses default signature settings
// and provides a simple interface for basic datagram communication needs.
// Example usage: session, err := sam.NewDatagramSession("my-session", keys, []string{"inbound.length=1"})
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	// Delegate to the package-level function for session creation
	// This provides consistency with the package API design
	return NewDatagramSession(s.SAM, id, keys, options)
}

// NewDatagramSessionWithSignature creates a new datagram session with custom signature type.
// This method allows specifying a custom cryptographic signature type for the session,
// enabling advanced security configurations beyond the default signature algorithm.
// Different signature types provide various security levels and compatibility options.
// Example usage: session, err := sam.NewDatagramSessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
func (s *SAM) NewDatagramSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*DatagramSession, error) {
	// Log session creation with signature type for debugging
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new DatagramSession with signature")

	// Create the base session using the common package with custom signature
	// This enables advanced cryptographic configuration for enhanced security
	session, err := s.SAM.NewGenericSessionWithSignature("DATAGRAM", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature")
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
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created DatagramSession with signature")
	return ds, nil
}

// NewDatagramSessionWithPorts creates a new datagram session with port specifications.
// This method allows configuring specific port ranges for the session, enabling fine-grained
// control over network communication ports for advanced routing scenarios. Port configuration
// is useful for applications requiring specific port mappings or firewall compatibility.
// This function creates a UDP listener for SAMv3 UDP forwarding (required for v3-only mode).
// Example usage: session, err := sam.NewDatagramSessionWithPorts(id, "8080", "8081", keys, options)
func (s *SAM) NewDatagramSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	log.WithFields(logrus.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	}).Debug("Creating new DatagramSession with ports")

	// Create UDP listener and inject forwarding parameters
	udpConn, options, err := setupDatagramUDPListener(options)
	if err != nil {
		return nil, err
	}

	// Create the base session with port configuration
	baseSession, err := createGenericDatagramSessionWithPorts(s.SAM, id, fromPort, toPort, keys, options)
	if err != nil {
		udpConn.Close()
		return nil, err
	}

	// Initialize the datagram session with UDP forwarding enabled
	ds := &DatagramSession{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
		udpConn:     udpConn,
		udpEnabled:  true,
	}

	log.Debug("Successfully created DatagramSession with ports and UDP forwarding")
	return ds, nil
}

// setupDatagramUDPListener creates a UDP listener and injects forwarding parameters.
// This helper creates a UDP listener for SAMv3 datagram forwarding and automatically configures
// the session options to include the HOST and PORT parameters required by the SAM bridge.
// Returns the UDP connection, updated options slice, and any error encountered.
func setupDatagramUDPListener(options []string) (*net.UDPConn, []string, error) {
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

// createGenericDatagramSessionWithPorts creates a validated BaseSession for datagram with ports.
// This helper creates the generic session through the SAM bridge, validates the session type,
// and ensures proper cleanup on error. It uses STYLE=DATAGRAM for legacy datagram support.
func createGenericDatagramSessionWithPorts(sam *common.SAM, id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*common.BaseSession, error) {
	// Create the base session using DATAGRAM style
	session, err := sam.NewGenericSessionWithSignatureAndPorts("DATAGRAM", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		log.WithError(err).Error("Failed to create generic session with ports")
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
