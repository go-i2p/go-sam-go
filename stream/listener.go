package stream

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

func (l *StreamListener) From() string {
	return l.session.SAM().Fromport
}

func (l *StreamListener) To() string {
	return l.session.SAM().Toport
}

// get our address
// implements net.Listener
func (l *StreamListener) Addr() net.Addr {
	return l.session.Addr()
}

// implements net.Listener
func (l *StreamListener) Close() error {
	return l.session.Close()
}

// implements net.Listener
func (l *StreamListener) Accept() (net.Conn, error) {
	return l.AcceptI2P()
}

// accept a new inbound connection
func (l *StreamListener) AcceptI2P() (*StreamConn, error) {
	log.Debug("StreamListener.AcceptI2P() called")
	s, err := common.NewSAM(l.session.SAM().Sam())
	if err == nil {
		log.Debug("Connected to SAM bridge")
		// we connected to sam
		// send accept() command
		_, err = io.WriteString(s.Conn, "STREAM ACCEPT ID="+l.session.SAM().ID()+" SILENT=false\n")
		if err != nil {
			log.WithError(err).Error("Failed to send STREAM ACCEPT command")
			s.Close()
			return nil, err
		}
		// read reply
		rd := bufio.NewReader(s.Conn)
		// read first line
		line, err := rd.ReadString(10)
		if err != nil {
			log.WithError(err).Error("Failed to read SAM bridge response")
			s.Close()
			return nil, err
		}
		log.WithField("response", line).Debug("Received SAM bridge response")
		log.Println(line)
		if strings.HasPrefix(line, "STREAM STATUS RESULT=OK") {
			// we gud read destination line
			destline, err := rd.ReadString(10)
			if err == nil {
				dest := common.ExtractDest(destline)
				l.session.SAM().Fromport = common.ExtractPairString(destline, "FROM_PORT")
				l.session.SAM().Toport = common.ExtractPairString(destline, "TO_PORT")
				// return wrapped connection
				dest = strings.Trim(dest, "\n")
				log.WithFields(logrus.Fields{
					"dest": dest,
					"from": l.From(),
					"to":   l.To(),
				}).Debug("Accepted new I2P connection")
				return &StreamConn{
					laddr: l.session.Addr(),
					raddr: i2pkeys.I2PAddr(dest),
					conn:  s.Conn,
				}, nil
			} else {
				log.WithError(err).Error("Failed to read destination line")
				s.Close()
				return nil, err
			}
		} else {
			log.WithField("line", line).Error("Invalid SAM response")
			s.Close()
			return nil, errors.New("invalid sam line: " + line)
		}
	} else {
		log.WithError(err).Error("Failed to connect to SAM bridge")
		s.Close()
		return nil, err
	}
}
