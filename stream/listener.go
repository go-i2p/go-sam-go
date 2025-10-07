package stream

import (
	"bufio"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/logger"
	"github.com/samber/oops"
)

// Accept waits for and returns the next connection to the listener.
// This method implements the net.Listener interface and provides compatibility
// with standard Go networking patterns. It returns a net.Conn interface that
// can be used with any Go networking code expecting standard connections.
// Example usage: conn, err := listener.Accept()
func (l *StreamListener) Accept() (net.Conn, error) {
	return l.AcceptStream()
}

// AcceptStream waits for and returns the next I2P streaming connection.
// This method provides I2P-specific connection acceptance, returning a StreamConn
// directly rather than the generic net.Conn interface. It offers more type safety
// and I2P-specific functionality compared to the generic Accept method.
// Example usage: conn, err := listener.AcceptStream()
func (l *StreamListener) AcceptStream() (*StreamConn, error) {
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return nil, oops.Errorf("listener is closed")
	}
	l.mu.RUnlock()

	select {
	case conn := <-l.acceptChan:
		return conn, nil
	case err := <-l.errorChan:
		return nil, err
	case <-l.closeChan:
		return nil, oops.Errorf("listener is closed")
	}
}

// Close closes the listener and stops accepting new connections.
// This method implements the net.Listener interface and is safe to call multiple times.
// It properly handles concurrent access and ensures clean shutdown of the accept loop
// with appropriate resource cleanup and error handling.
// Example usage: defer listener.Close()
func (l *StreamListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Closing StreamListener")

	l.closed = true
	close(l.closeChan)
	if l.cancel != nil {
		l.cancel()
	}

	// Unregister this listener from the session
	l.session.unregisterListener(l)

	// Remove the cleanup to prevent it from running on an already closed listener
	var zero runtime.Cleanup
	if l.cleanup != zero {
		l.cleanup.Stop()
		l.cleanup = zero
	}

	logger.Debug("Successfully closed StreamListener")
	return nil
}

// closeWithoutUnregister closes the listener without unregistering from the session.
// This method is used internally by closeAllListeners to avoid deadlock.
func (l *StreamListener) closeWithoutUnregister() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Closing StreamListener without unregister")

	l.closed = true

	// Set a read deadline to unblock any pending reads
	l.session.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	close(l.closeChan)
	if l.cancel != nil {
		l.cancel()
	}

	// Remove the cleanup to prevent it from running on an already closed listener
	var zero runtime.Cleanup
	if l.cleanup != zero {
		l.cleanup.Stop()
		l.cleanup = zero
	}

	logger.Debug("Successfully closed StreamListener without unregister")
	return nil
}

// Addr returns the listener's network address.
// This method implements the net.Listener interface and provides the I2P address
// that the listener is bound to. The returned address implements the net.Addr
// interface and can be used for logging or connection management.
// Example usage: addr := listener.Addr()
func (l *StreamListener) Addr() net.Addr {
	return &i2pAddr{addr: l.session.Addr()}
}

// acceptLoop continuously accepts incoming connections
func (l *StreamListener) acceptLoop() {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Starting accept loop")

	for {
		if l.shouldTerminateLoop(logger) {
			return
		}

		conn, err := l.acceptConnection()
		if err != nil {
			if l.handleAcceptError(err, logger) {
				return
			}
			continue
		}

		if l.deliverConnection(conn, logger) {
			return
		}
	}
}

// shouldTerminateLoop checks if the accept loop should terminate due to context cancellation or close signal.
func (l *StreamListener) shouldTerminateLoop(logger *logger.Entry) bool {
	select {
	case <-l.ctx.Done():
		logger.Debug("Accept loop terminated - listener closed (context)")
		return true
	case <-l.closeChan:
		logger.Debug("Accept loop terminated - listener closed (closeChan)")
		return true
	default:
		return false
	}
}

// handleAcceptError processes connection acceptance errors and delivers them to the error channel.
// It returns true if the loop should terminate, false if it should continue.
func (l *StreamListener) handleAcceptError(err error, logger *logger.Entry) bool {
	// Check if listener is closed before reporting error to avoid race conditions
	l.mu.RLock()
	closed := l.closed
	l.mu.RUnlock()

	if !closed {
		logger.WithError(err).Error("Failed to accept connection")
		// Non-blocking error delivery with fallback to close detection
		select {
		case l.errorChan <- err:
		case <-l.ctx.Done():
			return true
		case <-l.closeChan:
			return true
		}
	}
	return false
}

// deliverConnection delivers an accepted connection to the accept channel with proper cleanup.
// It returns true if the loop should terminate, false if it should continue.
func (l *StreamListener) deliverConnection(conn *StreamConn, logger *logger.Entry) bool {
	// Non-blocking connection delivery with proper cleanup on close
	select {
	case l.acceptChan <- conn:
		logger.Debug("Successfully accepted new connection")
		return false
	case <-l.ctx.Done():
		conn.Close()
		return true
	case <-l.closeChan:
		// Close the connection if listener is shutting down
		conn.Close()
		return true
	}
}

// acceptConnection handles the low-level connection acceptance using proper SAMv3 STREAM ACCEPT sequence.
// PROTOCOL: SAMv3 Section "SAM Virtual Streams: ACCEPT"
// Each STREAM ACCEPT requires a dedicated socket to the SAM bridge.
// The bridge responds with STREAM STATUS, then sends the destination
// line when a connection arrives, followed by streaming data on this socket.
func (l *StreamListener) acceptConnection() (*StreamConn, error) {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Starting STREAM ACCEPT sequence")

	// Step 1: Create dedicated socket to SAM bridge for this ACCEPT
	sam, err := l.createAcceptSocket()
	if err != nil {
		return nil, err
	}

	// Set up cleanup - always close socket on error
	var streamConn *StreamConn
	defer func() {
		if streamConn == nil {
			sam.Close()
		}
	}()

	// Step 2: Send STREAM ACCEPT command
	if err := l.sendStreamAcceptCommand(sam, logger); err != nil {
		return nil, err
	}

	// Step 3: Read STREAM STATUS response
	if err := l.parseStreamStatusResponse(sam, logger); err != nil {
		return nil, err
	}

	// Step 4: Wait for incoming connection - SAM sends destination line
	dest, err := l.readDestinationLine(sam, logger)
	if err != nil {
		return nil, err
	}

	// Step 5: Create StreamConn using the ACCEPT socket for data transfer
	streamConn, err = l.createStreamConnectionWithSocket(dest, sam)
	if err != nil {
		return nil, err
	}

	logger.Debug("Successfully accepted connection")
	return streamConn, nil
}

// createAcceptSocket creates a dedicated SAM connection for this ACCEPT operation.
// Per SAMv3 spec: "A client waits for an incoming connection request by: opening a new socket with the SAM bridge"
func (l *StreamListener) createAcceptSocket() (*common.SAM, error) {
	// Get the SAM address from the session's SAM instance
	samAddress := l.session.sam.SAMEmit.I2PConfig.SAMAddress()
	if samAddress == "" {
		// Fallback to default if not set
		samAddress = "127.0.0.1:7656"
	}

	sam, err := common.NewSAM(samAddress)
	if err != nil {
		return nil, oops.Errorf("failed to create SAM connection for ACCEPT: %w", err)
	}

	return sam, nil
}

// sendStreamAcceptCommand sends the STREAM ACCEPT command to the SAM bridge.
func (l *StreamListener) sendStreamAcceptCommand(sam *common.SAM, logger *logger.Entry) error {
	acceptCmd := fmt.Sprintf("STREAM ACCEPT ID=%s SILENT=false\n", l.session.ID())
	logger.WithField("command", strings.TrimSpace(acceptCmd)).Debug("Sending STREAM ACCEPT")

	if _, err := sam.Write([]byte(acceptCmd)); err != nil {
		return oops.Errorf("failed to send STREAM ACCEPT: %w", err)
	}

	return nil
}

// parseStreamStatusResponse reads and parses the STREAM STATUS RESULT=OK response.
func (l *StreamListener) parseStreamStatusResponse(sam *common.SAM, logger *logger.Entry) error {
	statusBuf := make([]byte, 4096)
	n, err := sam.Read(statusBuf)
	if err != nil {
		return oops.Errorf("failed to read STREAM STATUS: %w", err)
	}

	statusResponse := string(statusBuf[:n])
	logger.WithField("response", statusResponse).Debug("Received STREAM STATUS")

	// Parse STREAM STATUS RESULT=...
	scanner := bufio.NewScanner(strings.NewReader(statusResponse))
	scanner.Split(bufio.ScanWords)

	var result string
	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "RESULT=") {
			result = word[7:]
			break
		}
	}

	if result != "OK" {
		// Extract error message if present
		if strings.Contains(statusResponse, "RESULT=I2P_ERROR") {
			return oops.Errorf("STREAM ACCEPT failed: %s", statusResponse)
		}
		return oops.Errorf("STREAM ACCEPT failed with status: %s", statusResponse)
	}

	return nil
}

// readDestinationLine waits for and reads the destination line when a connection arrives.
// Format: "$destination FROM_PORT=nnn TO_PORT=nnn\n" (SAM 3.2+) or "$destination\n" (SAM 3.0/3.1)
func (l *StreamListener) readDestinationLine(sam *common.SAM, logger *logger.Entry) (string, error) {
	destBuf := make([]byte, 4096)
	n, err := sam.Read(destBuf)
	if err != nil {
		return "", oops.Errorf("failed to read destination: %w", err)
	}

	destLine := string(destBuf[:n])
	logger.WithField("destLine", destLine).Debug("Received destination")

	// Parse destination line
	// Format: "$destination FROM_PORT=nnn TO_PORT=nnn\n" (SAM 3.2+)
	// Or just: "$destination\n" (SAM 3.0/3.1)
	parts := strings.Fields(destLine)
	if len(parts) == 0 {
		return "", oops.Errorf("empty destination line")
	}

	return parts[0], nil
}

// createStreamConnectionWithSocket creates a new StreamConn using the provided accept socket.
func (l *StreamListener) createStreamConnectionWithSocket(dest string, sam *common.SAM) (*StreamConn, error) {
	remoteAddr, err := i2pkeys.NewI2PAddrFromString(dest)
	if err != nil {
		return nil, oops.Errorf("failed to parse remote address: %w", err)
	}

	// Create StreamConn using the accept socket, not the session socket
	streamConn := &StreamConn{
		session: l.session,
		conn:    sam, // Use the accept socket as data socket
		laddr:   l.session.Addr(),
		raddr:   remoteAddr,
	}

	return streamConn, nil
}

// i2pAddr implements net.Addr for I2P addresses, providing compatibility with standard Go networking.
// This internal type wraps an i2pkeys.I2PAddr to provide the net.Addr interface methods,
// enabling seamless integration with existing Go networking code that expects net.Addr.
// It provides string representation and network type identification for I2P addresses.
type i2pAddr struct {
	addr i2pkeys.I2PAddr
}

// Network returns the network type for this address.
// This method implements the net.Addr interface and always returns "i2p"
// to identify this as an I2P network address type for routing and logging purposes.
func (a *i2pAddr) Network() string {
	return "i2p"
}

// String returns the string representation of the I2P address.
// This method implements the net.Addr interface and returns the Base32 encoded
// representation of the I2P address for human-readable display and logging.
func (a *i2pAddr) String() string {
	return a.addr.Base32()
}
