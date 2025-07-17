package datagram

import (
	"sync"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// DatagramSession represents a datagram session that can send and receive datagrams
type DatagramSession struct {
	*common.BaseSession
	sam     *common.SAM
	options []string
	mu      sync.RWMutex
	closed  bool
}

// DatagramReader handles incoming datagram reception
type DatagramReader struct {
	session   *DatagramSession
	recvChan  chan *Datagram
	errorChan chan error
	closeChan chan struct{}
	doneChan  chan struct{}
	closed    bool
	mu        sync.RWMutex
	closeOnce sync.Once
}

// DatagramWriter handles outgoing datagram transmission
type DatagramWriter struct {
	session *DatagramSession
	timeout time.Duration
}

// Datagram represents an I2P datagram message
type Datagram struct {
	Data   []byte
	Source i2pkeys.I2PAddr
	Local  i2pkeys.I2PAddr
}

// DatagramAddr implements net.Addr for I2P datagram addresses
type DatagramAddr struct {
	addr i2pkeys.I2PAddr
}

// DatagramConn implements net.PacketConn for I2P datagrams
type DatagramConn struct {
	session    *DatagramSession
	reader     *DatagramReader
	writer     *DatagramWriter
	remoteAddr *i2pkeys.I2PAddr
	mu         sync.RWMutex
	closed     bool
}
