package stream

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
)

var ss common.Session = &StreamSession{}
var sl net.Listener = &StreamListener{}
var sc net.Conn = &StreamConn{}
