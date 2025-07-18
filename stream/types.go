package stream

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// StreamSession represents a streaming session that can create listeners and dialers.
// It provides TCP-like reliable connection capabilities over the I2P network, supporting
// both client and server operations. The session manages the underlying I2P connection and
// provides methods for creating listeners and dialers for stream-based communication.
// Example usage: session, err := NewStreamSession(sam, "my-session", keys, options)
type StreamSession struct {
	*common.BaseSession
	sam       *common.SAM
	options   []string
	listeners []*StreamListener
	mu        sync.RWMutex
	closed    bool
}

// StreamListener implements net.Listener for I2P streaming connections.
// It manages incoming connection acceptance and provides thread-safe operations
// for accepting connections from remote I2P destinations. The listener runs
// an internal accept loop to handle incoming connections asynchronously.
// A finalizer is attached to the listener to ensure that the accept loop is
// cleaned up if the listener is garbage collected without being closed.
// Example usage: listener, err := session.Listen(); conn, err := listener.Accept()
type StreamListener struct {
	session    *StreamSession
	acceptChan chan *StreamConn
	errorChan  chan error
	closeChan  chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	mu         sync.RWMutex
}

// StreamConn implements net.Conn for I2P streaming connections.
// It provides a standard Go networking interface for TCP-like reliable communication
// over I2P networks. The connection supports standard read/write operations with
// proper timeout handling and address information for both local and remote endpoints.
// Example usage: conn, err := session.Dial("destination.b32.i2p"); data, err := conn.Read(buffer)
type StreamConn struct {
	session *StreamSession
	conn    net.Conn
	laddr   i2pkeys.I2PAddr
	raddr   i2pkeys.I2PAddr
	closed  bool
	mu      sync.RWMutex
}

// StreamDialer handles client-side connection establishment for I2P streaming.
// It provides methods for dialing I2P destinations with configurable timeout support
// and context-based cancellation. The dialer can be configured with custom timeouts
// and supports both string destinations and native I2P addresses.
// Example usage: dialer := session.NewDialer().SetTimeout(60*time.Second); conn, err := dialer.Dial("dest.b32.i2p")
type StreamDialer struct {
	session *StreamSession
	timeout time.Duration
}
