package datagram2

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
)

// Compile-time interface checks to ensure types implement expected interfaces
var (
	ds  common.Session = &Datagram2Session{}
	dc  net.PacketConn = &Datagram2Conn{}
	dcc net.Conn       = &Datagram2Conn{}
	da  net.Addr       = &Datagram2Addr{}
)
