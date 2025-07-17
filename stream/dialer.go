package stream

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Dial establishes a connection to the specified destination using the default context.
// This method resolves the destination string and establishes a streaming connection
// using the dialer's configured timeout. It provides a simple interface for connection
// establishment without requiring explicit context management.
// Example usage: conn, err := dialer.Dial("destination.b32.i2p")
func (d *StreamDialer) Dial(destination string) (*StreamConn, error) {
	return d.DialContext(context.Background(), destination)
}

// DialI2P establishes a connection to the specified I2P address using native addressing.
// This method accepts an i2pkeys.I2PAddr directly, bypassing the need for destination
// resolution. It uses the dialer's configured timeout and provides efficient connection
// establishment for known I2P addresses.
// Example usage: conn, err := dialer.DialI2P(addr)
func (d *StreamDialer) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error) {
	return d.DialI2PContext(context.Background(), addr)
}

// DialContext establishes a connection with context support for cancellation and timeout.
// This method resolves the destination string and establishes a streaming connection
// with context-based cancellation support. The context can override the dialer's
// default timeout and provides fine-grained control over connection establishment.
// Example usage: conn, err := dialer.DialContext(ctx, "destination.b32.i2p")
func (d *StreamDialer) DialContext(ctx context.Context, destination string) (*StreamConn, error) {
	// First resolve the destination
	addr, err := d.session.sam.Lookup(destination)
	if err != nil {
		return nil, oops.Errorf("failed to resolve destination %s: %w", destination, err)
	}

	return d.DialI2PContext(ctx, addr)
}

// DialI2PContext establishes a connection to an I2P address with context support.
// This method provides the core dialing functionality with context-based cancellation
// and timeout support. It handles SAM protocol communication, connection establishment,
// and proper resource management for streaming connections over I2P.
// Example usage: conn, err := dialer.DialI2PContext(ctx, addr)
func (d *StreamDialer) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (*StreamConn, error) {
	// Check session state before starting dial operation
	d.session.mu.RLock()
	if d.session.closed {
		d.session.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}
	d.session.mu.RUnlock()

	logger := log.WithFields(logrus.Fields{
		"session_id":  d.session.ID(),
		"destination": addr.Base32(),
	})
	logger.Debug("Dialing I2P destination")

	// Create a new SAM connection for this dial to avoid interference with session
	sam, err := common.NewSAM(d.session.sam.Sam())
	if err != nil {
		logger.WithError(err).Error("Failed to create SAM connection")
		return nil, oops.Errorf("failed to create SAM connection: %w", err)
	}

	// Set up timeout if specified, preserving existing context
	var cancel context.CancelFunc
	if d.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, d.timeout)
		defer cancel()
	}

	// Perform the dial with timeout using goroutine and channels for cancellation
	connChan := make(chan *StreamConn, 1)
	errChan := make(chan error, 1)

	go func() {
		conn, err := d.performDial(sam, addr)
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	// Handle results with proper timeout and cancellation support
	select {
	case conn := <-connChan:
		logger.Debug("Successfully established connection")
		return conn, nil
	case err := <-errChan:
		sam.Close()
		logger.WithError(err).Error("Failed to establish connection")
		return nil, err
	case <-ctx.Done():
		sam.Close()
		logger.Error("Connection attempt timed out")
		return nil, oops.Errorf("connection attempt timed out: %w", ctx.Err())
	}
}

// performDial handles the actual SAM protocol for establishing connections
func (d *StreamDialer) performDial(sam *common.SAM, addr i2pkeys.I2PAddr) (*StreamConn, error) {
	logger := log.WithFields(logrus.Fields{
		"session_id":  d.session.ID(),
		"destination": addr.Base32(),
	})

	// Send STREAM CONNECT command
	connectCmd := fmt.Sprintf("STREAM CONNECT ID=%s DESTINATION=%s SILENT=false\n",
		d.session.ID(), addr.Base64())

	logger.WithField("command", strings.TrimSpace(connectCmd)).Debug("Sending STREAM CONNECT")

	_, err := sam.Write([]byte(connectCmd))
	if err != nil {
		return nil, oops.Errorf("failed to send STREAM CONNECT: %w", err)
	}

	// Read the response
	buf := make([]byte, 4096)
	n, err := sam.Read(buf)
	if err != nil {
		return nil, oops.Errorf("failed to read STREAM CONNECT response: %w", err)
	}

	response := string(buf[:n])
	logger.WithField("response", response).Debug("Received STREAM CONNECT response")

	// Parse the response
	if err := d.parseConnectResponse(response); err != nil {
		return nil, err
	}

	// Create the StreamConn
	conn := &StreamConn{
		session: d.session,
		conn:    sam,
		laddr:   d.session.Addr(),
		raddr:   addr,
	}

	return conn, nil
}

// parseConnectResponse parses the STREAM STATUS response
func (d *StreamDialer) parseConnectResponse(response string) error {
	scanner := bufio.NewScanner(strings.NewReader(response))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		switch word {
		case "STREAM", "STATUS":
			continue
		case "RESULT=OK":
			return nil
		case "RESULT=CANT_REACH_PEER":
			return oops.Errorf("cannot reach peer")
		case "RESULT=I2P_ERROR":
			return oops.Errorf("I2P internal error")
		case "RESULT=INVALID_KEY":
			return oops.Errorf("invalid destination key")
		case "RESULT=INVALID_ID":
			return oops.Errorf("invalid session ID")
		case "RESULT=TIMEOUT":
			return oops.Errorf("connection timeout")
		default:
			if strings.HasPrefix(word, "RESULT=") {
				return oops.Errorf("connection failed: %s", word[7:])
			}
		}
	}

	return oops.Errorf("unexpected response format: %s", response)
}
