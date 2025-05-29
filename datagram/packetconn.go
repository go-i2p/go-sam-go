package datagram

import (
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
)

// ReadFrom reads a datagram from the connection
func (c *DatagramConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, nil, oops.Errorf("connection is closed")
	}
	c.mu.RUnlock()

	// Start receive loop if not already started
	go c.reader.receiveLoop()

	datagram, err := c.reader.ReceiveDatagram()
	if err != nil {
		return 0, nil, err
	}

	// Copy data to the provided buffer
	n = copy(p, datagram.Data)
	addr = &DatagramAddr{addr: datagram.Source}

	return n, addr, nil
}

// WriteTo writes a datagram to the specified address
func (c *DatagramConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, oops.Errorf("connection is closed")
	}
	c.mu.RUnlock()

	// Convert address to I2P address
	i2pAddr, ok := addr.(*DatagramAddr)
	if !ok {
		return 0, oops.Errorf("address must be a DatagramAddr")
	}

	err = c.writer.SendDatagram(p, i2pAddr.addr)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close closes the datagram connection
func (c *DatagramConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	logger := log.WithField("session_id", c.session.ID())
	logger.Debug("Closing DatagramConn")

	c.closed = true

	// Close reader and writer - these are owned by this connection
	if c.reader != nil {
		c.reader.Close()
	}

	// DO NOT close the session - it's a shared resource that may be used by other connections
	// The session should be closed by the code that created it, not by individual connections
	// that use it. This follows the principle that the creator owns the resource.

	logger.Debug("Successfully closed DatagramConn")
	return nil
}

// LocalAddr returns the local address
func (c *DatagramConn) LocalAddr() net.Addr {
	return &DatagramAddr{addr: c.session.Addr()}
}

// SetDeadline sets the read and write deadlines
func (c *DatagramConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future ReadFrom calls
func (c *DatagramConn) SetReadDeadline(t time.Time) error {
	// For datagrams, we handle timeouts differently
	// This is a placeholder implementation
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls
func (c *DatagramConn) SetWriteDeadline(t time.Time) error {
	// Calculate timeout duration
	if !t.IsZero() {
		timeout := time.Until(t)
		c.writer.SetTimeout(timeout)
	}
	return nil
}

// Read implements net.Conn by wrapping ReadFrom.
// It reads data into the provided byte slice and returns the number of bytes read.
// When reading, it also updates the remote address of the connection.
// Note: This is not a typical use case for datagrams, as they are connectionless.
// However, for compatibility with net.Conn, we implement it this way.
func (c *DatagramConn) Read(b []byte) (n int, err error) {
	n, addr, err := c.ReadFrom(b)
	c.remoteAddr = addr.(*i2pkeys.I2PAddr)
	return n, err
}

// RemoteAddr returns the remote address of the connection.
// For datagram connections, this returns nil as there is no single remote address.
func (c *DatagramConn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return &DatagramAddr{addr: *c.remoteAddr}
	}
	return nil
}

// Write implements net.Conn by wrapping WriteTo.
// It writes data to the remote address and returns the number of bytes written.
// It uses the remote address set by the last Read operation.
// If no remote address is set, it returns an error.
// Note: This is not a typical use case for datagrams, as they are connectionless.
// However, for compatibility with net.Conn, we implement it this way.
func (c *DatagramConn) Write(b []byte) (n int, err error) {
	if c.remoteAddr == nil {
		return 0, oops.Errorf("no remote address set, use WriteTo or Read first")
	}

	addr := &DatagramAddr{addr: *c.remoteAddr}
	return c.WriteTo(b, addr)
}
