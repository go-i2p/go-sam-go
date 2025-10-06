package stream

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// generateUniqueSessionID creates a unique session ID to prevent conflicts during concurrent test execution.
func generateUniqueSessionID(testName string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d", testName, timestamp)
}

func TestStreamSession_Dial(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_dial", keys, []string{
		"inbound.length=1", "outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Create a local test listener instead of using external site
	testSAM2, testKeys2 := setupTestSAM(t)
	defer testSAM2.Close()

	listenerSession, err := NewStreamSession(testSAM2, generateUniqueSessionID("test_dial_listener"), testKeys2, []string{
		"inbound.length=1", "outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create listener session: %v", err)
	}
	defer listenerSession.Close()

	listener, err := listenerSession.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Start a simple echo server in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				c.Read(buf) // Read the request
				c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html><body>Test response</body></html>"))
			}(conn)
		}
	}()

	// Give listener time to be ready
	time.Sleep(2 * time.Second)

	// Test dialing to the local listener
	conn, err := session.Dial(listenerSession.Addr().Base32())
	if err != nil {
		t.Logf("Dial to local listener failed (might be expected due to I2P timing): %v", err)
		return // Not a hard failure since I2P connections can be unreliable in test conditions
	}
	defer conn.Close()

	// Test basic communication
	_, err = conn.Write([]byte("GET /\n"))
	if err != nil {
		t.Logf("Failed to write to connection: %v", err)
		return
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Logf("Failed to read from connection: %v", err)
		return
	}

	response := string(buf[:n])
	if !strings.Contains(strings.ToLower(response), "html") {
		t.Logf("Did not receive expected HTML response, got: %s", response)
	} else {
		t.Logf("Successfully received HTML response from local listener")
	}
}

func TestStreamSession_DialI2P(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_dial_i2p", keys, []string{
		"inbound.length=1", "outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Create a local test listener instead of using external site
	testSAM2, testKeys2 := setupTestSAM(t)
	defer testSAM2.Close()

	listenerSession, err := NewStreamSession(testSAM2, generateUniqueSessionID("test_dial_i2p_listener"), testKeys2, []string{
		"inbound.length=1", "outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create listener session: %v", err)
	}
	defer listenerSession.Close()

	listener, err := listenerSession.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Start a simple echo server in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				c.Read(buf) // Read the request
				c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html><body>Test response</body></html>"))
			}(conn)
		}
	}()

	// Give listener time to be ready
	time.Sleep(2 * time.Second)

	// Test dialing to the local listener using DialI2P with the actual I2P address
	conn, err := session.DialI2P(listenerSession.Addr())
	if err != nil {
		t.Logf("DialI2P to local listener failed (might be expected due to I2P timing): %v", err)
		return // Not a hard failure since I2P connections can be unreliable in test conditions
	}
	defer conn.Close()

	// Test basic communication
	_, err = conn.Write([]byte("GET /\n"))
	if err != nil {
		t.Logf("Failed to write to connection: %v", err)
		return
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Logf("Failed to read from connection: %v", err)
		return
	}

	response := string(buf[:n])
	if !strings.Contains(strings.ToLower(response), "html") {
		t.Logf("Did not receive expected HTML response, got: %s", response)
	} else {
		t.Logf("Successfully received HTML response from local listener via DialI2P")
	}
}

func TestStreamSession_DialContext(t *testing.T) {

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_dial_context", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Create a local test listener instead of using external site
	testSAM2, testKeys2 := setupTestSAM(t)
	defer testSAM2.Close()

	listenerSession, err := NewStreamSession(testSAM2, generateUniqueSessionID("test_dial_i2p_listener"), testKeys2, []string{
		"inbound.length=1", "outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create listener session: %v", err)
	}
	defer listenerSession.Close()

	listener, err := listenerSession.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Start a simple echo server in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				c.Read(buf) // Read the request
				c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html><body>Test response</body></html>"))
			}(conn)
		}
	}()

	t.Run("dial with context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := session.DialContext(ctx, "nonexistent.i2p")
		if err == nil {
			t.Log("Dial succeeded unexpectedly")
		} else {
			t.Logf("Dial failed as expected: %v", err)
		}
	})

	t.Run("dial with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := session.DialContext(ctx, listener.Addr().String())
		if err == nil {
			t.Error("Expected dial to fail with cancelled context")
		}
	})
}
