package stream

import (
	"net"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

type SAM struct {
	common.SAM
}

// Represents a streaming session.
type StreamSession struct {
	sam      SAM
	Timeout  time.Duration
	Deadline time.Time
	net.Conn
}

func (s *StreamSession) SAM() *SAM {
	return &s.sam
}

type StreamListener struct {
	// parent stream session
	session *StreamSession
}

type StreamConn struct {
	laddr i2pkeys.I2PAddr
	raddr i2pkeys.I2PAddr
	conn  net.Conn
}
