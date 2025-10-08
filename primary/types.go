package primary

import (
	"sync"

	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/go-i2p/go-sam-go/datagram3"
	"github.com/go-i2p/go-sam-go/raw"
	"github.com/go-i2p/go-sam-go/stream"
)

// SubSession represents a generic interface for sub-sessions that can be managed
// by a primary session. All sub-session types (stream, datagram, datagram3, raw) implement
// this interface to provide unified lifecycle management and identification.
type SubSession interface {
	// ID returns the unique identifier for this sub-session
	ID() string
	// Type returns the session type ("STREAM", "DATAGRAM", "DATAGRAM3", "RAW")
	Type() string
	// Close closes the sub-session and releases its resources
	Close() error
	// Active returns whether the sub-session is currently active
	Active() bool
}

// SubSessionRegistry manages a collection of sub-sessions with thread-safe access.
// It maintains mappings between session IDs and their corresponding session instances,
// enabling efficient lookup, registration, and cleanup operations for primary sessions.
type SubSessionRegistry struct {
	mu       sync.RWMutex
	sessions map[string]SubSession
	closed   bool
}

// NewSubSessionRegistry creates a new registry for managing sub-sessions.
// It initializes the internal data structures needed for thread-safe sub-session
// management and returns a ready-to-use registry instance.
func NewSubSessionRegistry() *SubSessionRegistry {
	return &SubSessionRegistry{
		sessions: make(map[string]SubSession),
		closed:   false,
	}
}

// Register adds a sub-session to the registry with the specified ID.
// Returns an error if the registry is closed or if a session with the same ID
// already exists. This method is thread-safe and can be called concurrently.
func (r *SubSessionRegistry) Register(id string, session SubSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return &PrimarySessionError{
			Op:  "register",
			Err: "registry is closed",
		}
	}

	if _, exists := r.sessions[id]; exists {
		return &PrimarySessionError{
			Op:  "register",
			Err: "session with this ID already exists",
		}
	}

	r.sessions[id] = session
	return nil
}

// Unregister removes a sub-session from the registry by ID.
// Returns an error if the registry is closed or if no session with the specified
// ID exists. This method is thread-safe and can be called concurrently.
func (r *SubSessionRegistry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return &PrimarySessionError{
			Op:  "unregister",
			Err: "registry is closed",
		}
	}

	if _, exists := r.sessions[id]; !exists {
		return &PrimarySessionError{
			Op:  "unregister",
			Err: "session with this ID does not exist",
		}
	}

	delete(r.sessions, id)
	return nil
}

// Get retrieves a sub-session by ID from the registry.
// Returns the session instance and true if found, or nil and false if not found.
// This method is thread-safe and provides read-only access to registered sessions.
func (r *SubSessionRegistry) Get(id string) (SubSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, false
	}

	session, exists := r.sessions[id]
	return session, exists
}

// List returns a copy of all currently registered sub-sessions.
// This method is thread-safe and returns a snapshot of the registry state
// that can be safely iterated without holding locks.
func (r *SubSessionRegistry) List() []SubSession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil
	}

	sessions := make([]SubSession, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Close closes the registry and all registered sub-sessions.
// This method ensures proper cleanup of all resources and marks the registry
// as closed to prevent further operations. It's safe to call multiple times.
func (r *SubSessionRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	// Close all registered sub-sessions
	for id, session := range r.sessions {
		if err := session.Close(); err != nil {
			log.WithError(err).WithField("id", id).Error("Failed to close sub-session")
		}
	}

	// Clear the registry and mark as closed
	r.sessions = nil
	r.closed = true
	return nil
}

// Count returns the number of currently registered sub-sessions.
// This method is thread-safe and provides a quick way to check registry size.
func (r *SubSessionRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return 0
	}

	return len(r.sessions)
}

// IsClosed returns whether the registry has been closed.
// This method is thread-safe and can be used to check registry state.
func (r *SubSessionRegistry) IsClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}

// PrimarySessionError represents errors specific to primary session operations.
// It provides structured error information with operation context for debugging
// and error handling in primary session management scenarios.
type PrimarySessionError struct {
	Op  string // The operation that caused the error
	Err string // The error description
}

// Error implements the error interface for PrimarySessionError.
// It provides a formatted error message that includes both the operation
// context and the specific error description for clear error reporting.
func (e *PrimarySessionError) Error() string {
	return "primary session " + e.Op + ": " + e.Err
}

// StreamSubSession wraps a stream.StreamSession to implement the SubSession interface.
// This adapter allows StreamSession instances to be managed by primary sessions
// while maintaining their full functionality and thread-safe operations.
type StreamSubSession struct {
	*stream.StreamSession
	id     string
	active bool
	mu     sync.RWMutex
}

// NewStreamSubSession creates a StreamSubSession wrapper around a StreamSession.
// This constructor initializes the wrapper with proper identification and state
// management to enable primary session integration.
func NewStreamSubSession(id string, session *stream.StreamSession) *StreamSubSession {
	return &StreamSubSession{
		StreamSession: session,
		id:            id,
		active:        true,
	}
}

// ID returns the unique identifier for this stream sub-session.
func (s *StreamSubSession) ID() string {
	return s.id
}

// Type returns the session type identifier for stream sessions.
func (s *StreamSubSession) Type() string {
	return "STREAM"
}

// Active returns whether this stream sub-session is currently active.
func (s *StreamSubSession) Active() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// Close closes the stream sub-session and marks it as inactive.
func (s *StreamSubSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return nil
	}

	s.active = false
	return s.StreamSession.Close()
}

// DatagramSubSession wraps a datagram.DatagramSession to implement the SubSession interface.
// This adapter allows DatagramSession instances to be managed by primary sessions
// while maintaining their full functionality and thread-safe operations.
type DatagramSubSession struct {
	*datagram.DatagramSession
	id     string
	active bool
	mu     sync.RWMutex
}

// NewDatagramSubSession creates a DatagramSubSession wrapper around a DatagramSession.
// This constructor initializes the wrapper with proper identification and state
// management to enable primary session integration.
func NewDatagramSubSession(id string, session *datagram.DatagramSession) *DatagramSubSession {
	return &DatagramSubSession{
		DatagramSession: session,
		id:              id,
		active:          true,
	}
}

// ID returns the unique identifier for this datagram sub-session.
func (s *DatagramSubSession) ID() string {
	return s.id
}

// Type returns the session type identifier for datagram sessions.
func (s *DatagramSubSession) Type() string {
	return "DATAGRAM"
}

// Active returns whether this datagram sub-session is currently active.
func (s *DatagramSubSession) Active() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// Close closes the datagram sub-session and marks it as inactive.
func (s *DatagramSubSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return nil
	}

	s.active = false
	return s.DatagramSession.Close()
}

// RawSubSession wraps a raw.RawSession to implement the SubSession interface.
// This adapter allows RawSession instances to be managed by primary sessions
// while maintaining their full functionality and thread-safe operations.
type RawSubSession struct {
	*raw.RawSession
	id     string
	active bool
	mu     sync.RWMutex
}

// NewRawSubSession creates a RawSubSession wrapper around a RawSession.
// This constructor initializes the wrapper with proper identification and state
// management to enable primary session integration.
func NewRawSubSession(id string, session *raw.RawSession) *RawSubSession {
	return &RawSubSession{
		RawSession: session,
		id:         id,
		active:     true,
	}
}

// ID returns the unique identifier for this raw sub-session.
func (s *RawSubSession) ID() string {
	return s.id
}

// Type returns the session type identifier for raw sessions.
func (s *RawSubSession) Type() string {
	return "RAW"
}

// Active returns whether this raw sub-session is currently active.
func (s *RawSubSession) Active() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// Close closes the raw sub-session and marks it as inactive.
func (s *RawSubSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return nil
	}

	s.active = false
	return s.RawSession.Close()
}

// Datagram3SubSession wraps a datagram3.Datagram3Session to implement the SubSession interface.
// This adapter allows Datagram3Session instances to be managed by primary sessions
// while maintaining their full functionality and thread-safe operations.
//
// ⚠️  SECURITY WARNING: DATAGRAM3 sources are NOT authenticated and can be spoofed!
// ⚠️  This sub-session type uses hash-based source identification which is unauthenticated.
// ⚠️  Do not trust source addresses without additional application-level authentication.
// ⚠️  If you need authenticated sources, use DatagramSubSession (DATAGRAM) instead.
type Datagram3SubSession struct {
	*datagram3.Datagram3Session
	id     string
	active bool
	mu     sync.RWMutex
}

// NewDatagram3SubSession creates a Datagram3SubSession wrapper around a Datagram3Session.
// This constructor initializes the wrapper with proper identification and state
// management to enable primary session integration.
//
// ⚠️  SECURITY WARNING: Sources are UNAUTHENTICATED and can be spoofed!
func NewDatagram3SubSession(id string, session *datagram3.Datagram3Session) *Datagram3SubSession {
	return &Datagram3SubSession{
		Datagram3Session: session,
		id:               id,
		active:           true,
	}
}

// ID returns the unique identifier for this datagram3 sub-session.
func (s *Datagram3SubSession) ID() string {
	return s.id
}

// Type returns the session type identifier for datagram3 sessions.
// Returns "DATAGRAM3" to distinguish from authenticated DATAGRAM sessions.
func (s *Datagram3SubSession) Type() string {
	return "DATAGRAM3"
}

// Active returns whether this datagram3 sub-session is currently active.
func (s *Datagram3SubSession) Active() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// Close closes the datagram3 sub-session and marks it as inactive.
func (s *Datagram3SubSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return nil
	}

	s.active = false
	return s.Datagram3Session.Close()
}
