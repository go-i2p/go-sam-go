package raw

import (
	"sync"
	"testing"

	"github.com/go-i2p/go-sam-go/common"
)

func TestRawReader_ConcurrentClose(t *testing.T) {
	// Test concurrent Close() calls don't panic
	session := &RawSession{
		BaseSession: &common.BaseSession{},
		closed:      false,
	}

	reader := &RawReader{
		session:   session,
		recvChan:  make(chan *RawDatagram, 10),
		errorChan: make(chan error, 1),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
		closed:    false,
		mu:        sync.RWMutex{},
	}

	// Start receive loop
	go reader.receiveLoop()

	// Simulate concurrent close attempts
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = reader.Close() // Should not panic
		}()
	}

	wg.Wait()

	// Verify reader is properly closed
	if !reader.closed {
		t.Error("Reader should be marked as closed")
	}
}

func TestRawReader_CloseRaceCondition(t *testing.T) {
	// Test that rapid close after start doesn't cause channel panic
	for i := 0; i < 100; i++ {
		session := &RawSession{closed: false}
		reader := session.NewReader()

		go reader.receiveLoop()

		// Close immediately to trigger race condition
		if err := reader.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	}
}
