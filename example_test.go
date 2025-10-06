package sam3_test

import (
	"fmt"
	"log"

	sam3 "github.com/go-i2p/go-sam-go"
)

// Example demonstrates basic usage of the sam3 library for I2P connectivity.
// This example shows how to establish a SAM connection, generate keys, and create sessions.
//
// Requirements: This example requires a running I2P router with SAM bridge enabled.
func Example() {
	// Connect to the local I2P SAM bridge
	sam, err := sam3.NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("Cannot connect to I2P: %v", err)
		return
	}
	defer sam.Close()

	// Generate I2P keys for this session
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate keys: %v", err)
		return
	}

	// Create a stream session for TCP-like connections
	session, err := sam.NewStreamSession("example-session", keys, sam3.Options_Default)
	if err != nil {
		fmt.Printf("Failed to create session: %v", err)
		return
	}
	defer session.Close()

	fmt.Println("Successfully connected to I2P and created a session")
	// Output: Successfully connected to I2P and created a session
}

// ExampleNewSAM demonstrates how to establish a connection to the I2P SAM bridge.
//
// Requirements: This example requires a running I2P router with SAM bridge enabled.
func ExampleNewSAM() {
	// Connect to the default I2P SAM bridge address
	sam, err := sam3.NewSAM(sam3.SAMDefaultAddr(""))
	if err != nil {
		fmt.Printf("Cannot connect to I2P: %v", err)
		return
	}
	defer sam.Close()

	fmt.Println("Connected to I2P SAM bridge")
}

// ExampleSAM_NewStreamSession demonstrates creating a stream session for reliable connections.
//
// Requirements: This example requires a running I2P router with SAM bridge enabled.
func ExampleSAM_NewStreamSession() {
	sam, err := sam3.NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("Cannot connect to I2P: %v", err)
		return
	}
	defer sam.Close()

	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate keys: %v", err)
		return
	}

	// Create a stream session with default tunnel configuration
	session, err := sam.NewStreamSession("my-app", keys, sam3.Options_Default)
	if err != nil {
		fmt.Printf("Failed to create stream session: %v", err)
		return
	}
	defer session.Close()

	fmt.Println("Stream session created successfully")
}

// ExampleSAM_NewPrimarySession demonstrates creating a primary session for managing sub-sessions.
//
// Requirements: This example requires a running I2P router with SAM bridge enabled.
func ExampleSAM_NewPrimarySession() {
	sam, err := sam3.NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("Cannot connect to I2P: %v", err)
		return
	}
	defer sam.Close()

	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate keys: %v", err)
		return
	}

	// Create a primary session that can manage multiple sub-sessions
	primary, err := sam.NewPrimarySession("master-session", keys, sam3.Options_Medium)
	if err != nil {
		fmt.Printf("Failed to create primary session: %v", err)
		return
	}
	defer primary.Close()

	fmt.Printf("Primary session created with %d sub-sessions", primary.SubSessionCount())
	// Output: Primary session created with 0 sub-sessions
}

// ExampleSAM_NewDatagramSession demonstrates creating a datagram session for UDP-like messaging.
//
// Requirements: This example requires a running I2P router with SAM bridge enabled.
func ExampleSAM_NewDatagramSession() {
	sam, err := sam3.NewSAM("127.0.0.1:7656")
	if err != nil {
		fmt.Printf("Cannot connect to I2P: %v", err)
		return
	}
	defer sam.Close()

	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate keys: %v", err)
		return
	}

	// Create a datagram session for authenticated messaging
	session, err := sam.NewDatagramSession("udp-app", keys, sam3.Options_Small, 0)
	if err != nil {
		fmt.Printf("Failed to create datagram session: %v", err)
		return
	}
	defer session.Close()

	fmt.Println("Datagram session created successfully")
}

// ExampleOptions demonstrates using predefined tunnel configuration options.
func ExampleOptions() {
	// Use predefined options for different traffic patterns

	// For applications with heavy traffic
	heavyTrafficOptions := sam3.Options_Large

	// For applications with medium traffic (most common)
	normalOptions := sam3.Options_Default

	// For lightweight applications
	lightOptions := sam3.Options_Small

	// For maximum anonymity with very heavy traffic
	maxAnonOptions := sam3.Options_Humongous

	fmt.Printf("Heavy: %d options, Normal: %d options, Light: %d options, Max Anon: %d options",
		len(heavyTrafficOptions), len(normalOptions), len(lightOptions), len(maxAnonOptions))
	// Output: Heavy: 8 options, Normal: 8 options, Light: 8 options, Max Anon: 8 options
}

// ExampleRandString demonstrates generating random session identifiers.
func ExampleRandString() {
	// Generate random strings for session IDs
	sessionID := sam3.RandString()
	fmt.Printf("Generated session ID length: %d", len(sessionID))
	// Output: Generated session ID length: 12
}

// ExampleSAMDefaultAddr demonstrates using the default SAM address with environment variable support.
func ExampleSAMDefaultAddr() {
	// Get default SAM address (uses environment variables if set)
	defaultAddr := sam3.SAMDefaultAddr("")
	fmt.Printf("Default SAM address: %s", defaultAddr)
	// Output: Default SAM address: 127.0.0.1:7656
}

// ExampleExtractDest demonstrates extracting destinations from SAM protocol strings.
func ExampleExtractDest() {
	// Extract the first word (destination) from a SAM response
	response := "abc123.b32.i2p RESULT=OK VERSION=3.3"
	dest := sam3.ExtractDest(response)
	fmt.Printf("Extracted destination: %s", dest)
	// Output: Extracted destination: abc123.b32.i2p
}

// ExampleExtractPairString demonstrates extracting string values from SAM protocol responses.
func ExampleExtractPairString() {
	// Extract specific parameters from SAM responses
	response := "RESULT=OK MESSAGE=Connected VERSION=3.3"
	result := sam3.ExtractPairString(response, "RESULT")
	message := sam3.ExtractPairString(response, "MESSAGE")

	fmt.Printf("Result: %s, Message: %s", result, message)
	// Output: Result: OK, Message: Connected
}

// ExampleExtractPairInt demonstrates extracting integer values from SAM protocol responses.
func ExampleExtractPairInt() {
	// Extract numeric parameters from SAM responses
	response := "RESULT=OK PORT=7656 COUNT=5"
	port := sam3.ExtractPairInt(response, "PORT")
	count := sam3.ExtractPairInt(response, "COUNT")

	fmt.Printf("Port: %d, Count: %d", port, count)
	// Output: Port: 7656, Count: 5
}

// ExampleSetType demonstrates configuring session types using the functional options pattern.
func ExampleSetType() {
	// Create a new SAM configuration
	emit, err := sam3.NewEmit(
		sam3.SetType("STREAM"),
		sam3.SetSAMHost("127.0.0.1"),
		sam3.SetSAMPort("7656"),
	)
	if err != nil {
		log.Printf("Configuration failed: %v", err)
		return
	}

	fmt.Printf("Configured session type: %s", emit.Style)
	// Output: Configured session type: STREAM
}

// ExampleNewEmit demonstrates creating SAM configuration with functional options.
func ExampleNewEmit() {
	// Create configuration with multiple options
	config, err := sam3.NewEmit(
		sam3.SetType("DATAGRAM"),
		sam3.SetInLength(2),
		sam3.SetOutLength(2),
		sam3.SetInQuantity(3),
		sam3.SetOutQuantity(3),
	)
	if err != nil {
		log.Printf("Configuration failed: %v", err)
		return
	}

	fmt.Printf("Session type: %s, Tunnels: in=%d out=%d",
		config.Style, config.I2PConfig.InQuantity, config.I2PConfig.OutQuantity)
	// Output: Session type: DATAGRAM, Tunnels: in=3 out=3
}
