package stream

import (
	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/go-i2p/logger"
)

// SAM wraps common.SAM to provide stream-specific functionality and convenience methods.
// It extends the base SAM connection with streaming-specific session creation methods,
// providing a more convenient API for creating streaming sessions without requiring
// direct interaction with the generic session creation methods.
// Example usage: sam := &SAM{SAM: commonSAM}; session, err := sam.NewStreamSession("id", keys, options)
type SAM struct {
	*common.SAM
}

// NewStreamSession creates a new streaming session with the SAM bridge using default signature.
// This is a convenience method that wraps the generic session creation with streaming-specific
// parameters. It uses the default Ed25519 signature type and provides a simpler API for
// creating streaming sessions without requiring explicit signature type specification.
// Example usage: session, err := sam.NewStreamSession("my-session", keys, options)
func (s *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	return NewStreamSession(s.SAM, id, keys, options)
}

// NewStreamSessionWithSignature creates a new streaming session with custom signature type.
// This method provides advanced control over the cryptographic signature type used for
// the I2P destination. It supports various signature algorithms like Ed25519, ECDSA,
// and DSA, allowing applications to choose the most appropriate signature type for their needs.
// Example usage: session, err := sam.NewStreamSessionWithSignature("my-session", keys, options, "EdDSA_SHA512_Ed25519")
func (s *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	logger := log.WithFields(logger.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new StreamSession with signature")

	// Create the base session using the common package with signature
	session, err := s.SAM.NewGenericSessionWithSignature("STREAM", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature")
		return nil, oops.Errorf("failed to create stream session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ss := &StreamSession{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created StreamSession with signature")
	return ss, nil
}

// NewStreamSessionWithPorts creates a new streaming session with port specifications
func (s *SAM) NewStreamSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	logger := log.WithFields(logger.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	})
	logger.Debug("Creating new StreamSession with ports")

	// Create the base session using the common package with ports
	session, err := s.SAM.NewGenericSessionWithSignatureAndPorts("STREAM", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with ports")
		return nil, oops.Errorf("failed to create stream session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ss := &StreamSession{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created StreamSession with ports")
	return ss, nil
}
