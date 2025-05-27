package stream

import (
	"context"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewStreamSession creates a new streaming session
func NewStreamSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	logger := log.WithFields(logrus.Fields{
		"id":      id,
		"options": options,
	})
	logger.Debug("Creating new StreamSession")

	// Create the base session using the common package
	session, err := sam.NewGenericSession("STREAM", id, keys, options)
	if err != nil {
		logger.WithError(err).Error("Failed to create generic session")
		return nil, oops.Errorf("failed to create stream session: %w", err)
	}

	baseSession, ok := session.(*common.BaseSession)
	if !ok {
		logger.Error("Session is not a BaseSession")
		session.Close()
		return nil, oops.Errorf("invalid session type")
	}

	ss := &StreamSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     options,
	}

	logger.Debug("Successfully created StreamSession")
	return ss, nil
}

// Listen creates a StreamListener that accepts incoming connections
func (s *StreamSession) Listen() (*StreamListener, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, oops.Errorf("session is closed")
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Creating StreamListener")

	listener := &StreamListener{
		session:    s,
		acceptChan: make(chan *StreamConn, 10), // Buffer for incoming connections
		errorChan:  make(chan error, 1),
		closeChan:  make(chan struct{}),
	}

	// Start accepting connections in a goroutine
	go listener.acceptLoop()

	logger.Debug("Successfully created StreamListener")
	return listener, nil
}

// NewDialer creates a StreamDialer for establishing outbound connections
func (s *StreamSession) NewDialer() *StreamDialer {
	return &StreamDialer{
		session: s,
		timeout: 30 * time.Second, // Default timeout
	}
}

// SetTimeout sets the default timeout for new dialers
func (d *StreamDialer) SetTimeout(timeout time.Duration) *StreamDialer {
	d.timeout = timeout
	return d
}

// Dial establishes a connection to the specified I2P destination
func (s *StreamSession) Dial(destination string) (*StreamConn, error) {
	return s.NewDialer().Dial(destination)
}

// DialI2P establishes a connection to the specified I2P address
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error) {
	return s.NewDialer().DialI2P(addr)
}

// DialContext establishes a connection with context support
func (s *StreamSession) DialContext(ctx context.Context, destination string) (*StreamConn, error) {
	return s.NewDialer().DialContext(ctx, destination)
}

// Close closes the streaming session and all associated resources
func (s *StreamSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	logger := log.WithField("id", s.ID())
	logger.Debug("Closing StreamSession")

	s.closed = true

	// Close the base session
	if err := s.BaseSession.Close(); err != nil {
		logger.WithError(err).Error("Failed to close base session")
		return oops.Errorf("failed to close stream session: %w", err)
	}

	logger.Debug("Successfully closed StreamSession")
	return nil
}

// Addr returns the I2P address of this session
func (s *StreamSession) Addr() i2pkeys.I2PAddr {
	return s.Keys().Addr()
}
