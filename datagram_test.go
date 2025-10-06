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
				_, err = ds2.WriteTo([]byte("Hello datagram-world! <3 <3 <3 <3 <3 <3"), ds.LocalAddr())
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
	// This example demonstrates I2P datagram communication by sending messages
	// and attempting to receive them through the I2P network.
	//
	// Requirements: This example requires a running I2P router with SAM bridge enabled.

	const samBridge = "127.0.0.1:7656"

	sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Printf("Failed to connect to I2P SAM bridge at %s: %v\n", samBridge, err)
		return
	}

	keys, err := sam.NewKeys()
	if err != nil {
		sam.Close()
		fmt.Printf("Failed to generate I2P keys: %v\n", err)
		return
	}
	myself := keys.Addr()

	// Create datagram session with small tunnel configuration for faster setup
	dg, err := sam.NewDatagramSession("DGTUN", keys, Options_Small, 0)
	if err != nil {
		sam.Close()
		fmt.Printf("Failed to create datagram session: %v\n", err)
		return
	}

	// For this example, we'll send messages to ourselves to demonstrate functionality
	fmt.Println("Sending datagrams...")

	// Send a message to ourselves (this should work if I2P tunnels are established)
	_, err = dg.WriteTo([]byte("Hello myself!"), myself)
	if err != nil {
		dg.Close()
		sam.Close()
		fmt.Printf("Failed to send datagram: %v\n", err)
		return
	}

	fmt.Println("Attempting to receive datagram...")

	// Set a reasonable deadline for the receive operation (I2P operations need time)
	dg.SetReadDeadline(time.Now().Add(60 * time.Second))

	buf := make([]byte, 31*1024)
	n, addr, err := dg.ReadFrom(buf)
	if err != nil {
		dg.Close()
		sam.Close()
		fmt.Printf("Failed to receive datagram within 60 seconds: %v\n", err)
		return
	}

	fmt.Printf("Received message: %s\n", string(buf[:n]))
	fmt.Printf("From address: %s\n", addr.String())

	// Clean up resources
	dg.Close()
	sam.Close()

	fmt.Println("DatagramSession example completed")

	// Output:
	// Sending datagrams...
	// Attempting to receive datagram...
	// Received message: Hello myself!
	// From address: [I2P address]
	// DatagramSession example completed
}
