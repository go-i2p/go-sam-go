package stream

import (
	"net"
	"sync"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// StreamSession represents a streaming session that can create listeners and dialers
type StreamSession struct {
	*common.BaseSession
	sam     *common.SAM
	options []string
	mu      sync.RWMutex
	closed  bool
}

// StreamListener implements net.Listener for I2P streaming connections
type StreamListener struct {
	session    *StreamSession
	acceptChan chan *StreamConn
	errorChan  chan error
	closeChan  chan struct{}
	closed     bool
	mu         sync.RWMutex
}

// StreamConn implements net.Conn for I2P streaming connections
type StreamConn struct {
	session *StreamSession
	conn    net.Conn
	laddr   i2pkeys.I2PAddr
	raddr   i2pkeys.I2PAddr
	closed  bool
	mu      sync.RWMutex
}

// StreamDialer handles client-side connection establishment
type StreamDialer struct {
	session *StreamSession
	timeout time.Duration
}
