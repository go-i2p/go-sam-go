package stream

import (
	"github.com/go-i2p/i2pkeys"
	"github.com/sirupsen/logrus"
)

// Creates a new StreamSession with the I2CP- and streaminglib options as
// specified. See the I2P documentation for a full list of options.
func (sam SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	log.WithFields(logrus.Fields{"id": id, "options": options}).Debug("Creating new StreamSession")
	conn, err := sam.NewGenericSession("STREAM", id, keys, []string{})
	if err != nil {
		return nil, err
	}
	log.WithField("id", id).Debug("Created new StreamSession")
	streamSession := &StreamSession{
		sam: sam,
	}
	streamSession.Conn = conn
	return streamSession, nil
}

// Creates a new StreamSession with the I2CP- and streaminglib options as
// specified. See the I2P documentation for a full list of options.
func (sam SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	log.WithFields(logrus.Fields{"id": id, "options": options, "sigType": sigType}).Debug("Creating new StreamSession with signature")
	conn, err := sam.NewGenericSessionWithSignature("STREAM", id, keys, sigType, []string{})
	if err != nil {
		return nil, err
	}
	log.WithFields(logrus.Fields{"id": id, "sigType": sigType}).Debug("Created new StreamSession with signature")
	log.WithField("id", id).Debug("Created new StreamSession")
	streamSession := &StreamSession{
		sam: sam,
	}
	streamSession.Conn = conn
	return streamSession, nil
}

// Creates a new StreamSession with the I2CP- and streaminglib options as
// specified. See the I2P documentation for a full list of options.
func (sam SAM) NewStreamSessionWithSignatureAndPorts(id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	log.WithFields(logrus.Fields{"id": id, "from": from, "to": to, "options": options, "sigType": sigType}).Debug("Creating new StreamSession with signature and ports")
	conn, err := sam.NewGenericSessionWithSignatureAndPorts("STREAM", id, from, to, keys, sigType, []string{})
	if err != nil {
		return nil, err
	}
	log.WithFields(logrus.Fields{"id": id, "from": from, "to": to, "sigType": sigType}).Debug("Created new StreamSession with signature and ports")
	log.WithField("id", id).Debug("Created new StreamSession")
	streamSession := &StreamSession{
		sam: sam,
	}
	streamSession.Conn = conn
	return streamSession, nil
}
