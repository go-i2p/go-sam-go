package sam3

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/go-i2p/i2pkeys"
)

/*
 * This file contains tests and examples for the stream session functionality of the sam3 package.
 * It was copied directly from the sam3 package and modified to fit the current context.
 * The tests cover creating stream sessions, dialing I2P addresses, and establishing
 * server-client communication over I2P streams. Examples demonstrate basic usage of
 * the sam3 library for connecting to I2P and creating stream sessions.
 *
 * Note: These tests require a running I2P router with SAM bridge enabled.
 */

func Test_StreamingDial(t *testing.T) {
	if testing.Short() {
		return
	}
	fmt.Println("Test_StreamingDial")
	sam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	defer sam.Close()
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	fmt.Println("\tBuilding tunnel")
	ss, err := sam.NewStreamSession("streamTun", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	fmt.Println("\tNotice: This may fail if your I2P node is not well integrated in the I2P network.")
	fmt.Println("\tLooking up i2p-projekt.i2p")
	forumAddr, err := sam.Lookup("i2p-projekt.i2p")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	fmt.Println("\tDialing i2p-projekt.i2p(", forumAddr.Base32(), forumAddr.DestHash().Hash(), ")")
	conn, err := ss.DialI2P(forumAddr)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	defer conn.Close()
	fmt.Println("\tSending HTTP GET /")
	if _, err := conn.Write([]byte("GET /\n")); err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if !strings.Contains(strings.ToLower(string(buf[:n])), "http") && !strings.Contains(strings.ToLower(string(buf[:n])), "html") {
		fmt.Printf("\tProbably failed to StreamSession.DialI2P(i2p-projekt.i2p)? It replied %d bytes, but nothing that looked like http/html", n)
	} else {
		fmt.Println("\tRead HTTP/HTML from i2p-projekt.i2p")
	}
}

func Test_StreamingServerClient(t *testing.T) {
	if testing.Short() {
		return
	}

	fmt.Println("Test_StreamingServerClient")
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
	fmt.Println("\tServer: Creating tunnel")
	ss, err := sam.NewStreamSession("serverTun", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		return
	}
	c, w := make(chan bool), make(chan bool)
	go func(c, w chan (bool)) {
		if !(<-w) {
			return
		}
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
		ss2, err := sam2.NewStreamSession("clientTun", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
		if err != nil {
			c <- false
			return
		}
		fmt.Println("\tClient: Connecting to server")
		conn, err := ss2.DialI2P(ss.Addr())
		if err != nil {
			c <- false
			return
		}
		fmt.Println("\tClient: Connected to tunnel")
		defer conn.Close()
		_, err = conn.Write([]byte("Hello world <3 <3 <3 <3 <3 <3"))
		if err != nil {
			c <- false
			return
		}
		c <- true
	}(c, w)
	l, err := ss.Listen()
	if err != nil {
		fmt.Println("ss.Listen(): " + err.Error())
		t.Fail()
		w <- false
		return
	}
	defer l.Close()
	w <- true
	fmt.Println("\tServer: Accept()ing on tunnel")
	conn, err := l.Accept()
	if err != nil {
		t.Fail()
		fmt.Println("Failed to Accept(): " + err.Error())
		return
	}
	defer conn.Close()
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	fmt.Printf("\tClient exited successfully: %t\n", <-c)
	fmt.Println("\tServer: received from Client: " + string(buf[:n]))
}

func ExampleStreamSession() {
	// Creates a new StreamingSession, dials to idk.i2p and gets a SAMConn
	// which behaves just like a normal net.Conn.
	//
	// Requirements: This example requires a running I2P router with SAM bridge enabled.

	const samBridge = "127.0.0.1:7656"

	sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Printf("Failed to connect to I2P SAM bridge: %v", err)
		return
	}
	defer sam.Close()
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate I2P keys: %v", err)
		return
	}
	// See the example Option_* variables.
	ss, err := sam.NewStreamSession("stream_example", keys, Options_Small)
	if err != nil {
		fmt.Printf("Failed to create stream session: %v", err)
		return
	}
	someone, err := sam.Lookup("idk.i2p")
	if err != nil {
		fmt.Printf("Failed to lookup idk.i2p: %v", err)
		return
	}

	conn, err := ss.DialI2P(someone)
	if err != nil {
		fmt.Printf("Failed to dial idk.i2p: %v", err)
		return
	}
	defer conn.Close()
	fmt.Println("Sending HTTP GET /")
	if _, err := conn.Write([]byte("GET /\n")); err != nil {
		fmt.Printf("Failed to write to connection: %v", err)
		return
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Failed to read from connection: %v", err)
		return
	}
	if !strings.Contains(strings.ToLower(string(buf[:n])), "http") && !strings.Contains(strings.ToLower(string(buf[:n])), "html") {
		fmt.Printf("Failed to get HTTP/HTML response from idk.i2p (got %d bytes)", n)
		return
	} else {
		fmt.Println("Read HTTP/HTML from idk.i2p")
		log.Println("Read HTTP/HTML from idk.i2p")
	}
	return

	// Output:
	// Sending HTTP GET /
	// Read HTTP/HTML from idk.i2p
}

func ExampleStreamListener() {
	// One server Accept()ing on a StreamListener, and one client that Dials
	// through I2P to the server. Server writes "Hello world!" through a SAMConn
	// (which implements net.Conn) and the client prints the message.
	//
	// Requirements: This example requires a running I2P router with SAM bridge enabled.

	const samBridge = "127.0.0.1:7656"

	sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Printf("Failed to connect to I2P SAM bridge: %v", err)
		return
	}
	defer sam.Close()
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate I2P keys: %v", err)
		return
	}

	quit := make(chan bool)

	// Client connecting to the server
	go func(server i2pkeys.I2PAddr) {
		csam, err := NewSAM(samBridge)
		if err != nil {
			fmt.Printf("Client failed to connect to I2P: %v", err)
			quit <- false
			return
		}
		defer csam.Close()
		keys, err := csam.NewKeys()
		if err != nil {
			fmt.Printf("Client failed to generate keys: %v", err)
			quit <- false
			return
		}
		cs, err := csam.NewStreamSession("client_example", keys, Options_Small)
		if err != nil {
			fmt.Printf("Client failed to create session: %v", err)
			quit <- false
			return
		}
		conn, err := cs.DialI2P(server)
		if err != nil {
			fmt.Printf("Client failed to dial server: %v", err)
			quit <- false
			return
		}
		buf := make([]byte, 256)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Client failed to read: %v", err)
			quit <- false
			return
		}
		fmt.Println(string(buf[:n]))
		quit <- true
	}(keys.Addr()) // end of client

	ss, err := sam.NewStreamSession("server_example", keys, Options_Small)
	if err != nil {
		fmt.Printf("Failed to create server session: %v", err)
		return
	}
	l, err := ss.Listen()
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
		return
	}
	conn, err := l.Accept()
	if err != nil {
		fmt.Printf("Failed to accept connection: %v", err)
		return
	}
	_, err = conn.Write([]byte("Hello world!"))
	if err != nil {
		fmt.Printf("Failed to write to client: %v", err)
		return
	}

	success := <-quit // waits for client to complete
	if !success {
		fmt.Printf("Client operation failed")
		return
	}

	// Output:
	// Hello world!
}
