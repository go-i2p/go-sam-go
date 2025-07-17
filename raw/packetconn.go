package raw

import (
	"net"
	"time"

	"github.com/samber/oops"
)

// ReadFrom reads a raw datagram from the connection.
// This method implements the net.PacketConn interface and blocks until a datagram
// is received or an error occurs, returning the data, source address, and any error.
// Example usage: n, addr, err := conn.ReadFrom(buffer)
func (c *RawConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, nil, oops.Errorf("connection is closed")
	}
	c.mu.RUnlock()

	datagram, err := c.reader.ReceiveDatagram()
	if err != nil {
		return 0, nil, err
	}

	// Copy data to the provided buffer
	n = copy(p, datagram.Data)
	addr = &RawAddr{addr: datagram.Source}

	return n, addr, nil
}

// WriteTo writes a raw datagram to the specified address.
// This method implements the net.PacketConn interface and sends the data
// to the destination address, returning the number of bytes written and any error.
// Example usage: n, err := conn.WriteTo(data, destAddr)
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

// Close closes the raw connection and cleans up associated resources.
// This method is safe to call multiple times and will only perform cleanup once.
// The underlying session remains open and can be used by other connections.
// Example usage: defer conn.Close()
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

// LocalAddr returns the local address of the connection.
// This method implements the net.PacketConn interface and returns the I2P address
// of the session wrapped in a RawAddr for compatibility with net.Addr.
// Example usage: addr := conn.LocalAddr()
func (c *RawConn) LocalAddr() net.Addr {
	return &RawAddr{addr: c.session.Addr()}
}

// SetDeadline sets the read and write deadlines for the connection.
// This method implements the net.PacketConn interface by applying the deadline
// to both read and write operations through separate deadline methods.
// Example usage: conn.SetDeadline(time.Now().Add(30*time.Second))
func (c *RawConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future ReadFrom calls.
// This method implements the net.PacketConn interface for timeout support.
// Currently this is a placeholder implementation for I2P raw datagrams.
// Example usage: conn.SetReadDeadline(time.Now().Add(10*time.Second))
func (c *RawConn) SetReadDeadline(t time.Time) error {
	// For raw datagrams, we handle timeouts differently
	// This is a placeholder implementation
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls.
// This method implements the net.PacketConn interface by configuring the writer timeout
// based on the deadline duration, providing timeout support for send operations.
// Example usage: conn.SetWriteDeadline(time.Now().Add(5*time.Second))
func (c *RawConn) SetWriteDeadline(t time.Time) error {
	// Calculate timeout duration from deadline and apply to writer
	// Zero deadline means no timeout should be applied
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
