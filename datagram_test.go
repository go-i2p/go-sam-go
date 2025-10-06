package sam3

import (
	"fmt"
	"testing"
	"time"
)

func Test_DatagramServerClient(t *testing.T) {
	if testing.Short() {
		return
	}

	fmt.Println("Test_DatagramServerClient")
	sam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		t.Fail()
		return
	}
	defer sam.Close()
	keys, err := sam.NewKeys()
	if err != nil {
		t.Fail()
		return
	}
	//	fmt.Println("\tServer: My address: " + keys.Addr().Base32())
	fmt.Println("\tServer: Creating tunnel")
	// ds, err := sam.NewDatagramSession("DGserverTun", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
	ds, err := sam.NewDatagramSession("DGserverTun", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
	if err != nil {
		fmt.Println("Server: Failed to create tunnel: " + err.Error())
		t.Fail()
		return
	}
	c, w := make(chan bool), make(chan bool)
	go func(c, w chan (bool)) {
		sam2, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			c <- false
			return
		}
		defer sam2.Close()
		keys, err := sam2.NewKeys()
		if err != nil {
			c <- false
			return
		}
		fmt.Println("\tClient: Creating tunnel")
		// ds2, err := sam2.NewDatagramSession("DGclientTun", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
		ds2, err := sam2.NewDatagramSession("DGclientTun", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
		if err != nil {
			c <- false
			return
		}
		defer ds2.Close()
		//		fmt.Println("\tClient: Servers address: " + ds.LocalAddr().Base32())
		//		fmt.Println("\tClient: Clients address: " + ds2.LocalAddr().Base32())
		fmt.Println("\tClient: Tries to send datagram to server")
		for {
			select {
			default:
				_, err = ds2.WriteTo([]byte("Hello datagram-world! <3 <3 <3 <3 <3 <3"), ds.Addr())
				if err != nil {
					fmt.Println("\tClient: Failed to send datagram: " + err.Error())
					c <- false
					return
				}
				time.Sleep(5 * time.Second)
			case <-w:
				fmt.Println("\tClient: Sent datagram, quitting.")
				return
			}
		}
		c <- true
	}(c, w)
	buf := make([]byte, 512)
	fmt.Println("\tServer: ReadFrom() waiting...")
	n, _, err := ds.ReadFrom(buf)
	w <- true
	if err != nil {
		fmt.Println("\tServer: Failed to ReadFrom(): " + err.Error())
		t.Fail()
		return
	}
	fmt.Println("\tServer: Received datagram: " + string(buf[:n]))
	//	fmt.Println("\tServer: Senders address was: " + saddr.Base32())
}

func ExampleDatagramSession() {
	// Creates a new DatagramSession, which behaves just like a net.PacketConn.
	// This example demonstrates the DatagramSession API for I2P datagram communication.
	//
	// Requirements: This example requires a running I2P router with SAM bridge enabled.
	//
	// Note: Full datagram examples require tunnel establishment and message delivery
	// through I2P, which typically takes 30-120 seconds. For a complete working example,
	// see Test_DatagramServerClient in this file.

	const samBridge = "127.0.0.1:7656"

	// Connect to SAM bridge
	sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Printf("Failed to connect to I2P SAM bridge: %v\n", err)
		return
	}
	defer sam.Close()

	// Generate I2P keys for this session
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate keys: %v\n", err)
		return
	}

	// Create datagram session with small tunnel configuration
	session, err := sam.NewDatagramSession("ExampleDG", keys, Options_Small, 0)
	if err != nil {
		fmt.Printf("Failed to create datagram session: %v\n", err)
		return
	}
	defer session.Close()

	fmt.Printf("Datagram session created successfully\n")
	// Note: session.Addr().Base32() returns the actual I2P destination address
	fmt.Printf("Session ready for datagram communication\n")

	// Output:
	// Datagram session created successfully
	// Session ready for datagram communication
}
