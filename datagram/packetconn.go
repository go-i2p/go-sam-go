package datagram

import (
	"net"
	"runtime"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
)

// ReadFrom reads a datagram from the connection and returns the number of bytes read,
// the source address, and any error encountered. This method implements the net.PacketConn interface.
// It starts the receive loop if not already started and blocks until a datagram is received.
// The data is copied to the provided buffer p, and the source address is returned as a DatagramAddr.
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

// WriteTo writes a datagram to the specified address and returns the number of bytes written
// and any error encountered. This method implements the net.PacketConn interface.
// The address must be a DatagramAddr containing a valid I2P destination.
// The entire byte slice p is sent as a single datagram message.
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

// Close closes the datagram connection and releases associated resources.
// This method implements the net.Conn interface. It closes the reader and writer
// but does not close the underlying session, which may be shared by other connections.
// Multiple calls to Close are safe and will return nil after the first call.
func (c *DatagramConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	var sessionID string
	if c.session != nil {
		sessionID = c.session.ID()
	} else {
		sessionID = "unknown"
	}
	logger := log.WithField("session_id", sessionID)
	logger.Debug("Closing DatagramConn")

	c.closed = true

	// Clear the finalizer since we're cleaning up explicitly
	c.clearFinalizer()

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

// LocalAddr returns the local network address as a DatagramAddr containing
// the I2P destination address of this connection's session. This method implements
// the net.Conn interface and provides access to the local I2P destination.
func (c *DatagramConn) LocalAddr() net.Addr {
	return &DatagramAddr{addr: c.session.Addr()}
}

// SetDeadline sets both read and write deadlines for the connection.
// This method implements the net.Conn interface by calling both SetReadDeadline
// and SetWriteDeadline with the same time value. If either deadline cannot be set,
// the first error encountered is returned.
func (c *DatagramConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future ReadFrom calls.
// This method implements the net.Conn interface. For datagram connections,
// this is currently a placeholder implementation that always returns nil.
// Timeout handling is managed differently for datagram operations.
func (c *DatagramConn) SetReadDeadline(t time.Time) error {
	// For datagrams, we handle timeouts differently
	// This is a placeholder implementation
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls.
// This method implements the net.Conn interface. If the deadline is not zero,
// it calculates the timeout duration and sets it on the writer for subsequent
// write operations.
func (c *DatagramConn) SetWriteDeadline(t time.Time) error {
	// Calculate timeout duration
	if !t.IsZero() {
		timeout := time.Until(t)
		c.writer.SetTimeout(timeout)
	}
	return nil
}

// Read implements net.Conn by wrapping ReadFrom for stream-like usage.
// It reads data into the provided byte slice and returns the number of bytes read.
// When reading, it also updates the remote address of the connection for subsequent
// Write calls. Note: This is not typical for datagrams which are connectionless,
// but provides compatibility with the net.Conn interface.
func (c *DatagramConn) Read(b []byte) (n int, err error) {
	n, addr, err := c.ReadFrom(b)
	c.remoteAddr = addr.(*i2pkeys.I2PAddr)
	return n, err
}

// RemoteAddr returns the remote network address of the connection.
// This method implements the net.Conn interface. For datagram connections,
// this returns the address of the last peer that sent data (set by Read),
// or nil if no data has been received yet.
func (c *DatagramConn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return &DatagramAddr{addr: *c.remoteAddr}
	}
	return nil
}

// Write implements net.Conn by wrapping WriteTo for stream-like usage.
// It writes data to the remote address set by the last Read operation and
// returns the number of bytes written. If no remote address has been set,
// it returns an error. Note: This is not typical for datagrams which are
// connectionless, but provides compatibility with the net.Conn interface.
func (c *DatagramConn) Write(b []byte) (n int, err error) {
	if c.remoteAddr == nil {
		return 0, oops.Errorf("no remote address set, use WriteTo or Read first")
	}

	addr := &DatagramAddr{addr: *c.remoteAddr}
	return c.WriteTo(b, addr)
}

// connFinalizer is called by the garbage collector to ensure resources are cleaned up
// even if the user forgets to call Close(). This prevents goroutine leaks.
func datagramConnFinalizer(c *DatagramConn) {
	c.mu.Lock()
	if !c.closed {
		log.Warn("DatagramConn was garbage collected without being closed - cleaning up resources")
		c.closed = true
		if c.reader != nil {
			c.reader.Close()
		}
	}
	c.mu.Unlock()
}

// setFinalizer sets up automatic cleanup for the connection to prevent resource leaks
func (c *DatagramConn) setFinalizer() {
	runtime.SetFinalizer(c, datagramConnFinalizer)
}

// clearFinalizer removes the finalizer when Close() is called explicitly
func (c *DatagramConn) clearFinalizer() {
	runtime.SetFinalizer(c, nil)
}
