package datagram3

import (
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
)

// ReadFrom reads a datagram from the connection.
//
// This method implements the net.PacketConn interface. It starts the receive loop if not
// already started and blocks until a datagram is received. The data is copied to the provided
// buffer p, and the source address is returned as a Datagram3Addr.
//
// The source address contains the 32-byte hash (not full destination). Applications must
// resolve the hash via ResolveSource() to reply.
func (c *Datagram3Conn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
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

	// Create address with hash
	// Applications can check addr.(*Datagram3Addr).hash for the hash
	addr = &Datagram3Addr{
		addr: datagram.Source,     // May be empty if not resolved
		hash: datagram.SourceHash, // 32-byte hash
	}

	return n, addr, nil
}

// WriteTo writes a datagram to the specified address.
// This method implements the net.PacketConn interface. The address must be a Datagram3Addr
// or i2pkeys.I2PAddr containing a valid I2P destination. The entire byte slice p is sent
// as a single datagram message.
//
// If the address is a Datagram3Addr with only a hash (not resolved), the hash will be
// resolved automatically before sending.
func (c *Datagram3Conn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, oops.Errorf("connection is closed")
	}
	c.mu.RUnlock()

	// Convert address to I2P address
	var i2pAddr i2pkeys.I2PAddr

	switch a := addr.(type) {
	case *Datagram3Addr:
		// If address has full destination, use it
		if a.addr != "" {
			i2pAddr = a.addr
		} else if len(a.hash) == 32 {
			// Only hash available - resolve it
			log.Debug("Resolving hash for WriteTo")
			resolved, err := c.session.resolver.ResolveHash(a.hash)
			if err != nil {
				return 0, oops.Errorf("failed to resolve hash: %w", err)
			}
			i2pAddr = resolved
		} else {
			return 0, oops.Errorf("address has neither full destination nor valid hash")
		}
	case *i2pkeys.I2PAddr:
		i2pAddr = *a
	case i2pkeys.I2PAddr:
		i2pAddr = a
	default:
		return 0, oops.Errorf("address must be Datagram3Addr or I2PAddr")
	}

	err = c.writer.SendDatagram(p, i2pAddr)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close closes the datagram3 connection and releases associated resources.
// This method implements the net.Conn interface. It closes the reader and writer
// but does not close the underlying session, which may be shared by other connections.
// Multiple calls to Close are safe and will return nil after the first call.
func (c *Datagram3Conn) Close() error {
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
	logger := log.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"style":      "DATAGRAM3",
	})
	logger.Debug("Closing Datagram3Conn")

	c.closed = true

	// Clear the finalizer since we're cleaning up explicitly
	c.clearCleanup()

	// Close reader and writer - these are owned by this connection
	if c.reader != nil {
		c.reader.Close()
	}

	// DO NOT close the session - it's a shared resource that may be used by other connections
	// The session should be closed by the code that created it, not by individual connections
	// that use it. This follows the principle that the creator owns the resource.

	logger.Debug("Successfully closed Datagram3Conn")
	return nil
}

// LocalAddr returns the local network address as a Datagram3Addr containing
// the I2P destination address of this connection's session. This method implements
// the net.Conn interface and provides access to the local I2P destination.
func (c *Datagram3Conn) LocalAddr() net.Addr {
	return &Datagram3Addr{addr: c.session.Addr()}
}

// SetDeadline sets both read and write deadlines for the connection.
// This method implements the net.Conn interface by calling both SetReadDeadline
// and SetWriteDeadline with the same time value. If either deadline cannot be set,
// the first error encountered is returned.
func (c *Datagram3Conn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future ReadFrom calls.
// This method implements the net.Conn interface. For datagram3 connections,
// this is currently a placeholder implementation that always returns nil.
// Timeout handling is managed differently for datagram operations.
func (c *Datagram3Conn) SetReadDeadline(t time.Time) error {
	// For datagrams, we handle timeouts differently
	// This is a placeholder implementation
	return nil
}

// SetWriteDeadline sets the deadline for future WriteTo calls.
// This method implements the net.Conn interface. If the deadline is not zero,
// it calculates the timeout duration and sets it on the writer for subsequent
// write operations.
func (c *Datagram3Conn) SetWriteDeadline(t time.Time) error {
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
// Write calls.
//
// Note: This is not typical for datagrams which are connectionless,
// but provides compatibility with the net.Conn interface.
func (c *Datagram3Conn) Read(b []byte) (n int, err error) {
	n, addr, err := c.ReadFrom(b)
	if err != nil {
		return n, err
	}

	// Store remote address for Write operations
	if dg3Addr, ok := addr.(*Datagram3Addr); ok {
		i2pAddr := dg3Addr.addr
		c.remoteAddr = &i2pAddr
	}

	return n, err
}

// RemoteAddr returns the remote network address of the connection.
// This method implements the net.Conn interface. For datagram3 connections,
// this returns the address of the last peer that sent data (set by Read),
// or nil if no data has been received yet.
func (c *Datagram3Conn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return &Datagram3Addr{addr: *c.remoteAddr}
	}
	return nil
}

// Write implements net.Conn by wrapping WriteTo for stream-like usage.
// It writes data to the remote address set by the last Read operation and
// returns the number of bytes written. If no remote address has been set,
// it returns an error. Note: This is not typical for datagrams which are
// connectionless, but provides compatibility with the net.Conn interface.
func (c *Datagram3Conn) Write(b []byte) (n int, err error) {
	if c.remoteAddr == nil {
		return 0, oops.Errorf("no remote address set, use WriteTo or Read first")
	}

	addr := &Datagram3Addr{addr: *c.remoteAddr}
	return c.WriteTo(b, addr)
}

// cleanupDatagram3Conn is called by AddCleanup to ensure resources are cleaned up
// even if the user forgets to call Close(). This prevents goroutine leaks.
func cleanupDatagram3Conn(c *Datagram3Conn) {
	c.mu.Lock()
	if !c.closed {
		log.Warn("Datagram3Conn was garbage collected without being closed - cleaning up resources")
		c.closed = true
		if c.reader != nil {
			c.reader.Close()
		}
	}
	c.mu.Unlock()
}

// addCleanup sets up automatic cleanup for the connection to prevent resource leaks
func (c *Datagram3Conn) addCleanup() {
	c.cleanup = runtime.AddCleanup(c, func(c *Datagram3Conn) {
		cleanupDatagram3Conn(c)
	}, c)
}

// clearCleanup removes the automatic cleanup if Close() is called explicitly
func (c *Datagram3Conn) clearCleanup() {
	c.cleanup.Stop()
}

// PacketConn returns a net.PacketConn interface for this datagram3 session.
// This method provides compatibility with standard Go networking code by wrapping
// the datagram3 session in a PacketConn interface. The returned connection manages
// its own reader and writer and implements all standard net.PacketConn methods.
//
// The connection is automatically cleaned up by a finalizer if Close() is not called,
// but explicit Close() calls are strongly recommended to prevent resource leaks.
//
// Example usage:
//
//	conn := session.PacketConn()
//	defer conn.Close()
//
//	// Receive source
//	n, addr, err := conn.ReadFrom(buffer)
//
//	// Send reply
//	n, err = conn.WriteTo(reply, addr)
func (s *Datagram3Session) PacketConn() net.PacketConn {
	conn := &Datagram3Conn{
		session: s,
		reader:  s.NewReader(),
		writer:  s.NewWriter(),
		mu:      sync.RWMutex{},
		closed:  false,
	}

	// Add cleanup to prevent resource leaks if user forgets to call Close()
	conn.addCleanup()

	return conn
}
