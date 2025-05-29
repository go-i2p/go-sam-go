package datagram

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
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
	logger := log.WithFields(logrus.Fields{
		"destination": destination,
	})
	logger.Debug("Dialing datagram destination")

	// Create a datagram connection
	conn := &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}

	// Start the reader loop
	go conn.reader.receiveLoop()

	logger.WithField("session_id", ds.ID()).Debug("Successfully created datagram connection")
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
	logger := log.WithFields(logrus.Fields{
		"destination": addr.Base32(),
	})
	logger.Debug("Dialing I2P datagram destination")

	// Create a datagram connection
	conn := &DatagramConn{
		session: ds,
		reader:  ds.NewReader(),
		writer:  ds.NewWriter(),
	}

	// Start the reader loop
	go conn.reader.receiveLoop()

	logger.WithField("session_id", ds.ID()).Debug("Successfully created I2P datagram connection")
	return conn, nil
}

// generateSessionID generates a unique session identifier
func generateSessionID() string {
	return fmt.Sprintf("datagram_%d", time.Now().UnixNano())
}
