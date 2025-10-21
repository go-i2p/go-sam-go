package datagram

import (
	"context"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/logger"
	"github.com/samber/oops"
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
	if err := ds.validateDialContext(ctx, destination); err != nil {
		return nil, err
	}

	logger := ds.createDialLogger(destination)
	conn := ds.createDatagramConnection()
	ds.initializeConnection(conn, logger)

	return conn, nil
}

// validateDialContext performs initial validation checks for the dial operation.
func (ds *DatagramSession) validateDialContext(ctx context.Context, destination string) error {
	// Check if the context is already cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate that the session is not closed before attempting connection
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.closed {
		return oops.Errorf("session is closed")
	}

	// Validate that the destination is not empty
	if destination == "" {
		return oops.Errorf("destination cannot be empty")
	}

	return nil
}

// createDialLogger sets up logging context for debugging connection establishment.
func (ds *DatagramSession) createDialLogger(destination string) *logger.Entry {
	logger := log.WithFields(logger.Fields{
		"destination": destination,
		"session_id":  ds.ID(),
	})
	logger.Debug("Dialing datagram destination")
	return logger
}

// createDatagramConnection creates a new datagram connection with integrated reader and writer.
func (ds *DatagramSession) createDatagramConnection() *DatagramConn {
	return &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}
}

// initializeConnection starts the connection's receive loop and sets up cleanup.
func (ds *DatagramSession) initializeConnection(conn *DatagramConn, logger *logger.Entry) {
	// Start the reader's receive loop for continuous datagram processing
	if conn.reader != nil {
		go conn.reader.receiveLoop()
	}

	// Set up cleanup to prevent resource leaks if Close() is not called
	conn.addCleanup()

	logger.Debug("Successfully created datagram connection")
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
	if err := ds.validateDatagramI2PDialContext(ctx); err != nil {
		return nil, err
	}

	if err := ds.validateDatagramI2PSessionState(); err != nil {
		return nil, err
	}

	logger := ds.createDatagramI2PDialLogger(addr)
	conn := ds.createDatagramI2PConnection()
	ds.initializeDatagramI2PConnection(conn, logger)

	return conn, nil
}

// validateDatagramI2PDialContext checks if the context is still valid before dialing.
func (ds *DatagramSession) validateDatagramI2PDialContext(ctx context.Context) error {
	// Check if the context is already cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// validateDatagramI2PSessionState ensures the session is open and ready for dialing.
func (ds *DatagramSession) validateDatagramI2PSessionState() error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if ds.closed {
		return oops.Errorf("session is closed")
	}
	return nil
}

// createDatagramI2PDialLogger sets up logging context for I2P dialing operation.
func (ds *DatagramSession) createDatagramI2PDialLogger(addr i2pkeys.I2PAddr) *logger.Entry {
	// Create logging context for debugging I2P connection establishment
	logger := log.WithFields(logger.Fields{
		"destination": addr.Base32(),
		"session_id":  ds.ID(),
	})
	logger.Debug("Dialing I2P datagram destination")
	return logger
}

// createDatagramI2PConnection creates a new datagram connection with reader and writer.
func (ds *DatagramSession) createDatagramI2PConnection() *DatagramConn {
	return &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}
}

// initializeDatagramI2PConnection starts the connection and sets up cleanup.
func (ds *DatagramSession) initializeDatagramI2PConnection(conn *DatagramConn, logger *logger.Entry) {
	// Start the reader's receive loop for continuous datagram processing
	if conn.reader != nil {
		go conn.reader.receiveLoop()
	}

	// Set up cleanup to prevent resource leaks if Close() is not called
	conn.addCleanup()

	logger.Debug("Successfully created I2P datagram connection")
}
