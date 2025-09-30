package stream

import (
	"sync"
	"testing"
	"time"
)

// TestListenDeadlockFix tests that Issue #8 from AUDIT.md is resolved
// This validates that Listen() no longer causes RLock/WLock deadlock
func TestListenDeadlockFix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deadlock test in short mode")
	}

	// Create a session
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_deadlock_fix", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test that multiple concurrent Listen() calls don't deadlock
	const numListeners = 3
	done := make(chan *StreamListener, numListeners)
	errors := make(chan error, numListeners)

	// Start multiple Listen() operations concurrently
	// This would previously deadlock due to RLock/WLock ordering issue
	var wg sync.WaitGroup
	for i := 0; i < numListeners; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listener, err := session.Listen()
			if err != nil {
				errors <- err
				return
			}
			done <- listener
		}()
	}

	// Use timeout to detect deadlock
	timeout := time.After(5 * time.Second)
	var listeners []*StreamListener

	// Collect listeners or detect timeout
	go func() {
		wg.Wait()
		close(done)
		close(errors)
	}()

	for {
		select {
		case listener, ok := <-done:
			if !ok {
				goto cleanup
			}
			listeners = append(listeners, listener)
		case err := <-errors:
			t.Errorf("Failed to create listener: %v", err)
		case <-timeout:
			t.Fatalf("Deadlock detected: Listen() operations did not complete within timeout")
		}
	}

cleanup:
	// Clean up listeners first
	for _, listener := range listeners {
		if listener != nil {
			listener.Close()
		}
	}

	// Close session
	session.Close()

	if len(listeners) != numListeners {
		t.Errorf("Expected %d successful listeners, got %d", numListeners, len(listeners))
	}
}

// TestListenConcurrentAccess tests concurrent access patterns that could trigger deadlock
func TestListenConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent access test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_concurrent_access", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	var wg sync.WaitGroup
	timeout := time.After(15 * time.Second)
	done := make(chan struct{})

	// Start operation that completes when all goroutines finish
	go func() {
		wg.Wait()
		close(done)
	}()

	// Test 1: Concurrent Listen() and session state access
	wg.Add(2)
	go func() {
		defer wg.Done()
		listener, err := session.Listen()
		if err != nil {
			t.Errorf("Failed to create listener: %v", err)
			return
		}
		defer listener.Close()
	}()

	go func() {
		defer wg.Done()
		// Access session state while Listen() is running
		addr := session.Addr()
		if addr.String() == "" {
			t.Error("Session address should not be empty")
		}
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// All operations completed successfully
	case <-timeout:
		t.Fatal("Test timed out - likely deadlock in concurrent Listen/access operations")
	}
}

// TestListenRapidCreation tests rapid listener creation to stress test lock handling
func TestListenRapidCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rapid creation test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_rapid_creation", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Create and close listeners rapidly to test for race conditions
	const iterations = 20
	timeout := time.After(30 * time.Second)
	done := make(chan bool, 1)

	go func() {
		for i := 0; i < iterations; i++ {
			listener, err := session.Listen()
			if err != nil {
				t.Errorf("Failed to create listener %d: %v", i, err)
				done <- false
				return
			}

			// Immediately close to test rapid lifecycle
			if err := listener.Close(); err != nil {
				t.Errorf("Failed to close listener %d: %v", i, err)
				done <- false
				return
			}
		}
		done <- true
	}()

	select {
	case success := <-done:
		if !success {
			t.Fatal("Rapid listener creation/destruction failed")
		}
	case <-timeout:
		t.Fatal("Rapid listener creation test timed out - possible deadlock")
	}
}
