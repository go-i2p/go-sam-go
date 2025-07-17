package raw

import (
	"net"

	"github.com/samber/oops"
)

// Accept waits for and returns the next raw connection to the listener
func (l *RawListener) Accept() (net.Conn, error) {
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

// Close closes the raw listener
func (l *RawListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return oops.Errorf("listener is already closed")
	}

	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Closing RawListener")

	l.closed = true
	close(l.closeChan)

	// Close the reader
	if l.reader != nil {
		l.reader.Close()
	}

	logger.Debug("Successfully closed RawListener")
	return nil
}

// Addr returns the listener's network address
func (l *RawListener) Addr() net.Addr {
	return &RawAddr{addr: l.session.Addr()}
}

// acceptLoop continuously accepts incoming raw connections
func (l *RawListener) acceptLoop() {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Starting raw accept loop")

	for {
		select {
		case <-l.closeChan:
			logger.Debug("Raw accept loop terminated - listener closed")
			return
		default:
			conn, err := l.acceptRawConnection()
			if err != nil {
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

			select {
			case l.acceptChan <- conn:
				logger.Debug("Successfully accepted new raw connection")
			case <-l.closeChan:
				conn.Close()
				return
			}
		}
	}
}

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
