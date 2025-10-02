package primary

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

const testSAMAddr = "127.0.0.1:7656"

func setupTestSAM(t *testing.T) (*common.SAM, i2pkeys.I2PKeys) {
	t.Helper()

	sam, err := common.NewSAM(testSAMAddr)
	if err != nil {
		t.Fatalf("Failed to create SAM connection: %v", err)
	}

	keys, err := sam.NewKeys()
	if err != nil {
		sam.Close()
		t.Fatalf("Failed to generate keys: %v", err)
	}

	return sam, keys
}

// generateUniqueSessionID creates a unique session ID to prevent conflicts during concurrent test execution.
// This ensures test isolation when multiple tests run simultaneously (e.g., during race detection).
func generateUniqueSessionID(testName string) string {
	// Use timestamp (nanoseconds) and random number to ensure uniqueness across concurrent executions
	timestamp := time.Now().UnixNano()
	random := rand.Intn(99999)
	return fmt.Sprintf("%s_%d_%05d", testName, timestamp, random)
}

func TestNewPrimarySession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name    string
		idBase  string
		options []string
		wantErr bool
	}{
		{
			name:    "basic primary session creation",
			idBase:  "test_primary_basic",
			options: nil,
			wantErr: false,
		},
		{
			name:    "primary session with options",
			idBase:  "test_primary_opts",
			options: []string{"inbound.length=1", "outbound.length=1"},
			wantErr: false,
		},
		{
			name:   "primary session with small tunnel config",
			idBase: "test_primary_small",
			options: []string{
				"inbound.length=0",
				"outbound.length=0",
				"inbound.quantity=1",
				"outbound.quantity=1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sam, keys := setupTestSAM(t)
			defer sam.Close()

			sessionID := generateUniqueSessionID(tt.idBase)

			session, err := NewPrimarySession(sam, sessionID, keys, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPrimarySession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				defer session.Close()

				// Test basic session properties
				if session.ID() != sessionID {
					t.Errorf("Session ID mismatch: got %s, want %s", session.ID(), sessionID)
				}

				if session.Keys().String() != keys.String() {
					t.Errorf("Session keys mismatch")
				}

				addr := session.Addr()
				if addr.String() == "" {
					t.Error("Session address is empty")
				}

				// Test initial state
				if session.SubSessionCount() != 0 {
					t.Errorf("Expected 0 sub-sessions, got %d", session.SubSessionCount())
				}

				subSessions := session.ListSubSessions()
				if len(subSessions) != 0 {
					t.Errorf("Expected empty sub-session list, got %d items", len(subSessions))
				}
			}
		})
	}
}

func TestPrimarySessionSubSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_primary_subsessions")
	session, err := NewPrimarySession(sam, sessionID, keys, []string{
		"inbound.length=1",
		"outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}
	defer session.Close()

	t.Run("create stream sub-session", func(t *testing.T) {
		streamSubID := "stream_sub_1"
		streamSub, err := session.NewStreamSubSession(streamSubID, []string{})
		if err != nil {
			t.Fatalf("Failed to create stream sub-session: %v", err)
		}
		defer streamSub.Close()

		if streamSub.ID() != streamSubID {
			t.Errorf("Stream sub-session ID mismatch: got %s, want %s", streamSub.ID(), streamSubID)
		}

		if streamSub.Type() != "STREAM" {
			t.Errorf("Stream sub-session type mismatch: got %s, want STREAM", streamSub.Type())
		}

		if !streamSub.Active() {
			t.Error("Stream sub-session should be active")
		}

		// Check it's registered
		if session.SubSessionCount() != 1 {
			t.Errorf("Expected 1 sub-session, got %d", session.SubSessionCount())
		}

		retrievedSub, err := session.GetSubSession(streamSubID)
		if err != nil {
			t.Fatalf("Failed to retrieve sub-session: %v", err)
		}

		if retrievedSub.ID() != streamSubID {
			t.Errorf("Retrieved sub-session ID mismatch: got %s, want %s", retrievedSub.ID(), streamSubID)
		}
	})

	t.Run("create datagram sub-session", func(t *testing.T) {
		datagramSubID := "datagram_sub_1"
		// DATAGRAM subsessions require a PORT parameter per SAM v3.3 specification
		datagramSub, err := session.NewDatagramSubSession(datagramSubID, []string{"PORT=8080"})
		if err != nil {
			t.Fatalf("Failed to create datagram sub-session: %v", err)
		}
		defer datagramSub.Close()

		if datagramSub.ID() != datagramSubID {
			t.Errorf("Datagram sub-session ID mismatch: got %s, want %s", datagramSub.ID(), datagramSubID)
		}

		if datagramSub.Type() != "DATAGRAM" {
			t.Errorf("Datagram sub-session type mismatch: got %s, want DATAGRAM", datagramSub.Type())
		}

		if !datagramSub.Active() {
			t.Error("Datagram sub-session should be active")
		}

		// Check it's registered (should be 2 now with stream sub-session)
		if session.SubSessionCount() != 2 {
			t.Errorf("Expected 2 sub-sessions, got %d", session.SubSessionCount())
		}
	})

	t.Run("create raw sub-session", func(t *testing.T) {
		rawSubID := "raw_sub_1"
		// RAW subsessions require a PORT parameter per SAM v3.3 specification
		rawSub, err := session.NewRawSubSession(rawSubID, []string{"PORT=8081"})
		if err != nil {
			t.Fatalf("Failed to create raw sub-session: %v", err)
		}
		defer rawSub.Close()

		if rawSub.ID() != rawSubID {
			t.Errorf("Raw sub-session ID mismatch: got %s, want %s", rawSub.ID(), rawSubID)
		}

		if rawSub.Type() != "RAW" {
			t.Errorf("Raw sub-session type mismatch: got %s, want RAW", rawSub.Type())
		}

		if !rawSub.Active() {
			t.Error("Raw sub-session should be active")
		}

		// Check it's registered (should be 3 now)
		if session.SubSessionCount() != 3 {
			t.Errorf("Expected 3 sub-sessions, got %d", session.SubSessionCount())
		}
	})

	t.Run("list all sub-sessions", func(t *testing.T) {
		subSessions := session.ListSubSessions()
		if len(subSessions) != 3 {
			t.Errorf("Expected 3 sub-sessions in list, got %d", len(subSessions))
		}

		// Check all types are present
		types := make(map[string]bool)
		for _, sub := range subSessions {
			types[sub.Type()] = true
		}

		expectedTypes := []string{"STREAM", "DATAGRAM", "RAW"}
		for _, expectedType := range expectedTypes {
			if !types[expectedType] {
				t.Errorf("Expected sub-session type %s not found", expectedType)
			}
		}
	})
}

func TestSubSessionCloseAndCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_primary_cleanup")
	session, err := NewPrimarySession(sam, sessionID, keys, []string{
		"inbound.length=1",
		"outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}
	defer session.Close()

	// Create a sub-session
	streamSubID := "stream_sub_cleanup"
	streamSub, err := session.NewStreamSubSession(streamSubID, []string{})
	if err != nil {
		t.Fatalf("Failed to create stream sub-session: %v", err)
	}

	// Verify it's registered
	if session.SubSessionCount() != 1 {
		t.Errorf("Expected 1 sub-session, got %d", session.SubSessionCount())
	}

	// Close the sub-session manually
	err = session.CloseSubSession(streamSubID)
	if err != nil {
		t.Fatalf("Failed to close sub-session: %v", err)
	}

	// Verify it's unregistered
	if session.SubSessionCount() != 0 {
		t.Errorf("Expected 0 sub-sessions after close, got %d", session.SubSessionCount())
	}

	// Verify sub-session is inactive
	if streamSub.Active() {
		t.Error("Sub-session should be inactive after close")
	}

	// Try to retrieve it (should fail)
	_, err = session.GetSubSession(streamSubID)
	if err == nil {
		t.Error("Expected error when retrieving closed sub-session")
	}
}

func TestPrimarySessionCascadeClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_primary_cascade")
	session, err := NewPrimarySession(sam, sessionID, keys, []string{
		"inbound.length=1",
		"outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}

	// Create multiple sub-sessions
	streamSub, err := session.NewStreamSubSession("stream_cascade", []string{})
	if err != nil {
		t.Fatalf("Failed to create stream sub-session: %v", err)
	}

	datagramSub, err := session.NewDatagramSubSession("datagram_cascade", []string{"PORT=8082"})
	if err != nil {
		t.Fatalf("Failed to create datagram sub-session: %v", err)
	}

	rawSub, err := session.NewRawSubSession("raw_cascade", []string{"PORT=8083"})
	if err != nil {
		t.Fatalf("Failed to create raw sub-session: %v", err)
	}

	// Verify all are active
	if session.SubSessionCount() != 3 {
		t.Errorf("Expected 3 sub-sessions, got %d", session.SubSessionCount())
	}

	if !streamSub.Active() || !datagramSub.Active() || !rawSub.Active() {
		t.Error("All sub-sessions should be active before primary close")
	}

	// Close the primary session
	err = session.Close()
	if err != nil {
		t.Fatalf("Failed to close primary session: %v", err)
	}

	// Verify all sub-sessions are inactive
	if streamSub.Active() || datagramSub.Active() || rawSub.Active() {
		t.Error("All sub-sessions should be inactive after primary close")
	}

	// Verify session count is 0
	if session.SubSessionCount() != 0 {
		t.Errorf("Expected 0 sub-sessions after primary close, got %d", session.SubSessionCount())
	}
}

func TestConcurrentSubSessionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_primary_concurrent")
	session, err := NewPrimarySession(sam, sessionID, keys, []string{
		"inbound.length=1",
		"outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}
	defer session.Close()

	const numGoroutines = 3 // Realistic number for I2P SAM PRIMARY session capabilities
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Concurrent sub-session creation with realistic usage patterns
	// Create one session of each type to test concurrency without SAM protocol limits
	sessionTypes := []string{"STREAM", "DATAGRAM", "RAW"}
	basePort := 8000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			subSessionID := fmt.Sprintf("concurrent_sub_%d", id)
			sessionType := sessionTypes[id] // One of each type
			port := basePort + id

			// Create sub-session with correct SAM protocol syntax for each session type
			var err error
			switch sessionType {
			case "STREAM":
				_, err = session.NewStreamSubSession(subSessionID, []string{"FROM_PORT=" + strconv.Itoa(port)})
			case "DATAGRAM":
				_, err = session.NewDatagramSubSession(subSessionID, []string{"PORT=" + strconv.Itoa(port)})
			case "RAW":
				_, err = session.NewRawSubSession(subSessionID, []string{"PORT=" + strconv.Itoa(port)})
			}

			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("goroutine %d (%s): %w", id, sessionType, err))
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
		}
	}

	// Verify expected number of sub-sessions
	if session.SubSessionCount() != numGoroutines {
		t.Errorf("Expected %d sub-sessions, got %d", numGoroutines, session.SubSessionCount())
	}

	// Concurrent list operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subSessions := session.ListSubSessions()
			if len(subSessions) != numGoroutines {
				mu.Lock()
				errors = append(errors, fmt.Errorf("concurrent list: expected %d, got %d", numGoroutines, len(subSessions)))
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Check for list operation errors
	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("Concurrent list operation error: %v", err)
		}
	}
}

func TestSubSessionRegistryOperations(t *testing.T) {
	registry := NewSubSessionRegistry()
	defer registry.Close()

	// Test with mock sub-sessions
	mockSub1 := &mockSubSession{id: "test1", sessionType: "STREAM", active: true}
	mockSub2 := &mockSubSession{id: "test2", sessionType: "DATAGRAM", active: true}

	// Test Register
	err := registry.Register("test1", mockSub1)
	if err != nil {
		t.Errorf("Failed to register sub-session: %v", err)
	}

	// Test duplicate registration
	err = registry.Register("test1", mockSub2)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test Get
	retrieved, exists := registry.Get("test1")
	if !exists {
		t.Error("Expected to find registered sub-session")
	}
	if retrieved.ID() != "test1" {
		t.Errorf("Retrieved wrong sub-session: got %s, want test1", retrieved.ID())
	}

	// Test Count
	if registry.Count() != 1 {
		t.Errorf("Expected count 1, got %d", registry.Count())
	}

	// Test register second
	err = registry.Register("test2", mockSub2)
	if err != nil {
		t.Errorf("Failed to register second sub-session: %v", err)
	}

	// Test List
	sessions := registry.List()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions in list, got %d", len(sessions))
	}

	// Test Unregister
	err = registry.Unregister("test1")
	if err != nil {
		t.Errorf("Failed to unregister sub-session: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected count 1 after unregister, got %d", registry.Count())
	}

	// Test unregister non-existent
	err = registry.Unregister("nonexistent")
	if err == nil {
		t.Error("Expected error for unregistering non-existent session")
	}
}

// mockSubSession implements SubSession for testing
type mockSubSession struct {
	id          string
	sessionType string
	active      bool
	closed      bool
}

func (m *mockSubSession) ID() string   { return m.id }
func (m *mockSubSession) Type() string { return m.sessionType }
func (m *mockSubSession) Active() bool { return m.active && !m.closed }
func (m *mockSubSession) Close() error { m.closed = true; m.active = false; return nil }
