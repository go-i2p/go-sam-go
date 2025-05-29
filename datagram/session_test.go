package datagram

import (
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

func TestNewDatagramSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name    string
		id      string
		options []string
		wantErr bool
	}{
		{
			name:    "basic session creation",
			id:      "test_datagram_session",
			options: nil,
			wantErr: false,
		},
		{
			name:    "session with options",
			id:      "test_datagram_with_opts",
			options: []string{"inbound.length=1", "outbound.length=1"},
			wantErr: false,
		},
		{
			name: "session with small tunnel config",
			id:   "test_datagram_small",
			options: []string{
				"inbound.length=0",
				"outbound.length=0",
				"inbound.lengthVariance=0",
				"outbound.lengthVariance=0",
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

			session, err := NewDatagramSession(sam, tt.id, keys, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatagramSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify session properties
				if session.ID() != tt.id {
					t.Errorf("Session ID = %v, want %v", session.ID(), tt.id)
				}

				if session.Keys().Addr().Base32() != keys.Addr().Base32() {
					t.Error("Session keys don't match provided keys")
				}

				addr := session.Addr()
				if addr.Base32() == "" {
					t.Error("Session address is empty")
				}

				// Clean up
				if err := session.Close(); err != nil {
					t.Errorf("Failed to close session: %v", err)
				}
			}
		})
	}
}

func TestDatagramSession_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_close", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Close the session
	err = session.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Closing again should not error
	err = session.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestDatagramSession_Addr(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_addr", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	addr := session.Addr()
	expectedAddr := keys.Addr()

	if addr.Base32() != expectedAddr.Base32() {
		t.Errorf("Addr() = %v, want %v", addr.Base32(), expectedAddr.Base32())
	}
}

func TestDatagramSession_NewReader(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_reader", keys, nil)
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
}

func TestDatagramSession_NewWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_writer", keys, nil)
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
}

func TestDatagramSession_PacketConn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_packetconn", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	conn := session.PacketConn()
	if conn == nil {
		t.Error("PacketConn() returned nil")
	}

	datagramConn, ok := conn.(*DatagramConn)
	if !ok {
		t.Error("PacketConn() did not return a DatagramConn")
	}

	if datagramConn.session != session {
		t.Error("DatagramConn session reference is incorrect")
	}

	if datagramConn.reader == nil {
		t.Error("DatagramConn reader is nil")
	}

	if datagramConn.writer == nil {
		t.Error("DatagramConn writer is nil")
	}
}

func TestDatagramSession_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_concurrent", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Test concurrent reader creation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			reader := session.NewReader()
			if reader == nil {
				t.Error("NewReader returned nil")
			}
			done <- true
		}()
	}

	// Test concurrent writer creation
	for i := 0; i < 10; i++ {
		go func() {
			writer := session.NewWriter()
			if writer == nil {
				t.Error("NewWriter returned nil")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	timeout := time.After(5 * time.Second)
	for i := 0; i < 20; i++ {
		select {
		case <-done:
			// OK
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

func TestDatagramAddr_Network(t *testing.T) {
	addr := &DatagramAddr{}
	if addr.Network() != "i2p-datagram" {
		t.Errorf("Network() = %v, want i2p-datagram", addr.Network())
	}
}

func TestDatagramAddr_String(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	addr := &DatagramAddr{addr: keys.Addr()}
	expected := keys.Addr().Base32()

	if addr.String() != expected {
		t.Errorf("String() = %v, want %v", addr.String(), expected)
	}
}
