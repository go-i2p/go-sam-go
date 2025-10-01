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
	if err := d.validateSessionState(); err != nil {
		return nil, err
	}

	d.logDialAttempt(addr)

	sam, err := d.createSAMConnection()
	if err != nil {
		return nil, err
	}

	ctx, cancel := d.setupTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	return d.performAsyncDial(ctx, sam, addr)
}

// validateSessionState checks if the session is valid and ready for dialing.
func (d *StreamDialer) validateSessionState() error {
	d.session.mu.RLock()
	defer d.session.mu.RUnlock()

	if d.session.closed {
		return oops.Errorf("session is closed")
	}
	return nil
}

// logDialAttempt logs the dial attempt with appropriate context fields.
func (d *StreamDialer) logDialAttempt(addr i2pkeys.I2PAddr) {
	log.WithFields(logrus.Fields{
		"session_id":  d.session.ID(),
		"destination": addr.Base32(),
	}).Debug("Dialing I2P destination")
}

// createSAMConnection creates a new SAM connection for the dial operation.
func (d *StreamDialer) createSAMConnection() (*common.SAM, error) {
	sam, err := common.NewSAM(d.session.sam.Sam())
	if err != nil {
		log.WithError(err).Error("Failed to create SAM connection")
		return nil, oops.Errorf("failed to create SAM connection: %w", err)
	}
	return sam, nil
}

// setupTimeout configures context timeout if specified in the dialer.
func (d *StreamDialer) setupTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if d.timeout > 0 {
		return context.WithTimeout(ctx, d.timeout)
	}
	return ctx, nil
}

// performAsyncDial executes the dial operation asynchronously with proper cancellation support.
func (d *StreamDialer) performAsyncDial(ctx context.Context, sam *common.SAM, addr i2pkeys.I2PAddr) (*StreamConn, error) {
	connChan, errChan, doneChan := d.setupDialChannels()
	
	go d.executeDialInBackground(ctx, sam, addr, connChan, errChan, doneChan)

	return d.handleDialResultWithCoordination(ctx, sam, connChan, errChan, doneChan)
}

// setupDialChannels creates the communication channels for async dialing coordination.
func (d *StreamDialer) setupDialChannels() (chan *StreamConn, chan error, chan struct{}) {
	connChan := make(chan *StreamConn, 1)
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})
	return connChan, errChan, doneChan
}

// executeDialInBackground performs the actual dial operation in a separate goroutine.
func (d *StreamDialer) executeDialInBackground(ctx context.Context, sam *common.SAM, addr i2pkeys.I2PAddr, connChan chan *StreamConn, errChan chan error, doneChan chan struct{}) {
	defer close(doneChan)
	
	conn, err := d.performDial(sam, addr)
	if err != nil {
		d.sendDialError(ctx, err, errChan)
		return
	}
	
	d.sendDialSuccess(ctx, conn, connChan)
}

// sendDialError safely sends error result checking for context cancellation.
func (d *StreamDialer) sendDialError(ctx context.Context, err error, errChan chan error) {
	select {
	case errChan <- err:
	case <-ctx.Done():
		// Context cancelled, clean up and exit
	}
}

// sendDialSuccess safely sends successful connection checking for context cancellation.
func (d *StreamDialer) sendDialSuccess(ctx context.Context, conn *StreamConn, connChan chan *StreamConn) {
	select {
	case connChan <- conn:
	case <-ctx.Done():
		// Context cancelled, close connection and exit
		if conn != nil {
			conn.Close()
		}
	}
}

// handleDialResult manages the result of the dial operation with timeout and cancellation support.
func (d *StreamDialer) handleDialResult(ctx context.Context, sam *common.SAM, connChan chan *StreamConn, errChan chan error) (*StreamConn, error) {
	select {
	case conn := <-connChan:
		log.Debug("Successfully established connection")
		return conn, nil
	case err := <-errChan:
		sam.Close()
		log.WithError(err).Error("Failed to establish connection")
		return nil, err
	case <-ctx.Done():
		sam.Close()
		log.Error("Connection attempt timed out")
		return nil, oops.Errorf("connection attempt timed out: %w", ctx.Err())
	}
}

// handleDialResultWithCoordination manages the result of the dial operation with proper goroutine coordination.
// This method ensures that the dial goroutine completes or is properly cancelled before returning.
func (d *StreamDialer) handleDialResultWithCoordination(ctx context.Context, sam *common.SAM, connChan chan *StreamConn, errChan chan error, doneChan chan struct{}) (*StreamConn, error) {
	select {
	case conn := <-connChan:
		log.Debug("Successfully established connection")
		// Wait for goroutine to complete cleanup
		<-doneChan
		return conn, nil
	case err := <-errChan:
		sam.Close()
		log.WithError(err).Error("Failed to establish connection")
		// Wait for goroutine to complete cleanup
		<-doneChan
		return nil, err
	case <-ctx.Done():
		sam.Close()
		log.Error("Connection attempt timed out")
		// Wait for goroutine to complete cleanup to prevent goroutine leak
		<-doneChan
		return nil, oops.Errorf("connection attempt timed out: %w", ctx.Err())
	}
}

// performDial handles the actual SAM protocol for establishing connections
func (d *StreamDialer) performDial(sam *common.SAM, addr i2pkeys.I2PAddr) (*StreamConn, error) {
	if err := d.sendStreamConnectCommand(sam, addr); err != nil {
		return nil, err
	}

	response, err := d.readStreamConnectResponse(sam)
	if err != nil {
		return nil, err
	}

	if err := d.parseConnectResponse(response); err != nil {
		return nil, err
	}

	return d.createStreamConnection(sam, addr), nil
}

// sendStreamConnectCommand sends the STREAM CONNECT command to the SAM bridge.
func (d *StreamDialer) sendStreamConnectCommand(sam *common.SAM, addr i2pkeys.I2PAddr) error {
	connectCmd := fmt.Sprintf("STREAM CONNECT ID=%s DESTINATION=%s SILENT=false\n",
		d.session.ID(), addr.Base64())

	log.WithFields(logrus.Fields{
		"session_id":  d.session.ID(),
		"destination": addr.Base32(),
		"command":     strings.TrimSpace(connectCmd),
	}).Debug("Sending STREAM CONNECT")

	_, err := sam.Write([]byte(connectCmd))
	if err != nil {
		return oops.Errorf("failed to send STREAM CONNECT: %w", err)
	}
	return nil
}

// readStreamConnectResponse reads and logs the response from the SAM bridge.
func (d *StreamDialer) readStreamConnectResponse(sam *common.SAM) (string, error) {
	buf := make([]byte, 4096)
	n, err := sam.Read(buf)
	if err != nil {
		return "", oops.Errorf("failed to read STREAM CONNECT response: %w", err)
	}

	response := string(buf[:n])
	log.WithFields(logrus.Fields{
		"session_id": d.session.ID(),
		"response":   response,
	}).Debug("Received STREAM CONNECT response")
	return response, nil
}

// createStreamConnection creates a new StreamConn instance with the established connection.
func (d *StreamDialer) createStreamConnection(sam *common.SAM, addr i2pkeys.I2PAddr) *StreamConn {
	return &StreamConn{
		session: d.session,
		conn:    sam,
		laddr:   d.session.Addr(),
		raddr:   addr,
	}
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
