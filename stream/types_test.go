package stream

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
)

var (
	ss common.Session = &StreamSession{}
	sl net.Listener   = &StreamListener{}
	sc net.Conn       = &StreamConn{}
)
