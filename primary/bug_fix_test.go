package primary

import (
	"testing"

	"github.com/go-i2p/go-sam-go/common"
)

// TestMultipleStreamSubSessionsBugFix tests that multiple stream sub-sessions
// can be created without port conflicts (fixes the "Duplicate protocol" bug)
func TestMultipleStreamSubSessionsBugFix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires I2P SAM bridge in short mode")
	}

	// Create SAM connection
	sam, err := common.NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge (is I2P running?): %v", err)
	}
	defer sam.Close()

	// Generate keys
	keys, err := sam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate test keys: %v", err)
	}

	// Create primary session
	session, err := NewPrimarySession(sam, "test-multi-stream-12345", keys, []string{})
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}
	defer session.Close()

	// Create first stream sub-session (should work)
	stream1, err := session.NewStreamSubSession("service1", []string{})
	if err != nil {
		t.Fatalf("Failed to create first stream sub-session: %v", err)
	}
	defer stream1.Close()

	// Create second stream sub-session (this should now work with auto-port assignment)
	stream2, err := session.NewStreamSubSession("service2", []string{})
	if err != nil {
		t.Fatalf("Failed to create second stream sub-session (bug fix failed): %v", err)
	}
	defer stream2.Close()

	// Create third stream sub-session to verify multiple auto-assignments work
	stream3, err := session.NewStreamSubSession("service3", []string{})
	if err != nil {
		t.Fatalf("Failed to create third stream sub-session: %v", err)
	}
	defer stream3.Close()

	// Verify all sub-sessions are active
	if session.SubSessionCount() != 3 {
		t.Errorf("Expected 3 sub-sessions, got %d", session.SubSessionCount())
	}

	t.Log("Successfully created multiple stream sub-sessions - bug fix verified!")
}

// TestPortCleanupOnClose tests that auto-assigned ports are properly released
func TestPortCleanupOnClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires I2P SAM bridge in short mode")
	}

	// Create SAM connection
	sam, err := common.NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge (is I2P running?): %v", err)
	}
	defer sam.Close()

	// Generate keys
	keys, err := sam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate test keys: %v", err)
	}

	// Create primary session
	session, err := NewPrimarySession(sam, "test-port-cleanup-67890", keys, []string{})
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}
	defer session.Close()

	// Create stream sub-session with auto-assigned port
	_, err = session.NewStreamSubSession("auto-port-test", []string{})
	if err != nil {
		t.Fatalf("Failed to create stream sub-session: %v", err)
	}

	// Check that port is tracked
	session.mu.RLock()
	initialPortCount := len(session.subSessionPorts)
	initialUsedCount := len(session.usedPorts)
	session.mu.RUnlock()

	if initialPortCount != 1 {
		t.Errorf("Expected 1 tracked port, got %d", initialPortCount)
	}

	// Close sub-session
	err = session.CloseSubSession("auto-port-test")
	if err != nil {
		t.Fatalf("Failed to close sub-session: %v", err)
	}

	// Check that port was released
	session.mu.RLock()
	finalPortCount := len(session.subSessionPorts)
	finalUsedCount := len(session.usedPorts)
	session.mu.RUnlock()

	if finalPortCount != 0 {
		t.Errorf("Expected 0 tracked ports after close, got %d", finalPortCount)
	}

	if finalUsedCount >= initialUsedCount {
		t.Errorf("Expected fewer used ports after close, initial: %d, final: %d", initialUsedCount, finalUsedCount)
	}

	t.Log("Port cleanup on close verified!")
}

// TestNewStreamSubSessionWithPort tests the explicit port assignment method
func TestNewStreamSubSessionWithPort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires I2P SAM bridge in short mode")
	}

	// Create SAM connection
	sam, err := common.NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge (is I2P running?): %v", err)
	}
	defer sam.Close()

	// Generate keys
	keys, err := sam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate test keys: %v", err)
	}

	// Create primary session
	session, err := NewPrimarySession(sam, "test-explicit-ports-99999", keys, []string{})
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}
	defer session.Close()

	// Test 1: Create sub-session with explicit FROM_PORT
	stream1, err := session.NewStreamSubSessionWithPort("explicit-from", []string{}, 50001, 0)
	if err != nil {
		t.Fatalf("Failed to create sub-session with explicit FROM_PORT: %v", err)
	}
	defer stream1.Close()

	// Test 2: Create sub-session with explicit TO_PORT
	stream2, err := session.NewStreamSubSessionWithPort("explicit-to", []string{}, 0, 50002)
	if err != nil {
		t.Fatalf("Failed to create sub-session with explicit TO_PORT: %v", err)
	}
	defer stream2.Close()

	// Test 3: Create sub-session with both ports
	stream3, err := session.NewStreamSubSessionWithPort("explicit-both", []string{}, 50003, 50004)
	if err != nil {
		t.Fatalf("Failed to create sub-session with both ports: %v", err)
	}
	defer stream3.Close()

	// Test 4: Try to create sub-session with duplicate port (should fail)
	_, err = session.NewStreamSubSessionWithPort("duplicate-port", []string{}, 50001, 0)
	if err == nil {
		t.Errorf("Expected error when using duplicate port, but got none")
	}

	// Test 5: Create sub-session with same port for FROM_PORT and TO_PORT (should work)
	stream5, err := session.NewStreamSubSessionWithPort("same-ports", []string{}, 50005, 50005)
	if err != nil {
		t.Fatalf("Failed to create sub-session with same FROM_PORT and TO_PORT: %v", err)
	}
	defer stream5.Close()

	// Verify all sub-sessions are active
	if session.SubSessionCount() != 4 { // stream1, stream2, stream3, stream5
		t.Errorf("Expected 4 sub-sessions, got %d", session.SubSessionCount())
	}

	t.Log("NewStreamSubSessionWithPort method verified!")
}
