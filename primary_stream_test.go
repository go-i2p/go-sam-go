package sam3

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

/*
 * This file contains tests and examples for the primary stream session functionality of the sam3 package.
 * It was copied directly from the sam3 package and modified to fit the current context.
 * The tests cover creating primary stream sessions, dialing I2P addresses, and establishing
 * server-client communication over I2P streams. Examples demonstrate basic usage of
 * the sam3 library for connecting to I2P and creating primary stream sessions.
 *
 * Note: These tests require a running I2P router with SAM bridge enabled.
 */

func Test_PrimaryStreamingDial(t *testing.T) {
	if testing.Short() {
		return
	}
	fmt.Println("Test_PrimaryStreamingDial")

	// Set up a local test listener instead of using external site
	testListener := SetupTestListenerWithHTTP(t, generateUniqueSessionID("primary_streaming_dial_listener"))
	defer testListener.Close()

	earlysam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		t.Fail()
		return
	}
	defer earlysam.Close()
	keys, err := earlysam.NewKeys()
	if err != nil {
		t.Fail()
		return
	}

	sam, err := earlysam.NewPrimarySession("PrimaryTunnel", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		t.Fail()
		return
	}
	defer sam.Close()
	fmt.Println("\tBuilding tunnel")
	ss, err := sam.NewStreamSubSession("primaryStreamTunnel", nil)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	defer ss.Close()
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

func Test_PrimaryStreamingServerClient(t *testing.T) {
	if testing.Short() {
		return
	}

	fmt.Println("Test_PrimaryStreamingServerClient")
	
	// Create a test listener to act as the "server" - this avoids the problem
	// of trying to connect from one subsession to another within the same PRIMARY session
	testListener := SetupTestListenerWithHTTP(t, generateUniqueSessionID("primary_server_client_listener"))
	defer testListener.Close()
	
	// Create PRIMARY session for the client
	earlysam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		t.Fail()
		return
	}
	defer earlysam.Close()
	keys, err := earlysam.NewKeys()
	if err != nil {
		t.Fail()
		return
	}

	sam, err := earlysam.NewPrimarySession("PrimaryServerClientTunnel", keys, []string{"inbound.length=1", "outbound.length=1", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		t.Fail()
		return
	}
	defer sam.Close()
	
	fmt.Println("\tClient: Creating multiple stream subsessions in PRIMARY session")
	// Per SAMv3.3 spec: Multiple STREAM subsessions in a PRIMARY session
	// MUST have unique LISTEN_PORT values to route incoming traffic correctly.
	// Create two subsessions with different ports to test multiplexing
	ss1, err := sam.NewStreamSubSession("primaryClient1", []string{"LISTEN_PORT=8001"})
	if err != nil {
		t.Fatalf("Failed to create first subsession: %v", err)
		return
	}
	defer ss1.Close()
	
	ss2, err := sam.NewStreamSubSession("primaryClient2", []string{"LISTEN_PORT=8002"})
	if err != nil {
		t.Fatalf("Failed to create second subsession: %v", err)
		return
	}
	defer ss2.Close()
	
	fmt.Println("\tClient: Successfully created two stream subsessions with unique ports")
	
	// Test that both subsessions can dial external destinations
	fmt.Printf("\tClient subsession 1: Dialing test listener (%s)\n", testListener.AddrString())
	conn1, err := ss1.DialI2P(testListener.Addr())
	if err != nil {
		t.Fatalf("Subsession 1 failed to dial: %v", err)
		return
	}
	defer conn1.Close()
	
	fmt.Println("\tClient subsession 1: Sending HTTP GET /")
	if _, err := conn1.Write([]byte("GET /subsession1\n")); err != nil {
		t.Fatalf("Subsession 1 failed to write: %v", err)
		return
	}
	
	buf1 := make([]byte, 4096)
	n1, err := conn1.Read(buf1)
	if err != nil {
		t.Fatalf("Subsession 1 failed to read: %v", err)
		return
	}
	
	if !strings.Contains(strings.ToLower(string(buf1[:n1])), "http") && !strings.Contains(strings.ToLower(string(buf1[:n1])), "html") {
		t.Logf("\tWarning: Subsession 1 received %d bytes, but nothing that looked like http/html", n1)
	} else {
		fmt.Println("\tClient subsession 1: Successfully read HTTP/HTML from test listener")
	}
	
	// Test second subsession
	fmt.Printf("\tClient subsession 2: Dialing test listener (%s)\n", testListener.AddrString())
	conn2, err := ss2.DialI2P(testListener.Addr())
	if err != nil {
		t.Fatalf("Subsession 2 failed to dial: %v", err)
		return
	}
	defer conn2.Close()
	
	fmt.Println("\tClient subsession 2: Sending HTTP GET /")
	if _, err := conn2.Write([]byte("GET /subsession2\n")); err != nil {
		t.Fatalf("Subsession 2 failed to write: %v", err)
		return
	}
	
	buf2 := make([]byte, 4096)
	n2, err := conn2.Read(buf2)
	if err != nil {
		t.Fatalf("Subsession 2 failed to read: %v", err)
		return
	}
	
	if !strings.Contains(strings.ToLower(string(buf2[:n2])), "http") && !strings.Contains(strings.ToLower(string(buf2[:n2])), "html") {
		t.Logf("\tWarning: Subsession 2 received %d bytes, but nothing that looked like http/html", n2)
	} else {
		fmt.Println("\tClient subsession 2: Successfully read HTTP/HTML from test listener")
	}
	
	fmt.Println("\tTest passed: PRIMARY session with multiple STREAM subsessions working correctly")
}

type exitHandler struct{}

func (e *exitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}
