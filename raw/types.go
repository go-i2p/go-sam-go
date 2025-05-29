package raw

import (
	"sync"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// RawSession represents a raw session that can send and receive raw datagrams
type RawSession struct {
	*common.BaseSession
	sam     *common.SAM
	options []string
	mu      sync.RWMutex
	closed  bool
}

// RawReader handles incoming raw datagram reception
type RawReader struct {
	session   *RawSession
	recvChan  chan *RawDatagram
	errorChan chan error
	closeChan chan struct{}
	doneChan  chan struct{}
	closed    bool
	mu        sync.RWMutex
}

// RawWriter handles outgoing raw datagram transmission
type RawWriter struct {
	session *RawSession
	timeout time.Duration
}

// RawDatagram represents an I2P raw datagram message
type RawDatagram struct {
	Data   []byte
	Source i2pkeys.I2PAddr
	Local  i2pkeys.I2PAddr
}

// RawAddr implements net.Addr for I2P raw addresses
type RawAddr struct {
	addr i2pkeys.I2PAddr
}

// RawConn implements net.PacketConn for I2P raw datagrams
type RawConn struct {
	session    *RawSession
	reader     *RawReader
	writer     *RawWriter
	remoteAddr *i2pkeys.I2PAddr
	mu         sync.RWMutex
	closed     bool
}

// RawListener implements net.Listener for I2P raw connections
type RawListener struct {
	session    *RawSession
	reader     *RawReader
	acceptChan chan *RawConn
	errorChan  chan error
	closeChan  chan struct{}
	closed     bool
	mu         sync.RWMutex
}
