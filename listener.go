package sam3

import (
	"net"

	"github.com/go-i2p/go-sam-go/stream"
)

// StreamListener implements net.Listener for I2P streams
type StreamListener struct {
	listener *stream.StreamListener
}

// Accept accepts new inbound connections
func (l *StreamListener) Accept() (net.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &SAMConn{conn: conn}, nil
}

// AcceptI2P accepts a new inbound I2P connection
func (l *StreamListener) AcceptI2P() (*SAMConn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &SAMConn{conn: conn}, nil
}

// Addr returns the listener's address
func (l *StreamListener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close closes the listener
func (l *StreamListener) Close() error {
	return l.listener.Close()
}

// From returns the from port
func (l *StreamListener) From() string {
	return ""
}

// To returns the to port
func (l *StreamListener) To() string {
	return ""
}
