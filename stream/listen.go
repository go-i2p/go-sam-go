package stream

import "github.com/sirupsen/logrus"

// create a new stream listener to accept inbound connections
func (s *StreamSession) Listen() (*StreamListener, error) {
	log.WithFields(logrus.Fields{"id": s.SAM().ID(), "laddr": s.Addr()}).Debug("Creating new StreamListener")
	return &StreamListener{
		session: s,
	}, nil
}
