package stream

import (
	"bufio"
	"net"
	"strings"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
)

// Accept waits for and returns the next connection to the listener
func (l *StreamListener) Accept() (net.Conn, error) {
	return l.AcceptStream()
}

// AcceptStream waits for and returns the next I2P streaming connection
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

// Close closes the listener
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

	logger.Debug("Successfully closed StreamListener")
	return nil
}

// Addr returns the listener's network address
func (l *StreamListener) Addr() net.Addr {
	return &i2pAddr{addr: l.session.Addr()}
}

// acceptLoop continuously accepts incoming connections
func (l *StreamListener) acceptLoop() {
	logger := log.WithField("session_id", l.session.ID())
	logger.Debug("Starting accept loop")

	for {
		select {
		case <-l.closeChan:
			logger.Debug("Accept loop terminated - listener closed")
			return
		default:
			conn, err := l.acceptConnection()
			if err != nil {
				l.mu.RLock()
				closed := l.closed
				l.mu.RUnlock()

				if !closed {
					logger.WithError(err).Error("Failed to accept connection")
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
				logger.Debug("Successfully accepted new connection")
			case <-l.closeChan:
				conn.Close()
				return
			}
		}
	}
}

// acceptConnection handles the low-level connection acceptance
func (l *StreamListener) acceptConnection() (*StreamConn, error) {
	logger := log.WithField("session_id", l.session.ID())

	// Read from the session connection for incoming connection requests
	buf := make([]byte, 4096)
	n, err := l.session.Read(buf)
	if err != nil {
		return nil, oops.Errorf("failed to read from session: %w", err)
	}

	response := string(buf[:n])
	logger.WithField("response", response).Debug("Received connection request")

	// Parse the STREAM STATUS response
	scanner := bufio.NewScanner(strings.NewReader(response))
	scanner.Split(bufio.ScanWords)

	var status, dest string
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

	if status != "OK" {
		return nil, oops.Errorf("connection failed with status: %s", status)
	}

	if dest == "" {
		return nil, oops.Errorf("no destination in connection request")
	}

	// Parse the remote destination
	remoteAddr, err := i2pkeys.NewI2PAddrFromString(dest)
	if err != nil {
		return nil, oops.Errorf("failed to parse remote address: %w", err)
	}

	// Create a new connection object
	streamConn := &StreamConn{
		session: l.session,
		conn:    l.session.BaseSession, // Use the session connection
		laddr:   l.session.Addr(),
		raddr:   remoteAddr,
	}

	return streamConn, nil
}

// i2pAddr implements net.Addr for I2P addresses
type i2pAddr struct {
	addr i2pkeys.I2PAddr
}

func (a *i2pAddr) Network() string {
	return "i2p"
}

func (a *i2pAddr) String() string {
	return a.addr.Base32()
}
