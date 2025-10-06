package sam3

import (
	"fmt"
	"testing"
	"time"
)

func TestDebugDatagramBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping debug test in short mode")
	}

	fmt.Println("=== DEBUG: Starting basic datagram test ===")

	// Step 1: Create SAM connection
	fmt.Println("DEBUG: Creating SAM connection...")
	sam, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("DEBUG: Failed to create SAM: %v\n", err)
		t.Fatal(err)
	}
	defer sam.Close()
	fmt.Println("DEBUG: SAM connection created successfully")

	// Step 2: Generate keys
	fmt.Println("DEBUG: Generating keys...")
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("DEBUG: Failed to generate keys: %v\n", err)
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: Keys generated, address: %s\n", keys.Addr().Base32())

	// Step 3: Create datagram session
	fmt.Println("DEBUG: Creating datagram session...")
	dg, err := sam.NewDatagramSession("DEBUG_TEST", keys, Options_Small, 0)
	if err != nil {
		fmt.Printf("DEBUG: Failed to create datagram session: %v\n", err)
		t.Fatal(err)
	}
	defer dg.Close()
	fmt.Println("DEBUG: Datagram session created successfully")

	// Step 4: Try to send a message to self
	fmt.Println("DEBUG: Sending message to self...")
	myself := keys.Addr()
	message := []byte("Hello debug test!")

	_, err = dg.WriteTo(message, myself)
	if err != nil {
		fmt.Printf("DEBUG: Failed to send message: %v\n", err)
		t.Fatal(err)
	}
	fmt.Println("DEBUG: Message sent successfully")

	// Step 5: Try to receive with timeout
	fmt.Println("DEBUG: Attempting to receive message...")

	// Set a reasonable timeout for I2P operations
	done := make(chan bool, 1)
	var readErr error
	var n int

	go func() {
		buf := make([]byte, 1024)
		n, _, readErr = dg.ReadFrom(buf)
		if readErr == nil {
			fmt.Printf("DEBUG: Received %d bytes: %s\n", n, string(buf[:n]))
		}
		done <- true
	}()

	select {
	case <-done:
		if readErr != nil {
			fmt.Printf("DEBUG: Read error: %v\n", readErr)
			t.Error(readErr)
		} else {
			fmt.Printf("DEBUG: Successfully received message\n")
		}
	case <-time.After(30 * time.Second):
		fmt.Println("DEBUG: Timeout waiting for message")
		t.Error("Timeout waiting for datagram")
	}
}
