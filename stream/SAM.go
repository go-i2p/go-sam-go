package stream

import (
	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SAM wraps common.SAM to provide stream-specific functionality
type SAM struct {
	*common.SAM
}

// NewStreamSession creates a new streaming session with the SAM bridge
func (s *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	return NewStreamSession(s.SAM, id, keys, options)
}

// NewStreamSessionWithSignature creates a new streaming session with custom signature type
func (s *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	logger := log.WithFields(logrus.Fields{
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
	logger := log.WithFields(logrus.Fields{
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
