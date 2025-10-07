package common

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-i2p/i2pkeys"
)

// I2PConfig is a struct which manages I2P configuration options.
type I2PConfig struct {
	SamHost string
	SamPort int
	TunName string

	SamMin string
	SamMax string

	Fromport string
	Toport   string

	Style   string
	TunType string

	DestinationKeys *i2pkeys.I2PKeys

	SigType                   string
	EncryptLeaseSet           bool
	LeaseSetKey               string
	LeaseSetPrivateKey        string
	LeaseSetPrivateSigningKey string
	LeaseSetKeys              i2pkeys.I2PKeys
	InAllowZeroHop            bool
	OutAllowZeroHop           bool
	InLength                  int
	OutLength                 int
	InQuantity                int
	OutQuantity               int
	InVariance                int
	OutVariance               int
	InBackupQuantity          int
	OutBackupQuantity         int
	FastRecieve               bool
	UseCompression            bool
	MessageReliability        string
	CloseIdle                 bool
	CloseIdleTime             int
	ReduceIdle                bool
	ReduceIdleTime            int
	ReduceIdleQuantity        int
	LeaseSetEncryption        string

	// SAMv3.2+ Authentication support
	User     string // Username for authenticated SAM bridges
	Password string // Password for authenticated SAM bridges

	// Streaming Library options
	AccessListType string
	AccessList     []string
}

// SAMEmit handles SAM protocol message generation and configuration.
// It embeds I2PConfig to provide access to all tunnel and session configuration options.
type SAMEmit struct {
	I2PConfig
}

// Used for controlling I2Ps SAMv3.
type SAM struct {
	SAMEmit
	SAMResolver
	net.Conn

	// Timeout for SAM connections
	Timeout time.Duration
	// Context for control of lifecycle
	Context context.Context
}

// SAMResolver provides I2P address resolution services through SAM protocol.
// It maintains a connection to the SAM bridge for performing address lookups.
type SAMResolver struct {
	*SAM
}

// options map
type Options map[string]string

// obtain sam options as list of strings
func (opts Options) AsList() (ls []string) {
	for k, v := range opts {
		ls = append(ls, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

// Session represents a generic I2P session interface for different connection types.
// It extends net.Conn with I2P-specific functionality for session identification and key management.
// All session implementations (stream, datagram, raw) must implement this interface.
type Session interface {
	net.Conn
	ID() string
	Keys() i2pkeys.I2PKeys
	Close() error
	// Add other session methods as needed
}

// BaseSession provides the underlying SAM session functionality.
// It manages the connection to the SAM bridge and handles session lifecycle.
type BaseSession struct {
	id     string
	conn   net.Conn
	keys   i2pkeys.I2PKeys
	SAM    SAM
	mu     sync.RWMutex
	closed bool
}

// Conn returns the underlying network connection for the session.
// This provides access to the raw connection for advanced operations.
func (bs *BaseSession) Conn() net.Conn {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.conn
}

// ID returns the unique session identifier used by the SAM bridge.
// This identifier is used to distinguish between multiple sessions on the same connection.
func (bs *BaseSession) ID() string { return bs.id }

// Keys returns the I2P cryptographic keys associated with this session.
// These keys define the session's I2P destination and identity.
func (bs *BaseSession) Keys() i2pkeys.I2PKeys { return bs.keys }

// Read reads data from the session connection into the provided buffer.
// Implements the io.Reader interface for standard Go I/O operations.
func (bs *BaseSession) Read(b []byte) (int, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	if bs.closed {
		return 0, net.ErrClosed
	}
	return bs.conn.Read(b)
}

// Write writes data from the buffer to the session connection.
// Implements the io.Writer interface for standard Go I/O operations.
func (bs *BaseSession) Write(b []byte) (int, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.conn.Write(b)
}

// Close closes the session connection and releases associated resources.
// Implements the io.Closer interface for proper resource cleanup.
func (bs *BaseSession) Close() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.closed {
		return nil
	}

	// Mark the session as closed
	bs.closed = true

	// Note: We do NOT close bs.conn here because it's a shared connection
	// with the parent SAM instance. Only the SAM instance should close its connection.
	// Sessions just mark themselves as closed and clean up their own resources.

	return nil
}

// LocalAddr returns the local network address of the session connection.
// Implements the net.Conn interface for network address information.
func (bs *BaseSession) LocalAddr() net.Addr {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.conn.LocalAddr()
}

// RemoteAddr returns the remote network address of the session connection.
// Implements the net.Conn interface for network address information.
func (bs *BaseSession) RemoteAddr() net.Addr {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.conn.RemoteAddr()
}

// SetDeadline sets read and write deadlines for the session connection.
// Implements the net.Conn interface for timeout control.
func (bs *BaseSession) SetDeadline(t time.Time) error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline for the session connection.
// Implements the net.Conn interface for read timeout control.
func (bs *BaseSession) SetReadDeadline(t time.Time) error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline for the session connection.
// Implements the net.Conn interface for write timeout control.
func (bs *BaseSession) SetWriteDeadline(t time.Time) error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.conn.SetWriteDeadline(t)
}

// From returns the configured source port for the session.
// Used in port-based session configurations for service identification.
func (bs *BaseSession) From() string {
	return bs.SAM.SAMEmit.I2PConfig.Fromport
}

// To returns the configured destination port for the session.
// Used in port-based session configurations for service identification.
func (bs *BaseSession) To() string {
	return bs.SAM.SAMEmit.I2PConfig.Toport
}
