package primary

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/stream"
	"github.com/sirupsen/logrus"
)

// Creates a new stream.StreamSession with the I2CP- and streaminglib options as
// specified. See the I2P documentation for a full list of options.
func (sam *PrimarySession) NewStreamSubSession(id string) (*stream.StreamSession, error) {
	log.WithField("id", id).Debug("NewStreamSubSession called")
	conn, err := sam.NewGenericSubSession("STREAM", id, []string{})
	if err != nil {
		log.WithError(err).Error("Failed to create new generic sub-session")
		return nil, err
	}
	return newFromPrimary(sam, conn), nil
}

// Creates a new stream.StreamSession with the I2CP- and streaminglib options as
// specified. See the I2P documentation for a full list of options.
func (sam *PrimarySession) NewUniqueStreamSubSession(id string) (*stream.StreamSession, error) {
	log.WithField("id", id).Debug("NewUniqueStreamSubSession called")
	conn, err := sam.NewGenericSubSession("STREAM", id, []string{})
	if err != nil {
		log.WithError(err).Error("Failed to create new generic sub-session")
		return nil, err
	}
	return newFromPrimary(sam, conn), nil
}

// Creates a new stream.StreamSession with the I2CP- and streaminglib options as
// specified. See the I2P documentation for a full list of options.
func (sam *PrimarySession) NewStreamSubSessionWithPorts(id, from, to string) (*stream.StreamSession, error) {
	log.WithFields(logrus.Fields{"id": id, "from": from, "to": to}).Debug("NewStreamSubSessionWithPorts called")
	conn, err := sam.NewGenericSubSessionWithSignatureAndPorts("STREAM", id, from, to, []string{})
	if err != nil {
		log.WithError(err).Error("Failed to create new generic sub-session with signature and ports")
		return nil, err
	}
	return newFromPrimary(sam, conn), nil
}

func newFromPrimary(sam *PrimarySession, conn net.Conn) *stream.StreamSession {
	streamSession := &stream.StreamSession{
		SAM: &stream.SAM{
			SAM: (*common.SAM)(sam.SAM),
		},
	}
	streamSession.Conn = conn

	return streamSession

}
