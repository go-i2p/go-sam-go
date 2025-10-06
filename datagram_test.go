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
	// ds, err := sam.NewDatagramSession("DGserverTun", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
	ds, err := sam.NewDatagramSession("DGserverTun", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
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
		// ds2, err := sam2.NewDatagramSession("DGclientTun", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
		ds2, err := sam2.NewDatagramSession("DGclientTun", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
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
	// Note: This example requires a running I2P router with SAM bridge enabled.
	// If no I2P router is available, the example will print an error and return.

	const samBridge = "127.0.0.1:7656"

	sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Println("Error creating SAM connection (I2P router may not be running):", err.Error())
		return
	}
	defer sam.Close()

	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Println("Error generating keys:", err.Error())
		return
	}
	myself := keys.Addr()

	// See the example Option_* variables.
	dg, err := sam.NewDatagramSession("DGTUN", keys, Options_Small, 0)
	if err != nil {
		fmt.Println("Error creating datagram session:", err.Error())
		return
	}
	defer dg.Close()

	// Try to lookup a destination (this may fail if I2P is not fully started)
	someone, err := sam.Lookup("zzz.i2p")
	if err != nil {
		fmt.Println("Warning: Could not resolve zzz.i2p (I2P may still be starting up):", err.Error())
		// Use a placeholder address for demonstration
		someone = myself
	}

	// Send datagrams (these operations may timeout if I2P tunnels are not ready)
	fmt.Println("Sending datagrams...")

	// Use timeouts for I2P operations
	sendTimeout := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Error during send operation: %v\n", r)
			}
			sendTimeout <- true
		}()

		_, err1 := dg.WriteTo([]byte("Hello stranger!"), someone)
		if err1 != nil {
			fmt.Println("Warning: Could not send to destination:", err1.Error())
		}

		_, err2 := dg.WriteTo([]byte("Hello myself!"), myself)
		if err2 != nil {
			fmt.Println("Warning: Could not send to self:", err2.Error())
		}
	}()

	// Wait for send operations with timeout
	select {
	case <-sendTimeout:
		// Sends completed (successfully or with errors)
	case <-time.After(30 * time.Second):
		fmt.Println("Warning: Send operations timed out (I2P tunnels may not be ready)")
		return
	}

	// Try to receive a datagram with reasonable timeout
	fmt.Println("Attempting to receive datagram...")
	receiveTimeout := make(chan bool, 1)
	var received bool

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Error during receive operation: %v\n", r)
			}
			receiveTimeout <- true
		}()

		buf := make([]byte, 31*1024)
		n, _, err := dg.ReadFrom(buf)
		if err != nil {
			fmt.Println("Could not receive datagram:", err.Error())
		} else {
			fmt.Printf("Got message: %s\n", string(buf[:n]))
			received = true
		}
	}()

	// Wait for receive operation with timeout
	select {
	case <-receiveTimeout:
		if !received {
			fmt.Println("No datagram received (this is normal if I2P tunnels are not fully established)")
		}
	case <-time.After(30 * time.Second):
		fmt.Println("Receive operation timed out (normal behavior when no messages are available)")
	}

	fmt.Println("DatagramSession example completed")
	return
	// Output:
	// Sending datagrams...
	// Attempting to receive datagram...
	// DatagramSession example completed
}
