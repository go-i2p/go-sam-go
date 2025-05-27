// package sam3 wraps the original sam3 API from github.com/go-i2p/sam3
package sam3

import (
	"github.com/go-i2p/go-sam-go/primary"
)

const (
	session_ADDOK = "SESSION STATUS RESULT=OK"
)

// Represents a primary session.
type PrimarySession struct {
	*primary.PrimarySession
}

var PrimarySessionSwitch = "MASTER"

func (p *PrimarySession) NewStreamSubSession(id string) (*StreamSession, error) {
	log.WithField("id", id).Debug("NewStreamSubSession called")
	session, err := p.PrimarySession.NewStreamSubSession(id)
	if err != nil {
		return nil, err
	}
	return &StreamSession{
		StreamSession: session,
	}, nil
}
