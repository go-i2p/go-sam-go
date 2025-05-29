package raw

import (
	"net"
	"time"

	"github.com/samber/oops"
)

// ReadFrom reads a raw datagram from the connection
func (c *RawConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
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
	addr = &RawAddr{addr: datagram.Source}

	return n, addr, nil
}

// WriteTo writes a raw datagram to the specified address
func (c *RawConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, oops.Errorf("connection is closed")
	}
	c.mu.RUnlock()

	// Convert address to I2P address
	i2pAddr, ok := addr.(*RawAddr)
	if !ok {
		return 0, oops.Errorf("address must be a RawAddr")
	}

	err = c.writer.SendDatagram(p, i2pAddr.addr)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close closes the raw connection
func (c *RawConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	logger := log.WithField("session_id", c.session.ID())
	logger.Debug("Closing RawConn")

	c.closed = true

	// Close reader and writer - these are owned by this connection
	if c.reader != nil {
		c.reader.Close()
	}

	// DO NOT close the session - it's a shared resource that may be used by other connections

	logger.Debug("Successfully closed RawConn")
	return nil
}

// LocalAddr returns the local address
func (c *RawConn) LocalAddr() net.Addr {
	return &RawAddr{addr: c.session.Addr()}
}

// SetDeadline sets the read and write deadlines
func (c *RawConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future ReadFrom calls
func (c *RawConn) SetReadDeadline(t time.Time) error {
	// For raw datagrams, we handle timeouts differently
	// This is a placeholder implementation
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls
func (c *RawConn) SetWriteDeadline(t time.Time) error {
	// Calculate timeout duration
	if !t.IsZero() {
		timeout := time.Until(t)
		c.writer.SetTimeout(timeout)
	}
	return nil
}

// Read implements net.Conn by wrapping ReadFrom
func (c *RawConn) Read(b []byte) (n int, err error) {
	n, addr, err := c.ReadFrom(b)
	if addr != nil {
		c.remoteAddr = &addr.(*RawAddr).addr
	}
	return n, err
}

// RemoteAddr returns the remote address of the connection
func (c *RawConn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return &RawAddr{addr: *c.remoteAddr}
	}
	return nil
}

// Write implements net.Conn by wrapping WriteTo
func (c *RawConn) Write(b []byte) (n int, err error) {
	if c.remoteAddr == nil {
		return 0, oops.Errorf("no remote address set, use WriteTo or Read first")
	}

	addr := &RawAddr{addr: *c.remoteAddr}
	return c.WriteTo(b, addr)
}
