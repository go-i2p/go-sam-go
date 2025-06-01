package datagram

import (
	"context"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Dial establishes a datagram connection to the specified destination
func (ds *DatagramSession) Dial(destination string) (net.PacketConn, error) {
	return ds.DialTimeout(destination, 30*time.Second)
}

// DialTimeout establishes a datagram connection with a timeout
func (ds *DatagramSession) DialTimeout(destination string, timeout time.Duration) (net.PacketConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ds.DialContext(ctx, destination)
}

// DialContext establishes a datagram connection with context support
func (ds *DatagramSession) DialContext(ctx context.Context, destination string) (net.PacketConn, error) {
	// Check if session is closed first
	ds.mu.RLock()
	if ds.closed {
		ds.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	ds.mu.RUnlock()

	logger := log.WithFields(logrus.Fields{
		"destination": destination,
	})
	logger.Debug("Dialing datagram destination")

	// Check context cancellation before proceeding
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Parse destination address
	destAddr, err := i2pkeys.NewI2PAddrFromString(destination)
	if err != nil {
		return nil, oops.Errorf("invalid destination address: %w", err)
	}

	// Create a datagram connection
	conn := &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}

	// Set remote address for the connection
	conn.remoteAddr = &destAddr

	// Start the reader loop in a goroutine with context cancellation
	go func() {
		select {
		case <-ctx.Done():
			// Context cancelled, close the reader
			conn.reader.Close()
			return
		default:
			conn.reader.receiveLoop()
		}
	}()

	logger.Debug("Successfully created datagram connection")
	return conn, nil
}

// DialI2P establishes a datagram connection to an I2P address
func (ds *DatagramSession) DialI2P(addr i2pkeys.I2PAddr) (net.PacketConn, error) {
	return ds.DialI2PTimeout(addr, 30*time.Second)
}

// DialI2PTimeout establishes a datagram connection to an I2P address with timeout
func (ds *DatagramSession) DialI2PTimeout(addr i2pkeys.I2PAddr, timeout time.Duration) (net.PacketConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ds.DialI2PContext(ctx, addr)
}

// DialI2PContext establishes a datagram connection to an I2P address with context support
func (ds *DatagramSession) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (net.PacketConn, error) {
	// Check if session is closed first
	ds.mu.RLock()
	if ds.closed {
		ds.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	ds.mu.RUnlock()

	logger := log.WithFields(logrus.Fields{
		"destination": addr.Base32(),
	})
	logger.Debug("Dialing I2P datagram destination")

	// Check context cancellation before proceeding
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Create a datagram connection
	conn := &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}

	// Set remote address for the connection
	conn.remoteAddr = &addr

	// Start the reader loop in a goroutine with context cancellation
	go func() {
		select {
		case <-ctx.Done():
			// Context cancelled, close the reader
			conn.reader.Close()
			return
		default:
			conn.reader.receiveLoop()
		}
	}()

	logger.Debug("Successfully created I2P datagram connection")
	return conn, nil
}
