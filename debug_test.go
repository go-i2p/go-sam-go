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

	// Step 1: Create first SAM connection for receiver
	fmt.Println("DEBUG: Creating receiver SAM connection...")
	sam1, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("DEBUG: Failed to create receiver SAM: %v\n", err)
		t.Fatal(err)
	}
	defer sam1.Close()
	fmt.Println("DEBUG: Receiver SAM connection created successfully")

	// Step 2: Generate keys for receiver
	fmt.Println("DEBUG: Generating receiver keys...")
	receiverKeys, err := sam1.NewKeys()
	if err != nil {
		fmt.Printf("DEBUG: Failed to generate receiver keys: %v\n", err)
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: Receiver keys generated, address: %s\n", receiverKeys.Addr().Base32())

	// Step 3: Create receiver datagram session
	fmt.Println("DEBUG: Creating receiver datagram session...")
	receiver, err := sam1.NewDatagramSession("DEBUG_RECEIVER", receiverKeys, Options_Small, 0)
	if err != nil {
		fmt.Printf("DEBUG: Failed to create receiver datagram session: %v\n", err)
		t.Fatal(err)
	}
	defer receiver.Close()
	fmt.Println("DEBUG: Receiver datagram session created successfully")

	// Step 4: Create second SAM connection for sender
	fmt.Println("DEBUG: Creating sender SAM connection...")
	sam2, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("DEBUG: Failed to create sender SAM: %v\n", err)
		t.Fatal(err)
	}
	defer sam2.Close()
	fmt.Println("DEBUG: Sender SAM connection created successfully")

	// Step 5: Generate keys for sender
	fmt.Println("DEBUG: Generating sender keys...")
	senderKeys, err := sam2.NewKeys()
	if err != nil {
		fmt.Printf("DEBUG: Failed to generate sender keys: %v\n", err)
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: Sender keys generated, address: %s\n", senderKeys.Addr().Base32())

	// Step 6: Create sender datagram session
	fmt.Println("DEBUG: Creating sender datagram session...")
	sender, err := sam2.NewDatagramSession("DEBUG_SENDER", senderKeys, Options_Small, 0)
	if err != nil {
		fmt.Printf("DEBUG: Failed to create sender datagram session: %v\n", err)
		t.Fatal(err)
	}
	defer sender.Close()
	fmt.Println("DEBUG: Sender datagram session created successfully")

	// Step 7: Send message from sender to receiver
	fmt.Println("DEBUG: Sending message from sender to receiver...")
	fmt.Println("DEBUG: Note: I2P datagram delivery requires tunnel establishment and may take 1-5 minutes")
	
	receiverAddr := receiverKeys.Addr()
	message := []byte("Hello from sender to receiver!")

	// Set up receiver goroutine first
	done := make(chan bool, 1)
	var readErr error
	var n int
	var receivedData []byte

	go func() {
		fmt.Println("DEBUG: Receiver waiting for datagram...")
		buf := make([]byte, 1024)
		n, _, readErr = receiver.ReadFrom(buf)
		if readErr == nil {
			receivedData = buf[:n]
			fmt.Printf("DEBUG: Received %d bytes: %s\n", n, string(receivedData))
		} else {
			fmt.Printf("DEBUG: Receiver error: %v\n", readErr)
		}
		done <- true
	}()

	// Allow time for receiver to start listening
	time.Sleep(2 * time.Second)

	// Send the message
	_, err = sender.WriteTo(message, receiverAddr)
	if err != nil {
		fmt.Printf("DEBUG: Failed to send message: %v\n", err)
		t.Fatal(err)
	}
	fmt.Println("DEBUG: Message sent successfully from sender to receiver")

	// Step 8: Wait for message reception with progress indicator
	fmt.Println("DEBUG: Waiting for message reception...")

	// Progress indicator for I2P operations
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	start := time.Now()

	for {
		select {
		case <-done:
			elapsed := time.Since(start)
			if readErr != nil {
				fmt.Printf("DEBUG: Read error after %v: %v\n", elapsed, readErr)
				t.Error(readErr)
			} else {
				fmt.Printf("DEBUG: Successfully received message after %v\n", elapsed)
				if string(receivedData) != string(message) {
					t.Errorf("Message mismatch: sent %q, received %q", string(message), string(receivedData))
				}
			}
			return
		case <-ticker.C:
			elapsed := time.Since(start)
			fmt.Printf("DEBUG: Still waiting for I2P datagram after %v (tunnels may still be establishing)...\n", elapsed)
		case <-time.After(5 * time.Minute):
			fmt.Println("DEBUG: Timeout waiting for message after 5 minutes")
			t.Error("Timeout waiting for datagram - I2P tunnel establishment may require more time")
			return
		}
	}
}
