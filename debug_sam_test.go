package sam3

import (
	"fmt"
	"testing"
)

func TestDEBUG_SAM_HELP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping debug test in short mode")
	}

	fmt.Println("=== DEBUG: Testing SAM HELP command ===")

	// Step 1: Create SAM connection
	fmt.Println("DEBUG: Creating SAM connection...")
	sam, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("DEBUG: Failed to create SAM: %v\n", err)
		t.Fatal(err)
	}
	defer sam.Close()
	fmt.Println("DEBUG: SAM connection created successfully")

	// Step 2: Send HELP command to understand what commands are available
	fmt.Println("DEBUG: Sending HELP command...")

	// Access the underlying connection through the SAM struct
	_, err = sam.Write([]byte("HELP\n"))
	if err != nil {
		fmt.Printf("DEBUG: Failed to send HELP: %v\n", err)
		t.Fatal(err)
	}

	// Read response
	buffer := make([]byte, 4096)
	n, err := sam.Read(buffer)
	if err != nil {
		fmt.Printf("DEBUG: Failed to read response: %v\n", err)
		t.Fatal(err)
	}

	response := string(buffer[:n])
	fmt.Printf("DEBUG: SAM HELP response:\n%s\n", response)
}
