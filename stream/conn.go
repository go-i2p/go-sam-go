package stream

import (
	"net"
	"time"

	"github.com/samber/oops"
	"github.com/go-i2p/logger"
)

// Read reads data from the connection into the provided buffer.
// This method implements the net.Conn interface and provides thread-safe reading
// from the underlying I2P streaming connection. It handles connection state checking
// and proper error reporting for closed connections.
// Example usage: n, err := conn.Read(buffer)
func (c *StreamConn) Read(b []byte) (int, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, oops.Errorf("connection is closed")
	}
	conn := c.conn
	c.mu.RUnlock()

	n, err := conn.Read(b)
	if err != nil {
		log.WithFields(logger.Fields{
			"local":  c.laddr.Base32(),
			"remote": c.raddr.Base32(),
		}).WithError(err).Debug("Read error")
	}
	return n, err
}

// Write writes data to the connection from the provided buffer.
// This method implements the net.Conn interface and provides thread-safe writing
// to the underlying I2P streaming connection. It handles connection state checking
// and proper error reporting for closed connections.
// Example usage: n, err := conn.Write(data)
func (c *StreamConn) Write(b []byte) (int, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, oops.Errorf("connection is closed")
	}
	conn := c.conn
	c.mu.RUnlock()

	n, err := conn.Write(b)
	if err != nil {
		log.WithFields(logger.Fields{
			"local":  c.laddr.Base32(),
			"remote": c.raddr.Base32(),
		}).WithError(err).Debug("Write error")
	}
	return n, err
}

// Close closes the connection and releases all associated resources.
// This method implements the net.Conn interface and is safe to call multiple times.
// It properly handles concurrent access and ensures clean shutdown of the underlying
// I2P streaming connection with appropriate error handling.
// Example usage: defer conn.Close()
func (c *StreamConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	logger := log.WithFields(logger.Fields{
		"local":  c.laddr.Base32(),
		"remote": c.raddr.Base32(),
	})
	logger.Debug("Closing StreamConn")

	c.closed = true

	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			logger.WithError(err).Error("Failed to close underlying connection")
			return oops.Errorf("failed to close connection: %w", err)
		}
	}

	logger.Debug("Successfully closed StreamConn")
	return nil
}

// LocalAddr returns the local network address of the connection.
// This method implements the net.Conn interface and provides the I2P address
// of the local endpoint. The returned address implements the net.Addr interface
// and can be used for logging or connection management.
// Example usage: localAddr := conn.LocalAddr()
func (c *StreamConn) LocalAddr() net.Addr {
	return &i2pAddr{addr: c.laddr}
}

// RemoteAddr returns the remote network address of the connection.
// This method implements the net.Conn interface and provides the I2P address
// of the remote endpoint. The returned address implements the net.Addr interface
// and can be used for logging, authentication, or connection management.
// Example usage: remoteAddr := conn.RemoteAddr()
func (c *StreamConn) RemoteAddr() net.Addr {
	return &i2pAddr{addr: c.raddr}
}

// SetDeadline sets the read and write deadlines for the connection.
// This method implements the net.Conn interface and sets both read and write
// deadlines to the same time. It provides a convenient way to set overall
// connection timeouts for both read and write operations.
// Example usage: conn.SetDeadline(time.Now().Add(30*time.Second))
func (c *StreamConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls on the connection.
// This method implements the net.Conn interface and allows setting read-specific
// timeouts. A zero time value disables the deadline, and the deadline applies
// to all future and pending Read calls.
// Example usage: conn.SetReadDeadline(time.Now().Add(30*time.Second))
func (c *StreamConn) SetReadDeadline(t time.Time) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return oops.Errorf("connection is nil")
	}

	return conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls on the connection.
// This method implements the net.Conn interface and allows setting write-specific
// timeouts. A zero time value disables the deadline, and the deadline applies
// to all future and pending Write calls.
// Example usage: conn.SetWriteDeadline(time.Now().Add(30*time.Second))
func (c *StreamConn) SetWriteDeadline(t time.Time) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return oops.Errorf("connection is nil")
	}

	return conn.SetWriteDeadline(t)
}
