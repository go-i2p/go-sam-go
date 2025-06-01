package raw

import (
	"net"

	"github.com/go-i2p/go-sam-go/common"
)

var (
	ds common.Session = &RawSession{}
	dl net.Listener   = &RawListener{}
	dc net.PacketConn = &RawConn{}
)
