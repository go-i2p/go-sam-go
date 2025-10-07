package sam3

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

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

	// Set up a local test listener instead of using external site
	testListener := SetupTestListenerWithHTTP(t, generateUniqueSessionID("streaming_dial_listener"))
	defer testListener.Close()

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
	fmt.Println("\tNotice: Using local test listener instead of external I2P site for improved test stability.")
	fmt.Printf("\tDialing test listener (%s)\n", testListener.AddrString())
	conn, err := ss.DialI2P(testListener.Addr())
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
		fmt.Printf("\tProbably failed to StreamSession.DialI2P(test listener)? It replied %d bytes, but nothing that looked like http/html", n)
	} else {
		fmt.Println("\tRead HTTP/HTML from test listener")
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
	// Creates a new StreamingSession with a local server and client.
	// Demonstrates how a SAMConn behaves just like a normal net.Conn.
	//
	// Requirements: This example requires a running I2P router with SAM bridge enabled.

	const samBridge = "127.0.0.1:7656"

	// Create server session first
	server_sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Printf("Failed to connect to I2P SAM bridge: %v", err)
		return
	}
	defer server_sam.Close()

	server_keys, err := server_sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate server I2P keys: %v", err)
		return
	}

	server_session, err := server_sam.NewStreamSession("stream_server", server_keys, Options_Small)
	if err != nil {
		fmt.Printf("Failed to create server stream session: %v", err)
		return
	}
	defer server_session.Close()

	// Synchronization channels
	serverReady := make(chan bool)
	clientDone := make(chan bool)

	// Server goroutine - listens and responds with HTTP-like content
	go func() {
		listener, err := server_session.Listen()
		if err != nil {
			fmt.Printf("Server failed to listen: %v", err)
			serverReady <- false
			return
		}
		defer listener.Close()
		serverReady <- true

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read client request
		buf := make([]byte, 256)
		_, _ = conn.Read(buf)

		// Send HTTP-like response
		response := "HTTP/1.0 200 OK\r\nContent-Type: text/html\r\n\r\n<html><body>Test</body></html>"
		conn.Write([]byte(response))
	}()

	// Wait for server to be ready
	if !<-serverReady {
		return
	}

	// Client operations in separate goroutine
	go func() {
		client_sam, err := NewSAM(samBridge)
		if err != nil {
			fmt.Printf("Client failed to connect to I2P: %v", err)
			clientDone <- false
			return
		}
		defer client_sam.Close()

		client_keys, err := client_sam.NewKeys()
		if err != nil {
			fmt.Printf("Client failed to generate keys: %v", err)
			clientDone <- false
			return
		}

		client_session, err := client_sam.NewStreamSession("stream_client", client_keys, Options_Small)
		if err != nil {
			fmt.Printf("Client failed to create session: %v", err)
			clientDone <- false
			return
		}
		defer client_session.Close()

		// I2P tunnel establishment can take 30-120 seconds, so we retry dialing
		var conn net.Conn
		var dialErr error
		for attempt := 0; attempt < 6; attempt++ {
			conn, dialErr = client_session.DialI2P(server_keys.Addr())
			if dialErr == nil {
				break
			}
			if attempt < 5 {
				// Exponential backoff: 15s, 30s, 45s, 60s, 75s
				sleepTime := time.Duration(15*(attempt+1)) * time.Second
				fmt.Printf("Dial attempt %d failed: %v. Retrying in %v...\n", attempt+1, dialErr, sleepTime)
				time.Sleep(sleepTime)
			}
		}
		if dialErr != nil {
			fmt.Printf("Client failed to dial server after retries: %v", dialErr)
			clientDone <- false
			return
		}
		defer conn.Close()

		fmt.Println("Sending HTTP GET /")
		if _, err := conn.Write([]byte("GET /\n")); err != nil {
			fmt.Printf("Failed to write to connection: %v", err)
			clientDone <- false
			return
		}

		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Failed to read from connection: %v", err)
			clientDone <- false
			return
		}

		response := string(buf[:n])
		if strings.Contains(strings.ToLower(response), "http") || strings.Contains(strings.ToLower(response), "html") {
			fmt.Println("Read HTTP/HTML from test destination")
		} else {
			fmt.Printf("Failed to get HTTP/HTML response (got %d bytes)", n)
		}
		clientDone <- true
	}()

	// Wait for client to complete
	<-clientDone

	// Output:
	// Sending HTTP GET /
	// Read HTTP/HTML from test destination
}

func ExampleStreamListener() {
	// One server Accept()ing on a StreamListener, and one client that Dials
	// through I2P to the server. Server writes "Hello world!" through a SAMConn
	// (which implements net.Conn) and the client prints the message.
	//
	// Requirements: This example requires a running I2P router with SAM bridge enabled.

	const samBridge = "127.0.0.1:7656"

	server_sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Printf("Failed to connect to I2P SAM bridge: %v", err)
		return
	}
	defer server_sam.Close()
	keys, err := server_sam.NewKeys()
	if err != nil {
		fmt.Printf("Failed to generate I2P keys: %v", err)
		return
	}

	// Create server session BEFORE starting client goroutine
	server_session, err := server_sam.NewStreamSession("server_example", keys, Options_Small)
	if err != nil {
		fmt.Printf("Failed to create server session: %v", err)
		return
	}

	// Channels for synchronization and results
	quit := make(chan bool)
	serverReady := make(chan bool)

	// Client connecting to the server - waits for server readiness signal
	go func(server i2pkeys.I2PAddr) {
		// Wait for server to be fully ready (session + listener created)
		if !<-serverReady {
			quit <- false
			return
		}

		client_sam, err := NewSAM(samBridge)
		if err != nil {
			fmt.Printf("Client failed to connect to I2P: %v", err)
			quit <- false
			return
		}
		defer client_sam.Close()
		keys, err := client_sam.NewKeys()
		if err != nil {
			fmt.Printf("Client failed to generate keys: %v", err)
			quit <- false
			return
		}
		client_session, err := client_sam.NewStreamSession("client_example", keys, Options_Small)
		if err != nil {
			fmt.Printf("Client failed to create session: %v", err)
			quit <- false
			return
		}
		client_conn, err := client_session.DialI2P(server)
		if err != nil {
			fmt.Printf("Client failed to dial server: %v", err)
			quit <- false
			return
		}
		buf := make([]byte, 256)
		n, err := client_conn.Read(buf)
		if err != nil {
			fmt.Printf("Client failed to read: %v", err)
			quit <- false
			return
		}
		fmt.Println(string(buf[:n]))
		quit <- true
	}(keys.Addr()) // end of client - pass server address

	// Create listener and signal client can proceed
	l, err := server_session.Listen()
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
		serverReady <- false // Signal failure to client
		return
	}
	defer l.Close()
	serverReady <- true // Signal success to client

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
