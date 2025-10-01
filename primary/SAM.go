package primary

import (
	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SAM wraps common.SAM to provide primary session functionality for creating and managing
// master sessions that can contain multiple sub-sessions of different types. This type
// extends the base SAM functionality with methods specifically designed for primary session
// management, including session creation with various configuration options and signature types.
// Example usage: sam := &SAM{SAM: baseSAM}; session, err := sam.NewPrimarySession(id, keys, options)
type SAM struct {
	*common.SAM
}

// NewPrimarySession creates a new primary session with the SAM bridge using default settings.
// This method establishes a new primary session for managing multiple sub-sessions over I2P
// with the specified session ID, cryptographic keys, and configuration options. It uses default
// signature settings and provides a simple interface for basic primary session needs.
//
// The primary session acts as a master container that can create and manage multiple sub-sessions
// of different types (stream, datagram, raw) while sharing the same I2P identity and tunnel
// infrastructure for enhanced efficiency and consistent anonymity properties.
//
// Example usage:
//
//	session, err := sam.NewPrimarySession("my-primary", keys, []string{"inbound.length=2"})
//	streamSub, err := session.NewStreamSubSession("stream-1", streamOptions)
func (s *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	// Delegate to the package-level function for session creation
	// This provides consistency with the package API design pattern
	return NewPrimarySession(s.SAM, id, keys, options)
}

// NewPrimarySessionWithSignature creates a new primary session with custom signature type.
// This method allows specifying a custom cryptographic signature type for the session,
// enabling advanced security configurations beyond the default signature algorithm.
// Different signature types provide various security levels, compatibility options,
// and performance characteristics for different I2P network requirements.
//
// The primary session created with custom signature maintains the same multi-session
// management capabilities while using the specified cryptographic parameters for
// enhanced security or compatibility with specific I2P network configurations.
//
// Example usage:
//
//	session, err := sam.NewPrimarySessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
//	datagramSub, err := session.NewDatagramSubSession("datagram-1", datagramOptions)
func (s *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error) {
	// Log session creation with signature type for debugging and monitoring
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new PrimarySession with signature")

	// Create the base session using the common package with custom signature
	// This enables advanced cryptographic configuration for enhanced security
	session, err := s.SAM.NewGenericSessionWithSignature("PRIMARY", id, keys, sigType, options)
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
		sam:         s.SAM,
		options:     options,
		registry:    NewSubSessionRegistry(),
	}

	logger.Debug("Successfully created PrimarySession with signature")
	return ps, nil
}

// NewPrimarySessionWithPorts creates a new primary session with port specifications.
// This method allows configuring specific port ranges for the session, enabling fine-grained
// control over network communication ports for advanced routing scenarios. Port configuration
// is useful for applications requiring specific port mappings, firewall compatibility,
// or integration with existing network infrastructure and service discovery mechanisms.
//
// The primary session created with port configuration maintains full multi-session management
// capabilities while using the specified port parameters for network communication optimization
// and compatibility with existing network configurations or security requirements.
//
// Example usage:
//
//	session, err := sam.NewPrimarySessionWithPorts(id, "8080", "8081", keys, options)
//	rawSub, err := session.NewRawSubSession("raw-1", rawOptions)
func (s *SAM) NewPrimarySessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	// Log session creation with port configuration for debugging and network analysis
	logger := log.WithFields(logrus.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	})
	logger.Debug("Creating new PrimarySession with ports")

	// Create the base session using the common package with port configuration
	// This enables advanced port management for specific networking requirements
	session, err := s.SAM.NewGenericSessionWithSignatureAndPorts("PRIMARY", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic primary session with ports")
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
		sam:         s.SAM,
		options:     options,
		registry:    NewSubSessionRegistry(),
	}

	logger.Debug("Successfully created PrimarySession with ports")
	return ps, nil
}
