package raw

import (
	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/go-i2p/logger"
)

// SAM wraps common.SAM to provide raw-specific functionality for creating and managing
// raw datagram sessions. This type extends the base SAM functionality with methods
// specifically designed for raw I2P datagram communication.
// SAM wraps common.SAM to provide raw-specific functionality
type SAM struct {
	*common.SAM
}

// NewRawSession creates a new raw session with the SAM bridge using default settings.
// This method establishes a new raw datagram session with the specified ID, keys, and options.
// Raw sessions enable unencrypted datagram transmission over the I2P network.
// NewRawSession creates a new raw session with the SAM bridge
func (s *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error) {
	return NewRawSession(s.SAM, id, keys, options)
}

// NewRawSessionWithSignature creates a new raw session with custom signature type.
// This method allows specifying a custom cryptographic signature type for the session,
// enabling advanced security configurations beyond the default signature algorithm.
// NewRawSessionWithSignature creates a new raw session with custom signature type
func (s *SAM) NewRawSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*RawSession, error) {
	logger := log.WithFields(logger.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new RawSession with signature")

	// Create the base session using the common package with signature
	session, err := s.SAM.NewGenericSessionWithSignature("RAW", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature")
		return nil, oops.Errorf("failed to create raw session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	rs := &RawSession{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created RawSession with signature")
	return rs, nil
}

// NewRawSessionWithPorts creates a new raw session with port specifications.
// This method allows configuring specific port ranges for the session, enabling
// fine-grained control over network communication ports for advanced routing scenarios.
// NewRawSessionWithPorts creates a new raw session with port specifications
func (s *SAM) NewRawSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error) {
	logger := log.WithFields(logger.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	})
	logger.Debug("Creating new RawSession with ports")

	// Create the base session using the common package with ports
	session, err := s.SAM.NewGenericSessionWithSignatureAndPorts("RAW", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with ports")
		return nil, oops.Errorf("failed to create raw session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	rs := &RawSession{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created RawSession with ports")
	return rs, nil
}
