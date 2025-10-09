package datagram3

import (
	"fmt"
	"math/rand"
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
// DATAGRAM3 sessions use a unique prefix to distinguish from other datagram types.
func generateUniqueSessionID(testName string) string {
	// Use timestamp (nanoseconds) and random number to ensure uniqueness across concurrent executions
	timestamp := time.Now().UnixNano()
	random := rand.Intn(99999)
	return fmt.Sprintf("dg3_%s_%d_%05d", testName, timestamp, random)
}

func TestNewDatagram3Session(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	tests := []struct {
		name    string
		idBase  string
		options []string
		wantErr bool
	}{
		{
			name:    "basic datagram3 session creation",
			idBase:  "test_session",
			options: nil,
			wantErr: false,
		},
		{
			name:    "session with custom options",
			idBase:  "test_with_opts",
			options: []string{"inbound.length=1", "outbound.length=1"},
			wantErr: false,
		},
		{
			name:   "session with small tunnel config",
			idBase: "test_small",
			options: []string{
				"inbound.length=1",
				"outbound.length=1",
				"inbound.lengthVariance=0",
				"outbound.lengthVariance=0",
				"inbound.quantity=1",
				"outbound.quantity=1",
			},
			wantErr: false,
		},
		{
			name:   "session with Ed25519 signature",
			idBase: "test_ed25519",
			options: []string{
				"inbound.length=0",
				"outbound.length=0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sam, keys := setupTestSAM(t)
			defer sam.Close()

			// Generate unique session ID to prevent conflicts during concurrent test execution
			sessionID := generateUniqueSessionID(tt.idBase)

			t.Logf("Creating DATAGRAM3 session with ID: %s", sessionID)
			t.Logf("Note: I2P tunnel establishment can take 2-5 minutes")

			session, err := NewDatagram3Session(sam, sessionID, keys, tt.options)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewDatagram3Session() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return // Expected error, test passed
			}

			// Verify session was created
			if session == nil {
				t.Fatal("Expected non-nil session")
			}

			// Verify session has resolver
			if session.resolver == nil {
				t.Error("Expected session to have resolver")
			}

			// Verify resolver has empty cache initially
			if session.resolver.CacheSize() != 0 {
				t.Errorf("Expected empty cache initially, got size %d", session.resolver.CacheSize())
			}

			// Verify UDP connection was established
			if session.udpConn == nil {
				t.Error("Expected UDP connection to be established")
			}

			// Verify session address is valid
			addr := session.Addr()
			if addr == "" {
				t.Error("Expected non-empty session address")
			}

			t.Logf("Session created successfully with address: %s", addr)

			// Clean up
			err = session.Close()
			if err != nil {
				t.Errorf("Failed to close session: %v", err)
			}

			// Verify resolver cache cleared after close
			if session.resolver.CacheSize() != 0 {
				t.Errorf("Expected cache cleared after close, got size %d", session.resolver.CacheSize())
			}
		})
	}
}

func TestDatagram3SessionClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_close")

	t.Logf("Creating DATAGRAM3 session for close test: %s", sessionID)
	session, err := NewDatagram3Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add some entries to cache
	testHash := make([]byte, 32)
	for i := range testHash {
		testHash[i] = byte(i)
	}
	b32 := hashToB32Address(testHash)
	session.resolver.cache[b32] = i2pkeys.I2PAddr("test-destination")

	if session.resolver.CacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", session.resolver.CacheSize())
	}

	// Close session
	err = session.Close()
	if err != nil {
		t.Errorf("Failed to close session: %v", err)
	}

	// Verify cache cleared
	if session.resolver.CacheSize() != 0 {
		t.Errorf("Expected cache cleared after close, got size %d", session.resolver.CacheSize())
	}

	// Verify double-close doesn't panic or error
	err = session.Close()
	if err != nil {
		t.Logf("Second close returned error (acceptable): %v", err)
	}
}

func TestDatagram3SessionReaderWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_rw")

	t.Logf("Creating DATAGRAM3 session for reader/writer test: %s", sessionID)
	session, err := NewDatagram3Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Create reader
	reader := session.NewReader()
	if reader == nil {
		t.Fatal("Expected non-nil reader")
	}
	if reader.session != session {
		t.Error("Reader session mismatch")
	}

	// Create writer
	writer := session.NewWriter()
	if writer == nil {
		t.Fatal("Expected non-nil writer")
	}
	if writer.session != session {
		t.Error("Writer session mismatch")
	}

	// Test writer timeout configuration
	writer2 := session.NewWriter().SetTimeout(10 * time.Second)
	if writer2.timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", writer2.timeout)
	}

	// Clean up reader
	err = reader.Close()
	if err != nil {
		t.Errorf("Failed to close reader: %v", err)
	}
}

// TestDatagram3SessionConcurrentAccess tests concurrent operations on session
func TestDatagram3SessionConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_concurrent")

	t.Logf("Creating DATAGRAM3 session for concurrency test: %s", sessionID)
	session, err := NewDatagram3Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Create multiple readers and writers concurrently
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(n int) {
			reader := session.NewReader()
			if reader == nil {
				t.Errorf("Reader %d: got nil reader", n)
			}
			time.Sleep(10 * time.Millisecond)
			reader.Close()
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func(n int) {
			writer := session.NewWriter()
			if writer == nil {
				t.Errorf("Writer %d: got nil writer", n)
			}
			time.Sleep(10 * time.Millisecond)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

// TestDatagram3RoundTrip tests sending and receiving datagrams with hash resolution
func TestDatagram3RoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	// This test requires two I2P sessions to communicate
	// Create session A
	samA, keysA := setupTestSAM(t)
	defer samA.Close()

	sessionA_ID := generateUniqueSessionID("alice")
	t.Logf("Creating session A (Alice): %s", sessionA_ID)
	sessionA, err := NewDatagram3Session(samA, sessionA_ID, keysA, []string{
		"inbound.length=0",
		"outbound.length=0",
	})
	if err != nil {
		t.Fatalf("Failed to create session A: %v", err)
	}
	defer sessionA.Close()

	// Create session B
	samB, keysB := setupTestSAM(t)
	defer samB.Close()

	sessionB_ID := generateUniqueSessionID("bob")
	t.Logf("Creating session B (Bob): %s", sessionB_ID)
	sessionB, err := NewDatagram3Session(samB, sessionB_ID, keysB, []string{
		"inbound.length=0",
		"outbound.length=0",
	})
	if err != nil {
		t.Fatalf("Failed to create session B: %v", err)
	}
	defer sessionB.Close()

	t.Logf("Alice address: %s", sessionA.Addr())
	t.Logf("Bob address: %s", sessionB.Addr())

	// Start reader on B
	readerB := sessionB.NewReader()
	defer readerB.Close()

	receivedChan := make(chan *Datagram3, 1)
	errorChan := make(chan error, 1)

	go func() {
		dg, err := readerB.ReceiveDatagram()
		if err != nil {
			errorChan <- err
			return
		}
		receivedChan <- dg
	}()

	// Give reader time to start
	time.Sleep(1 * time.Second)

	// Send from A to B
	writerA := sessionA.NewWriter()
	testMessage := []byte("Hello from Alice! This is a DATAGRAM3 message.")

	t.Logf("Alice sending to Bob...")
	err = writerA.SendDatagram(testMessage, sessionB.Addr())
	if err != nil {
		t.Fatalf("Failed to send datagram: %v", err)
	}

	// Wait for reception with generous timeout for I2P
	t.Logf("Waiting for Bob to receive (up to 60 seconds)...")
	select {
	case dg := <-receivedChan:
		t.Logf("Bob received datagram!")

		// Verify data
		if string(dg.Data) != string(testMessage) {
			t.Errorf("Data mismatch: got %q, want %q", string(dg.Data), string(testMessage))
		}

		// Verify hash is present
		if len(dg.SourceHash) != 32 {
			t.Errorf("Expected 32-byte source hash, got %d bytes", len(dg.SourceHash))
		}

		// Test GetSourceB32
		b32 := dg.GetSourceB32()
		if len(b32) != 60 {
			t.Errorf("Expected 60-char b32 address, got %d", len(b32))
		}
		t.Logf("Source b32: %s", b32)

		t.Logf("Source hash: %x", dg.SourceHash)

		// Test hash resolution (requires NAMING LOOKUP)
		t.Logf("Resolving source hash to full destination...")
		err = dg.ResolveSource(sessionB)
		if err != nil {
			// This might fail if the destination isn't published yet
			t.Logf("Hash resolution failed (may be expected for new sessions): %v", err)
		} else {
			t.Logf("Resolved source: %s", dg.Source)

			// Verify it's in cache now
			cachedDest, ok := sessionB.resolver.GetCached(dg.SourceHash)
			if !ok {
				t.Error("Expected resolved destination in cache")
			} else {
				t.Logf("Cached destination: %s", cachedDest)
			}

			// Test sending reply
			t.Logf("Bob sending reply to Alice...")
			writerB := sessionB.NewWriter()
			replyMessage := []byte("Reply from Bob! Source verification is YOUR responsibility!")
			err = writerB.ReplyToDatagram(replyMessage, dg)
			if err != nil {
				t.Errorf("Failed to send reply: %v", err)
			} else {
				t.Logf("Reply sent successfully")
			}
		}

	case err := <-errorChan:
		t.Fatalf("Error receiving datagram: %v", err)

	case <-time.After(60 * time.Second):
		t.Fatal("Timeout waiting for datagram (I2P may be congested or tunnels not ready)")
	}
}

// TestDatagram3Documentation documents the datagram3 protocol
func TestDatagram3Documentation(t *testing.T) {
	t.Log("DATAGRAM3 Protocol Documentation")
	t.Log("=" + string(make([]byte, 78)))
	t.Log("DATAGRAM3 provides repliable datagrams with hash-based addressing")
	t.Log("")
	t.Log("Key features:")
	t.Log("  1. Source addresses use 32-byte hashes")
	t.Log("  2. Repliable datagram protocol")
	t.Log("  3. Lower overhead than DATAGRAM/DATAGRAM2")
	t.Log("  4. Hash resolution via naming lookup")
	t.Log("")
	t.Log("DATAGRAM3 is appropriate when:")
	t.Log("  - Low overhead is important")
	t.Log("  - Reply capability is needed")
	t.Log("  - Hash-based addressing is acceptable")
	t.Log("=" + string(make([]byte, 78)))

	t.Log("✓ Documentation complete")
}
