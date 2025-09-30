package datagram

import (
	"context"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Dial establishes a datagram connection to the specified I2P destination.
// This method creates a net.PacketConn interface for sending and receiving datagrams
// with the specified destination. It uses a default timeout of 30 seconds for connection
// establishment and provides UDP-like communication over I2P networks.
// Example usage: conn, err := session.Dial("destination.b32.i2p")
func (ds *DatagramSession) Dial(destination string) (net.PacketConn, error) {
	// Use the timeout variant with default 30-second timeout
	// This provides reasonable default behavior for most applications
	return ds.DialTimeout(destination, 30*time.Second)
}

// DialTimeout establishes a datagram connection with specified timeout duration.
// This method creates a net.PacketConn interface with timeout support for connection
// establishment. Zero or negative timeout values disable the timeout mechanism.
// The timeout only applies to the initial connection setup, not to subsequent operations.
// Example usage: conn, err := session.DialTimeout("destination.b32.i2p", 60*time.Second)
func (ds *DatagramSession) DialTimeout(destination string, timeout time.Duration) (net.PacketConn, error) {
	// Handle zero or negative timeout by disabling timeout completely
	if timeout <= 0 {
		return ds.DialContext(context.Background(), destination)
	}

	// Create a context with the specified timeout for connection establishment
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ds.DialContext(ctx, destination)
}

// DialContext establishes a datagram connection with context support for cancellation.
// This method provides the core dialing functionality with context-based cancellation support,
// allowing for proper resource cleanup and operation cancellation through the provided context.
// It validates the destination and session state before attempting connection establishment.
// Example usage: conn, err := session.DialContext(ctx, "destination.b32.i2p")
func (ds *DatagramSession) DialContext(ctx context.Context, destination string) (net.PacketConn, error) {
	// Check if the context is already cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate that the session is not closed before attempting connection
	ds.mu.RLock()
	if ds.closed {
		ds.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	ds.mu.RUnlock()

	// Validate that the destination is not empty
	if destination == "" {
		return nil, oops.Errorf("destination cannot be empty")
	}

	// Create logging context for debugging connection establishment
	logger := log.WithFields(logrus.Fields{
		"destination": destination,
		"session_id":  ds.ID(),
	})
	logger.Debug("Dialing datagram destination")

	// Create a datagram connection with integrated reader and writer
	// This provides the net.PacketConn interface for standard Go networking compatibility
	conn := &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}

	// Start the reader's receive loop for continuous datagram processing
	if conn.reader != nil {
		go conn.reader.receiveLoop()
	}

	// Set up cleanup to prevent resource leaks if Close() is not called
	conn.addCleanup()

	logger.Debug("Successfully created datagram connection")
	return conn, nil
}

// DialI2P establishes a datagram connection to an I2P address using native addressing.
// This method creates a net.PacketConn interface for communicating with the specified I2P
// address using the native i2pkeys.I2PAddr type. It uses a default timeout of 30 seconds
// and provides type-safe addressing for I2P destinations.
// Example usage: conn, err := session.DialI2P(i2pAddress)
func (ds *DatagramSession) DialI2P(addr i2pkeys.I2PAddr) (net.PacketConn, error) {
	// Use the timeout variant with default 30-second timeout
	return ds.DialI2PTimeout(addr, 30*time.Second)
}

// DialI2PTimeout establishes a datagram connection to an I2P address with timeout.
// This method provides time-bounded connection establishment using native I2P addressing.
// Zero or negative timeout values disable the timeout mechanism. The timeout only applies
// to the initial connection setup, not to subsequent datagram operations.
// Example usage: conn, err := session.DialI2PTimeout(i2pAddress, 60*time.Second)
func (ds *DatagramSession) DialI2PTimeout(addr i2pkeys.I2PAddr, timeout time.Duration) (net.PacketConn, error) {
	// Handle zero or negative timeout by disabling timeout completely
	if timeout <= 0 {
		return ds.DialI2PContext(context.Background(), addr)
	}

	// Create a context with the specified timeout for connection establishment
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ds.DialI2PContext(ctx, addr)
}

// DialI2PContext establishes a datagram connection to an I2P address with context support.
// This method provides the core I2P dialing functionality with context-based cancellation,
// allowing for proper resource cleanup and operation cancellation through the provided context.
// It validates the session state and creates a connection with integrated reader and writer.
// Example usage: conn, err := session.DialI2PContext(ctx, i2pAddress)
func (ds *DatagramSession) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (net.PacketConn, error) {
	// Check if the context is already cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate that the session is not closed before attempting connection
	ds.mu.RLock()
	if ds.closed {
		ds.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	ds.mu.RUnlock()

	// Create logging context for debugging I2P connection establishment
	logger := log.WithFields(logrus.Fields{
		"destination": addr.Base32(),
		"session_id":  ds.ID(),
	})
	logger.Debug("Dialing I2P datagram destination")

	// Create a datagram connection with integrated reader and writer
	conn := &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}

	// Start the reader's receive loop for continuous datagram processing
	if conn.reader != nil {
		go conn.reader.receiveLoop()
	}

	// Set up cleanup to prevent resource leaks if Close() is not called
	conn.addCleanup()

	logger.Debug("Successfully created I2P datagram connection")
	return conn, nil
}
