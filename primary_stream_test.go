package sam3

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
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
	fmt.Println("\tNotice: This may fail if your I2P node is not well integrated in the I2P network.")
	fmt.Println("\tLooking up i2p-projekt.i2p")
	forumAddr, err := earlysam.Lookup("i2p-projekt.i2p")
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

func Test_PrimaryStreamingServerClient(t *testing.T) {
	if testing.Short() {
		return
	}

	fmt.Println("Test_StreamingServerClient")
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
	fmt.Println("\tServer: Creating tunnel")
	ss, err := sam.NewUniqueStreamSubSession("PrimaryServerClientTunnel")
	if err != nil {
		return
	}
	defer ss.Close()
	time.Sleep(time.Second * 10)
	c, w := make(chan bool), make(chan bool)
	go func(c, w chan (bool)) {
		if !(<-w) {
			return
		}
		/*
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
		*/

		fmt.Println("\tClient: Creating tunnel")
		ss2, err := sam.NewStreamSubSession("primaryExampleClientTun", nil)
		if err != nil {
			c <- false
			return
		}
		defer ss2.Close()
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

type exitHandler struct{}

func (e *exitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}
