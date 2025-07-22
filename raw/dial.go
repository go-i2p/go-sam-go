package raw

import (
	"context"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Dial establishes a raw connection to the specified I2P destination address.
// This method creates a net.PacketConn interface for sending and receiving raw datagrams
// with the specified destination. It uses a default timeout of 30 seconds.
// Dial establishes a raw connection to the specified destination
func (rs *RawSession) Dial(destination string) (net.PacketConn, error) {
	return rs.DialTimeout(destination, 30*time.Second)
}

// DialTimeout establishes a raw connection with a specified timeout duration.
// This method creates a net.PacketConn interface with timeout support, allowing
// for time-bounded connection establishment. Zero or negative timeout values disable the timeout.
// DialTimeout establishes a raw connection with a timeout
func (rs *RawSession) DialTimeout(destination string, timeout time.Duration) (net.PacketConn, error) {
	// Handle zero or negative timeout - no timeout should be applied
	if timeout <= 0 {
		return rs.DialContext(context.Background(), destination)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rs.DialContext(ctx, destination)
}

// DialContext establishes a raw connection with context support for cancellation.
// This method provides the core dialing functionality with context-based cancellation support,
// allowing for proper resource cleanup and operation cancellation through the provided context.
// DialContext establishes a raw connection with context support
func (rs *RawSession) DialContext(ctx context.Context, destination string) (net.PacketConn, error) {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate session state first
	rs.mu.RLock()
	if rs.closed {
		rs.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	rs.mu.RUnlock()

	// Validate destination
	if destination == "" {
		return nil, oops.Errorf("destination cannot be empty")
	}

	logger := log.WithFields(logrus.Fields{
		"destination": destination,
	})
	logger.Debug("Dialing raw destination")

	// Create a raw connection
	conn := &RawConn{
		session: rs,
		reader:  rs.NewReader(),
		writer:  rs.NewWriter(),
	}

	// Start the reader loop once for this connection
	if conn.reader != nil {
		go conn.reader.receiveLoop()
	}

	// Set up finalizer to prevent resource leaks if Close() is not called
	conn.setFinalizer()

	logger.WithField("session_id", rs.ID()).Debug("Successfully created raw connection")
	return conn, nil
}

// DialI2P establishes a raw connection to an I2P address using native I2P addressing.
// This method creates a net.PacketConn interface for communicating with the specified I2P address
// using the native i2pkeys.I2PAddr type. It uses a default timeout of 30 seconds.
// DialI2P establishes a raw connection to an I2P address
func (rs *RawSession) DialI2P(addr i2pkeys.I2PAddr) (net.PacketConn, error) {
	return rs.DialI2PTimeout(addr, 30*time.Second)
}

// DialI2PTimeout establishes a raw connection to an I2P address with timeout support.
// This method provides time-bounded connection establishment using native I2P addressing.
// Zero or negative timeout values disable the timeout mechanism.
// DialI2PTimeout establishes a raw connection to an I2P address with timeout
func (rs *RawSession) DialI2PTimeout(addr i2pkeys.I2PAddr, timeout time.Duration) (net.PacketConn, error) {
	// Handle zero or negative timeout - no timeout should be applied
	if timeout <= 0 {
		return rs.DialI2PContext(context.Background(), addr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rs.DialI2PContext(ctx, addr)
}

// DialI2PContext establishes a raw connection to an I2P address with context support.
// This method provides the core I2P dialing functionality with context-based cancellation,
// allowing for proper resource cleanup and operation cancellation through the provided context.
// DialI2PContext establishes a raw connection to an I2P address with context support
func (rs *RawSession) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (net.PacketConn, error) {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate session state first
	rs.mu.RLock()
	if rs.closed {
		rs.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	rs.mu.RUnlock()

	logger := log.WithFields(logrus.Fields{
		"destination": addr.Base32(),
	})
	logger.Debug("Dialing I2P raw destination")

	// Create a raw connection
	conn := &RawConn{
		session: rs,
		reader:  rs.NewReader(),
		writer:  rs.NewWriter(),
	}

	// Start the reader loop once for this connection
	if conn.reader != nil {
		go conn.reader.receiveLoop()
	}

	// Set up finalizer to prevent resource leaks if Close() is not called
	conn.setFinalizer()

	logger.WithField("session_id", rs.ID()).Debug("Successfully created I2P raw connection")
	return conn, nil
}
