package datagram

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
)

var (
	ds  common.Session = &DatagramSession{}
	dl  net.Listener   = &DatagramListener{}
	dc  net.PacketConn = &DatagramConn{}
	dcc net.Conn       = &DatagramConn{}
)
