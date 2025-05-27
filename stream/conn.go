package stream

import (
	"net"
	"time"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Read reads data from the connection
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
		log.WithFields(logrus.Fields{
			"local":  c.laddr.Base32(),
			"remote": c.raddr.Base32(),
		}).WithError(err).Debug("Read error")
	}
	return n, err
}

// Write writes data to the connection
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
		log.WithFields(logrus.Fields{
			"local":  c.laddr.Base32(),
			"remote": c.raddr.Base32(),
		}).WithError(err).Debug("Write error")
	}
	return n, err
}

// Close closes the connection
func (c *StreamConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	logger := log.WithFields(logrus.Fields{
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

// LocalAddr returns the local network address
func (c *StreamConn) LocalAddr() net.Addr {
	return &i2pAddr{addr: c.laddr}
}

// RemoteAddr returns the remote network address
func (c *StreamConn) RemoteAddr() net.Addr {
	return &i2pAddr{addr: c.raddr}
}

// SetDeadline sets the read and write deadlines
func (c *StreamConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
func (c *StreamConn) SetReadDeadline(t time.Time) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return oops.Errorf("connection is nil")
	}

	return conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
func (c *StreamConn) SetWriteDeadline(t time.Time) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return oops.Errorf("connection is nil")
	}

	return conn.SetWriteDeadline(t)
}
