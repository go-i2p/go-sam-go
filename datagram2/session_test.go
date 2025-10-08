package datagram2

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
// DATAGRAM2 sessions use a unique prefix to distinguish from legacy DATAGRAM sessions.
func generateUniqueSessionID(testName string) string {
	// Use timestamp (nanoseconds) and random number to ensure uniqueness across concurrent executions
	timestamp := time.Now().UnixNano()
	random := rand.Intn(99999)
	return fmt.Sprintf("dg2_%s_%d_%05d", testName, timestamp, random)
}

func TestNewDatagram2Session(t *testing.T) {
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
			name:    "basic datagram2 session creation",
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

			t.Logf("Creating DATAGRAM2 session with ID: %s", sessionID)
			t.Logf("Note: I2P tunnel establishment can take 2-5 minutes")

			session, err := NewDatagram2Session(sam, sessionID, keys, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatagram2Session() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify session properties
				if session.ID() != sessionID {
					t.Errorf("Session ID = %v, want %v", session.ID(), sessionID)
				}

				if session.Keys().Addr().Base32() != keys.Addr().Base32() {
					t.Error("Session keys don't match provided keys")
				}

				addr := session.Addr()
				if addr.Base32() == "" {
					t.Error("Session address is empty")
				}

				// Verify UDP forwarding is enabled (DATAGRAM2 requires this)
				if !session.udpEnabled {
					t.Error("UDP forwarding should be enabled for DATAGRAM2")
				}

				if session.udpConn == nil {
					t.Error("UDP connection should be initialized for DATAGRAM2")
				}

				t.Logf("Session created successfully: %s", addr.Base32())

				// Clean up
				if err := session.Close(); err != nil {
					t.Errorf("Failed to close session: %v", err)
				}
			}
		})
	}
}

func TestDatagram2Session_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	// Generate unique session ID to prevent conflicts during concurrent test execution
	sessionID := generateUniqueSessionID("test_close")

	t.Logf("Creating DATAGRAM2 session for close test")
	session, err := NewDatagram2Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Close the session
	err = session.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Closing again should not error (idempotent)
	err = session.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}

	// Verify session is marked as closed
	if !session.closed {
		t.Error("Session should be marked as closed")
	}
}

func TestDatagram2Session_Addr(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	// Generate unique session ID to prevent conflicts during concurrent test execution
	sessionID := generateUniqueSessionID("test_addr")

	session, err := NewDatagram2Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	addr := session.Addr()
	expectedAddr := keys.Addr()

	if addr.Base32() != expectedAddr.Base32() {
		t.Errorf("Addr() = %v, want %v", addr.Base32(), expectedAddr.Base32())
	}

	if addr.Base64() != expectedAddr.Base64() {
		t.Error("Base64 addresses don't match")
	}
}

func TestDatagram2Session_NewReader(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	// Generate unique session ID to prevent conflicts during concurrent test execution
	sessionID := generateUniqueSessionID("test_reader")

	session, err := NewDatagram2Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	reader := session.NewReader()
	if reader == nil {
		t.Error("NewReader() returned nil")
	}

	if reader.session != session {
		t.Error("Reader session reference is incorrect")
	}

	// Verify channels are initialized
	if reader.recvChan == nil {
		t.Error("Reader recvChan is nil")
	}
	if reader.errorChan == nil {
		t.Error("Reader errorChan is nil")
	}
	if reader.closeChan == nil {
		t.Error("Reader closeChan is nil")
	}
	if reader.doneChan == nil {
		t.Error("Reader doneChan is nil")
	}

	// Clean up reader
	if err := reader.Close(); err != nil {
		t.Errorf("Failed to close reader: %v", err)
	}
}

func TestDatagram2Session_NewWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	// Generate unique session ID to prevent conflicts during concurrent test execution
	sessionID := generateUniqueSessionID("test_writer")

	session, err := NewDatagram2Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	writer := session.NewWriter()
	if writer == nil {
		t.Error("NewWriter() returned nil")
	}

	if writer.session != session {
		t.Error("Writer session reference is incorrect")
	}

	if writer.timeout != 30 {
		t.Errorf("Writer timeout = %v, want 30", writer.timeout)
	}

	// Test method chaining with SetTimeout
	writer2 := writer.SetTimeout(60 * time.Second)
	if writer2 != writer {
		t.Error("SetTimeout should return the same writer instance")
	}
	if writer.timeout != 60*time.Second {
		t.Errorf("Writer timeout = %v, want 60s", writer.timeout)
	}
}

func TestDatagram2Session_PacketConn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	// Generate unique session ID to prevent conflicts during concurrent test execution
	sessionID := generateUniqueSessionID("test_packetconn")

	session, err := NewDatagram2Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	conn := session.PacketConn()
	if conn == nil {
		t.Error("PacketConn() returned nil")
	}

	datagram2Conn, ok := conn.(*Datagram2Conn)
	if !ok {
		t.Error("PacketConn() did not return a Datagram2Conn")
	}

	if datagram2Conn.session != session {
		t.Error("Datagram2Conn session reference is incorrect")
	}

	if datagram2Conn.reader == nil {
		t.Error("Datagram2Conn reader is nil")
	}

	if datagram2Conn.writer == nil {
		t.Error("Datagram2Conn writer is nil")
	}

	// Test LocalAddr
	localAddr := conn.LocalAddr()
	if localAddr == nil {
		t.Error("LocalAddr() returned nil")
	}

	if localAddr.Network() != "datagram2" {
		t.Errorf("LocalAddr().Network() = %v, want datagram2", localAddr.Network())
	}

	// Clean up
	if err := conn.Close(); err != nil {
		t.Errorf("Failed to close PacketConn: %v", err)
	}
}

func TestDatagram2Addr_Network(t *testing.T) {
	addr := &Datagram2Addr{}
	if addr.Network() != "datagram2" {
		t.Errorf("Network() = %v, want datagram2", addr.Network())
	}
}

func TestDatagram2Addr_String(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	addr := &Datagram2Addr{addr: keys.Addr()}
	expected := keys.Addr().Base32()

	if addr.String() != expected {
		t.Errorf("String() = %v, want %v", addr.String(), expected)
	}
}

func TestEnsureUDPForwardingParameters(t *testing.T) {
	tests := []struct {
		name     string
		options  []string
		udpPort  int
		wantHost bool
		wantPort bool
	}{
		{
			name:     "empty options",
			options:  []string{},
			udpPort:  12345,
			wantHost: true,
			wantPort: true,
		},
		{
			name:     "existing HOST",
			options:  []string{"HOST=192.168.1.1"},
			udpPort:  12345,
			wantHost: true,
			wantPort: true,
		},
		{
			name:     "existing PORT",
			options:  []string{"PORT=9999"},
			udpPort:  12345,
			wantHost: true,
			wantPort: true,
		},
		{
			name:     "both existing",
			options:  []string{"HOST=192.168.1.1", "PORT=9999"},
			udpPort:  12345,
			wantHost: true,
			wantPort: true,
		},
		{
			name:     "with other options",
			options:  []string{"inbound.length=3", "outbound.length=3"},
			udpPort:  54321,
			wantHost: true,
			wantPort: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureUDPForwardingParameters(tt.options, tt.udpPort)

			hasHost := false
			hasPort := false

			for _, opt := range result {
				if len(opt) >= 5 && opt[:5] == "HOST=" {
					hasHost = true
				}
				if len(opt) >= 5 && opt[:5] == "PORT=" {
					hasPort = true
				}
			}

			if hasHost != tt.wantHost {
				t.Errorf("Has HOST = %v, want %v", hasHost, tt.wantHost)
			}

			if hasPort != tt.wantPort {
				t.Errorf("Has PORT = %v, want %v", hasPort, tt.wantPort)
			}

			// Verify original options are preserved
			for _, orig := range tt.options {
				found := false
				for _, res := range result {
					if res == orig {
						found = true
						break
					}
				}
				// Only check non-HOST/PORT options
				if len(orig) < 5 || (orig[:5] != "HOST=" && orig[:5] != "PORT=") {
					if !found {
						t.Errorf("Original option %q not found in result", orig)
					}
				}
			}
		})
	}
}

func TestDatagram2Reader_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_reader_close")

	session, err := NewDatagram2Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	reader := session.NewReader()

	// Close the reader
	err = reader.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Closing again should not error (idempotent)
	err = reader.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}

	// Verify reader is marked as closed
	reader.mu.RLock()
	closed := reader.closed
	reader.mu.RUnlock()

	if !closed {
		t.Error("Reader should be marked as closed")
	}
}

func TestDatagram2Writer_SetTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping I2P integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	sessionID := generateUniqueSessionID("test_writer_timeout")

	session, err := NewDatagram2Session(sam, sessionID, keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	writer := session.NewWriter()

	// Test various timeout values
	timeouts := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		60 * time.Second,
		2 * time.Minute,
	}

	for _, timeout := range timeouts {
		returned := writer.SetTimeout(timeout)
		if returned != writer {
			t.Error("SetTimeout should return the same writer for chaining")
		}
		if writer.timeout != timeout {
			t.Errorf("SetTimeout(%v): timeout = %v, want %v", timeout, writer.timeout, timeout)
		}
	}
}

// BenchmarkNewDatagram2Session measures session creation performance
// Note: This requires an I2P router and can be slow (2-5 minutes per iteration)
func BenchmarkNewDatagram2Session(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping I2P integration benchmark in short mode")
	}

	sam, err := common.NewSAM(testSAMAddr)
	if err != nil {
		b.Fatalf("Failed to create SAM connection: %v", err)
	}
	defer sam.Close()

	keys, err := sam.NewKeys()
	if err != nil {
		b.Fatalf("Failed to generate keys: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := generateUniqueSessionID(fmt.Sprintf("bench_%d", i))
		session, err := NewDatagram2Session(sam, sessionID, keys, nil)
		if err != nil {
			b.Errorf("Failed to create session: %v", err)
			continue
		}
		session.Close()
	}
}
