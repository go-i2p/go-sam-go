package datagram

import (
	"net"
	"sync"

	"github.com/samber/oops"
)

// DatagramListener implements net.DatagramListener for I2P datagram connections
type DatagramListener struct {
	session    *DatagramSession
	reader     *DatagramReader
	acceptChan chan net.Conn
	errorChan  chan error
	closeChan  chan struct{}
	closed     bool
	mu         sync.RWMutex
}

// Accept waits for and returns the next packet connection to the listener
func (l *DatagramListener) Accept() (net.Conn, error) {
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

// Close closes the packet listener
func (l *DatagramListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Closing PacketListener")

	l.closed = true
	close(l.closeChan)

	// Close the reader
	if l.reader != nil {
		l.reader.Close()
	}

	logger.Debug("Successfully closed PacketListener")
	return nil
}

// Addr returns the listener's network address
func (l *DatagramListener) Addr() net.Addr {
	return &DatagramAddr{addr: l.session.Addr()}
}

// acceptLoop continuously accepts incoming packet connections
func (l *DatagramListener) acceptLoop() {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Starting packet accept loop")

	for {
		select {
		case <-l.closeChan:
			logger.Debug("Packet accept loop terminated - listener closed")
			return
		default:
			conn, err := l.acceptPacketConnection()
			if err != nil {
				l.mu.RLock()
				closed := l.closed
				l.mu.RUnlock()

				if !closed {
					logger.WithError(err).Error("Failed to accept packet connection")
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
				logger.Debug("Successfully accepted new packet connection")
			case <-l.closeChan:
				conn.Close()
				return
			}
		}
	}
}

// acceptPacketConnection creates a new packet connection for incoming datagrams
func (l *DatagramListener) acceptPacketConnection() (net.Conn, error) {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Creating new packet connection")

	// For datagram sessions, we create a new DatagramConn that shares the session
	// but has its own reader/writer for handling the specific connection
	conn := &DatagramConn{
		session: l.session,
		reader:  l.session.NewReader(),
		writer:  l.session.NewWriter(),
	}

	// Start the reader loop for this connection
	go conn.reader.receiveLoop()

	return conn, nil
}
