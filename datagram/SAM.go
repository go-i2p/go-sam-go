package datagram

import (
	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SAM wraps common.SAM to provide datagram-specific functionality
type SAM struct {
	*common.SAM
}

// NewDatagramSession creates a new datagram session with the SAM bridge
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	return NewDatagramSession(s.SAM, id, keys, options)
}

// NewDatagramSessionWithSignature creates a new datagram session with custom signature type
func (s *SAM) NewDatagramSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*DatagramSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
		"sigType": sigType,
	})
	logger.Debug("Creating new DatagramSession with signature")

	// Create the base session using the common package with signature
	session, err := s.SAM.NewGenericSessionWithSignature("DATAGRAM", id, keys, sigType, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with signature")
		return nil, oops.Errorf("failed to create datagram session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ds := &DatagramSession{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created DatagramSession with signature")
	return ds, nil
}

// NewDatagramSessionWithPorts creates a new datagram session with port specifications
func (s *SAM) NewDatagramSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":       id,
		"fromPort": fromPort,
		"toPort":   toPort,
		"options":  options,
	})
	logger.Debug("Creating new DatagramSession with ports")

	// Create the base session using the common package with ports
	session, err := s.SAM.NewGenericSessionWithSignatureAndPorts("DATAGRAM", id, fromPort, toPort, keys, common.SIG_EdDSA_SHA512_Ed25519, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session with ports")
		return nil, oops.Errorf("failed to create datagram session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ds := &DatagramSession{
		BaseSession: baseSession,
		sam:         s.SAM,
		options:     options,
	}

	logger.Debug("Successfully created DatagramSession with ports")
	return ds, nil
}
