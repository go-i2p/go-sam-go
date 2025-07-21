package stream

import (
	"bufio"
	"net"
	"runtime"
	"strings"
	"time"

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

	// Remove the finalizer to prevent it from running on an already closed listener
	runtime.SetFinalizer(l, nil)

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

	// Remove the finalizer to prevent it from running on an already closed listener
	runtime.SetFinalizer(l, nil)

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

// acceptConnection handles the low-level connection acceptance
func (l *StreamListener) acceptConnection() (*StreamConn, error) {
	logger := log.WithField("session_id", l.session.ID())

	response, err := l.readConnectionRequest()
	if err != nil {
		return nil, err
	}

	logger.WithField("response", response).Debug("Received connection request")

	status, dest, err := l.parseConnectionResponse(response)
	if err != nil {
		return nil, err
	}

	if err := l.validateConnectionStatus(status); err != nil {
		return nil, err
	}

	if err := l.validateDestination(dest); err != nil {
		return nil, err
	}

	return l.createStreamConnection(dest)
}

// readConnectionRequest reads the incoming connection request from the session.
func (l *StreamListener) readConnectionRequest() (string, error) {
	buf := make([]byte, 4096)
	n, err := l.session.Read(buf)
	if err != nil {
		return "", oops.Errorf("failed to read from session: %w", err)
	}
	return string(buf[:n]), nil
}

// parseConnectionResponse parses the STREAM STATUS response and extracts status and destination.
func (l *StreamListener) parseConnectionResponse(response string) (status, dest string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(response))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		switch {
		case word == "STREAM":
			continue
		case word == "STATUS":
			continue
		case strings.HasPrefix(word, "RESULT="):
			status = word[7:]
		case strings.HasPrefix(word, "DESTINATION="):
			dest = word[12:]
		}
	}
	return status, dest, nil
}

// validateConnectionStatus checks if the connection status indicates success.
func (l *StreamListener) validateConnectionStatus(status string) error {
	if status != "OK" {
		return oops.Errorf("connection failed with status: %s", status)
	}
	return nil
}

// validateDestination ensures that a destination address was provided in the response.
func (l *StreamListener) validateDestination(dest string) error {
	if dest == "" {
		return oops.Errorf("no destination in connection request")
	}
	return nil
}

// createStreamConnection creates a new StreamConn from the parsed destination.
func (l *StreamListener) createStreamConnection(dest string) (*StreamConn, error) {
	remoteAddr, err := i2pkeys.NewI2PAddrFromString(dest)
	if err != nil {
		return nil, oops.Errorf("failed to parse remote address: %w", err)
	}

	streamConn := &StreamConn{
		session: l.session,
		conn:    l.session.BaseSession,
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
