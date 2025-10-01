package primary

import (
	"testing"

	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/go-i2p/go-sam-go/raw"
	"github.com/go-i2p/go-sam-go/stream"
)

func TestPrimarySessionError(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		err      string
		expected string
	}{
		{
			name:     "register error",
			op:       "register",
			err:      "session already exists",
			expected: "primary session register: session already exists",
		},
		{
			name:     "unregister error",
			op:       "unregister",
			err:      "session not found",
			expected: "primary session unregister: session not found",
		},
		{
			name:     "close error",
			op:       "close",
			err:      "connection failed",
			expected: "primary session close: connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &PrimarySessionError{
				Op:  tt.op,
				Err: tt.err,
			}

			if err.Error() != tt.expected {
				t.Errorf("Error() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestSubSessionWrappers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test SAM connection and keys
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	t.Run("StreamSubSession wrapper", func(t *testing.T) {
		// Create a real stream session
		streamSession, err := stream.NewStreamSession(sam, "test_stream_wrapper", keys, []string{"inbound.length=1"})
		if err != nil {
			t.Fatalf("Failed to create stream session: %v", err)
		}
		defer streamSession.Close()

		// Wrap it in a sub-session
		subSession := NewStreamSubSession("wrapped_stream", streamSession)

		// Test interface compliance
		if subSession.ID() != "wrapped_stream" {
			t.Errorf("ID() = %v, want wrapped_stream", subSession.ID())
		}

		if subSession.Type() != "STREAM" {
			t.Errorf("Type() = %v, want STREAM", subSession.Type())
		}

		if !subSession.Active() {
			t.Error("Active() = false, want true")
		}

		// Test close
		err = subSession.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		if subSession.Active() {
			t.Error("Active() = true after close, want false")
		}

		// Test double close (should be safe)
		err = subSession.Close()
		if err != nil {
			t.Errorf("Second Close() error = %v", err)
		}
	})

	t.Run("DatagramSubSession wrapper", func(t *testing.T) {
		// Create a real datagram session
		datagramSession, err := datagram.NewDatagramSession(sam, "test_datagram_wrapper", keys, []string{"inbound.length=1"})
		if err != nil {
			t.Fatalf("Failed to create datagram session: %v", err)
		}
		defer datagramSession.Close()

		// Wrap it in a sub-session
		subSession := NewDatagramSubSession("wrapped_datagram", datagramSession)

		// Test interface compliance
		if subSession.ID() != "wrapped_datagram" {
			t.Errorf("ID() = %v, want wrapped_datagram", subSession.ID())
		}

		if subSession.Type() != "DATAGRAM" {
			t.Errorf("Type() = %v, want DATAGRAM", subSession.Type())
		}

		if !subSession.Active() {
			t.Error("Active() = false, want true")
		}

		// Test close
		err = subSession.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		if subSession.Active() {
			t.Error("Active() = true after close, want false")
		}
	})

	t.Run("RawSubSession wrapper", func(t *testing.T) {
		// Create a real raw session
		rawSession, err := raw.NewRawSession(sam, "test_raw_wrapper", keys, []string{"inbound.length=1"})
		if err != nil {
			t.Fatalf("Failed to create raw session: %v", err)
		}
		defer rawSession.Close()

		// Wrap it in a sub-session
		subSession := NewRawSubSession("wrapped_raw", rawSession)

		// Test interface compliance
		if subSession.ID() != "wrapped_raw" {
			t.Errorf("ID() = %v, want wrapped_raw", subSession.ID())
		}

		if subSession.Type() != "RAW" {
			t.Errorf("Type() = %v, want RAW", subSession.Type())
		}

		if !subSession.Active() {
			t.Error("Active() = false, want true")
		}

		// Test close
		err = subSession.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		if subSession.Active() {
			t.Error("Active() = true after close, want false")
		}
	})
}

func TestSubSessionRegistry(t *testing.T) {
	t.Run("registry lifecycle", func(t *testing.T) {
		registry := NewSubSessionRegistry()

		// Test initial state
		if registry.Count() != 0 {
			t.Errorf("Initial count = %d, want 0", registry.Count())
		}

		if registry.IsClosed() {
			t.Error("Registry should not be closed initially")
		}

		sessions := registry.List()
		if len(sessions) != 0 {
			t.Errorf("Initial list length = %d, want 0", len(sessions))
		}

		// Test operations on empty registry
		_, exists := registry.Get("nonexistent")
		if exists {
			t.Error("Get() on empty registry should return false")
		}

		err := registry.Unregister("nonexistent")
		if err == nil {
			t.Error("Unregister() on empty registry should return error")
		}

		// Close and test closed state
		err = registry.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		if !registry.IsClosed() {
			t.Error("Registry should be closed after Close()")
		}

		// Test operations on closed registry
		mockSub := &mockSubSession{id: "test", sessionType: "STREAM", active: true}
		err = registry.Register("test", mockSub)
		if err == nil {
			t.Error("Register() on closed registry should return error")
		}

		err = registry.Unregister("test")
		if err == nil {
			t.Error("Unregister() on closed registry should return error")
		}

		_, exists = registry.Get("test")
		if exists {
			t.Error("Get() on closed registry should return false")
		}

		if registry.Count() != 0 {
			t.Errorf("Count() on closed registry = %d, want 0", registry.Count())
		}

		sessions = registry.List()
		if sessions != nil {
			t.Error("List() on closed registry should return nil")
		}

		// Test double close (should be safe)
		err = registry.Close()
		if err != nil {
			t.Errorf("Second Close() error = %v", err)
		}
	})

	t.Run("registry operations", func(t *testing.T) {
		registry := NewSubSessionRegistry()
		defer registry.Close()

		mockSub1 := &mockSubSession{id: "sub1", sessionType: "STREAM", active: true}
		mockSub2 := &mockSubSession{id: "sub2", sessionType: "DATAGRAM", active: true}
		mockSub3 := &mockSubSession{id: "sub3", sessionType: "RAW", active: true}

		// Test registration
		err := registry.Register("sub1", mockSub1)
		if err != nil {
			t.Errorf("Register() error = %v", err)
		}

		err = registry.Register("sub2", mockSub2)
		if err != nil {
			t.Errorf("Register() error = %v", err)
		}

		err = registry.Register("sub3", mockSub3)
		if err != nil {
			t.Errorf("Register() error = %v", err)
		}

		// Test duplicate registration
		err = registry.Register("sub1", mockSub1)
		if err == nil {
			t.Error("Duplicate registration should return error")
		}

		// Test count
		if registry.Count() != 3 {
			t.Errorf("Count() = %d, want 3", registry.Count())
		}

		// Test get
		sub, exists := registry.Get("sub2")
		if !exists {
			t.Error("Get() should find existing session")
		}
		if sub.ID() != "sub2" {
			t.Errorf("Get() returned wrong session: %s", sub.ID())
		}

		// Test list
		sessions := registry.List()
		if len(sessions) != 3 {
			t.Errorf("List() length = %d, want 3", len(sessions))
		}

		// Verify all sessions are in the list
		sessionIDs := make(map[string]bool)
		for _, s := range sessions {
			sessionIDs[s.ID()] = true
		}

		expectedIDs := []string{"sub1", "sub2", "sub3"}
		for _, id := range expectedIDs {
			if !sessionIDs[id] {
				t.Errorf("Session %s not found in list", id)
			}
		}

		// Test unregister
		err = registry.Unregister("sub2")
		if err != nil {
			t.Errorf("Unregister() error = %v", err)
		}

		if registry.Count() != 2 {
			t.Errorf("Count() after unregister = %d, want 2", registry.Count())
		}

		_, exists = registry.Get("sub2")
		if exists {
			t.Error("Get() should not find unregistered session")
		}

		// Test unregister non-existent
		err = registry.Unregister("nonexistent")
		if err == nil {
			t.Error("Unregister() non-existent should return error")
		}
	})

	t.Run("registry close with sessions", func(t *testing.T) {
		registry := NewSubSessionRegistry()

		mockSub1 := &mockSubSession{id: "sub1", sessionType: "STREAM", active: true}
		mockSub2 := &mockSubSession{id: "sub2", sessionType: "DATAGRAM", active: true}

		registry.Register("sub1", mockSub1)
		registry.Register("sub2", mockSub2)

		// Verify sessions are active
		if !mockSub1.Active() || !mockSub2.Active() {
			t.Error("Sessions should be active before registry close")
		}

		// Close registry
		err := registry.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		// Verify sessions are closed
		if mockSub1.Active() || mockSub2.Active() {
			t.Error("Sessions should be inactive after registry close")
		}
	})
}
