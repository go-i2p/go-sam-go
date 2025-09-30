package datagram

import (
	"testing"
	"time"
)

func TestDatagramSession_ConcurrentOperations(t *testing.T) {
	// Add overall test timeout
	timeout := time.After(30 * time.Second)
	done := make(chan bool)

	go func() {
		sam, keys := setupTestSAM(t)
		defer sam.Close()

		session, err := NewDatagramSession(sam, "test_concurrent", keys, nil)
		if err != nil {
			t.Errorf("Failed to create session: %v", err)
			done <- false
			return
		}
		defer session.Close()

		// Test concurrent reader creation
		readerDone := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				reader := session.NewReader()
				if reader == nil {
					t.Error("NewReader returned nil")
				}
				// Immediately close reader to prevent resource leaks
				reader.Close()
				readerDone <- true
			}()
		}

		// Test concurrent writer creation
		writerDone := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				writer := session.NewWriter()
				if writer == nil {
					t.Error("NewWriter returned nil")
				}
				writerDone <- true
			}()
		}

		// Wait for all goroutines with timeout
		for i := 0; i < 10; i++ {
			select {
			case <-readerDone:
			case <-time.After(5 * time.Second):
				t.Error("Timeout waiting for reader creation")
				done <- false
				return
			}
		}

		for i := 0; i < 10; i++ {
			select {
			case <-writerDone:
			case <-time.After(2 * time.Second):
				t.Error("Timeout waiting for writer creation")
				done <- false
				return
			}
		}

		done <- true
	}()

	select {
	case success := <-done:
		if !success {
			t.Fatal("Test failed")
		}
	case <-timeout:
		t.Fatal("Test timeout - likely goroutine leak or blocking operation")
	}
}
