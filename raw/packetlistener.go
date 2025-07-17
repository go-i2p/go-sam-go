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

	// Close the reader to stop receiving datagrams
	if l.reader != nil {
		l.reader.Close()
	}

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
// acceptLoop continuously accepts incoming raw connections
func (l *RawListener) acceptLoop() {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Starting raw accept loop")

	// Continuously accept connections until the listener is closed
	for {
		select {
		case <-l.closeChan:
			logger.Debug("Raw accept loop terminated - listener closed")
			return
		default:
			// Try to accept a new raw connection
			conn, err := l.acceptRawConnection()
			if err != nil {
				// Check if the listener is still open before sending error
				l.mu.RLock()
				closed := l.closed
				l.mu.RUnlock()

				if !closed {
					logger.WithError(err).Error("Failed to accept raw connection")
					select {
					case l.errorChan <- err:
					case <-l.closeChan:
						return
					}
				}
				continue
			}

			// Send the new connection to the accept channel
			select {
			case l.acceptChan <- conn:
				logger.Debug("Successfully accepted new raw connection")
			case <-l.closeChan:
				// Clean up the connection if listener was closed
				conn.Close()
				return
			}
		}
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

	// Start the reader loop once for this connection
	if conn.reader != nil {
		go conn.reader.receiveLoop()
	}

	return conn, nil
}
