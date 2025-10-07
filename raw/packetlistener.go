package raw

import (
	"net"

	"github.com/samber/oops"
)

// Accept waits for and returns the next raw connection to the listener.
// This method implements the net.Listener interface and blocks until a connection
// is available or an error occurs, returning the connection or error.
// Example usage: conn, err := listener.Accept()
func (l *RawListener) Accept() (net.Conn, error) {
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return nil, oops.Errorf("listener is closed")
	}
	l.mu.RUnlock()

	// Use select to handle multiple channels atomically
	// This ensures proper handling of connections, errors, and close signals
	select {
	case conn := <-l.acceptChan:
		return conn, nil
	case err := <-l.errorChan:
		return nil, err
	case <-l.closeChan:
		return nil, oops.Errorf("listener is closed")
	}
}

// Close closes the raw listener and stops accepting new connections.
// This method is safe to call multiple times and will clean up all resources
// including the reader and associated channels.
// Example usage: defer listener.Close()
func (l *RawListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return oops.Errorf("listener is already closed")
	}

	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Closing RawListener")

	l.closed = true
	// Signal the accept loop to terminate
	close(l.closeChan)

	// Close the main reader to stop receiving datagrams
	if l.reader != nil {
		l.reader.Close()
	}

	// Close all active readers created by acceptRawConnection
	logger.WithField("active_readers", len(l.activeReaders)).Debug("Cleaning up active readers")
	for _, reader := range l.activeReaders {
		if reader != nil {
			reader.Close()
		}
	}
	// Clear the slice after cleanup
	l.activeReaders = nil

	logger.Debug("Successfully closed RawListener")
	return nil
}

// Addr returns the listener's network address.
// This method implements the net.Listener interface and returns the I2P address
// of the session wrapped in a RawAddr for compatibility with net.Addr.
// Example usage: addr := listener.Addr()
func (l *RawListener) Addr() net.Addr {
	return &RawAddr{addr: l.session.Addr()}
}

// acceptLoop continuously accepts incoming raw connections in a separate goroutine.
// This method manages the connection acceptance lifecycle, handles error conditions,
// and maintains the acceptChan buffer for incoming connections until the listener is closed.
func (l *RawListener) acceptLoop() {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Starting raw accept loop")

	for {
		select {
		case <-l.closeChan:
			logger.Debug("Raw accept loop terminated - listener closed")
			return
		default:
			if !l.processIncomingConnection() {
				return
			}
		}
	}
}

// processIncomingConnection handles a single incoming connection attempt.
// Returns false if the accept loop should terminate, true to continue.
func (l *RawListener) processIncomingConnection() bool {
	conn, err := l.acceptRawConnection()
	if err != nil {
		return l.handleConnectionError(err)
	}
	return l.dispatchConnection(conn)
}

// handleConnectionError processes errors from connection acceptance.
// Returns false if the accept loop should terminate, true to continue.
func (l *RawListener) handleConnectionError(err error) bool {
	l.mu.RLock()
	closed := l.closed
	l.mu.RUnlock()

	if !closed {
		logger := log.WithField("session_id", l.session.ID())
		logger.WithError(err).Error("Failed to accept raw connection")
		select {
		case l.errorChan <- err:
		case <-l.closeChan:
			return false
		}
	}
	return true
}

// dispatchConnection sends an accepted connection to the accept channel.
// Returns false if the accept loop should terminate, true to continue.
func (l *RawListener) dispatchConnection(conn *RawConn) bool {
	logger := log.WithField("session_id", l.session.ID())
	select {
	case l.acceptChan <- conn:
		logger.Debug("Successfully accepted new raw connection")
		return true
	case <-l.closeChan:
		conn.Close()
		return false
	default:
		// Non-blocking: if acceptChan is full, drop the connection to prevent deadlock
		logger.Warn("Accept channel full, dropping connection to prevent deadlock")
		conn.Close()
		return true
	}
}

// acceptRawConnection creates a new raw connection for handling incoming datagrams.
// For raw sessions, this method creates a RawConn that shares the session resources
// but has its own dedicated reader and writer components for handling the specific connection.
// acceptRawConnection creates a new raw connection for incoming datagrams
func (l *RawListener) acceptRawConnection() (*RawConn, error) {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Creating new raw connection")

	// For raw sessions, we create a new RawConn that shares the session
	// but has its own reader/writer for handling the specific connection
	conn := &RawConn{
		session: l.session,
		reader:  l.session.NewReader(),
		writer:  l.session.NewWriter(),
	}

	// Track the reader for proper cleanup
	if conn.reader != nil {
		l.mu.Lock()
		l.activeReaders = append(l.activeReaders, conn.reader)
		l.mu.Unlock()
		
		// Start the reader loop once for this connection
		go conn.reader.receiveLoop()
	}

	return conn, nil
}
