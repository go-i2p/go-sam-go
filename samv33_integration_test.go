//go:build integration

package sam3

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-i2p/go-sam-go/datagram2"
	"github.com/go-i2p/go-sam-go/datagram3"
)

var integrationTunnelOptions = []string{
	"inbound.length=1",
	"outbound.length=1",
	"inbound.lengthVariance=0",
	"outbound.lengthVariance=0",
	"inbound.quantity=1",
	"outbound.quantity=1",
}

func integrationSAMAddr() string {
	if addr := strings.TrimSpace(os.Getenv("SAM_BRIDGE_ADDR")); addr != "" {
		return addr
	}
	return "127.0.0.1:7656"
}

func ensureBridgeReady(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", integrationSAMAddr(), 2*time.Second)
		if err == nil {
			_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
			_, _ = io.WriteString(conn, "HELLO VERSION MIN=3.0 MAX=3.3\n")
			buf := make([]byte, 1024)
			n, _ := conn.Read(buf)
			_ = conn.Close()
			if strings.Contains(string(buf[:n]), "HELLO REPLY RESULT=OK") {
				SAM_HOST = strings.Split(integrationSAMAddr(), ":")[0]
				SAM_PORT = strings.Split(integrationSAMAddr(), ":")[1]
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("SAM bridge did not become ready at %s", integrationSAMAddr())
}

func uniqueID(prefix string) string {
	return prefix + "-" + RandString()
}

func receiveWithTimeout[T any](timeout time.Duration, fn func() (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := fn()
		ch <- result{value: v, err: err}
	}()

	select {
	case out := <-ch:
		return out.value, out.err
	case <-time.After(timeout):
		var zero T
		return zero, fmt.Errorf("timeout after %v", timeout)
	}
}

func TestSAMv33_ControlPlaneE2E(t *testing.T) {
	ensureBridgeReady(t)

	conn, err := net.DialTimeout("tcp", integrationSAMAddr(), 10*time.Second)
	if err != nil {
		t.Fatalf("failed to connect to SAM bridge: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	mustWrite := func(cmd string) {
		t.Helper()
		if _, err := io.WriteString(conn, cmd); err != nil {
			t.Fatalf("failed to write %q: %v", strings.TrimSpace(cmd), err)
		}
	}

	mustReadLine := func() string {
		t.Helper()
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed to read SAM reply line: %v", err)
		}
		return line
	}

	mustWrite("HELLO VERSION MIN=3.0 MAX=3.3\n")
	hello := mustReadLine()
	if !strings.Contains(hello, "HELLO REPLY RESULT=OK") {
		t.Fatalf("unexpected HELLO reply: %q", hello)
	}

	mustWrite("PING 42\n")
	ping := mustReadLine()
	if !strings.Contains(ping, "PONG 42") {
		t.Fatalf("unexpected PING reply: %q", ping)
	}

	mustWrite("DEST GENERATE SIGNATURE_TYPE=EdDSA_SHA512_Ed25519\n")
	destReply := mustReadLine()
	if !strings.Contains(destReply, "PUB=") || !strings.Contains(destReply, "PRIV=") {
		t.Fatalf("unexpected DEST GENERATE reply: %q", destReply)
	}

	mustWrite("HELP\n")
	helpLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read HELP output: %v", err)
	}
	if !strings.Contains(strings.ToUpper(helpLine), "HELP") && !strings.Contains(strings.ToUpper(helpLine), "SAM") {
		t.Fatalf("unexpected HELP output: %q", helpLine)
	}

	mustWrite("QUIT\n")
}

func TestSAMv33_StreamAndNamingE2E(t *testing.T) {
	ensureBridgeReady(t)

	serverSAM, err := NewSAM(integrationSAMAddr())
	if err != nil {
		t.Fatalf("failed to create server SAM: %v", err)
	}
	defer serverSAM.Close()

	clientSAM, err := NewSAM(integrationSAMAddr())
	if err != nil {
		t.Fatalf("failed to create client SAM: %v", err)
	}
	defer clientSAM.Close()

	serverKeys, err := serverSAM.NewKeys()
	if err != nil {
		t.Fatalf("failed to create server keys: %v", err)
	}
	clientKeys, err := clientSAM.NewKeys()
	if err != nil {
		t.Fatalf("failed to create client keys: %v", err)
	}

	serverSession, err := serverSAM.NewStreamSession(uniqueID("stream-server"), serverKeys, integrationTunnelOptions)
	if err != nil {
		t.Fatalf("failed to create server stream session: %v", err)
	}
	defer serverSession.Close()

	clientSession, err := clientSAM.NewStreamSession(uniqueID("stream-client"), clientKeys, integrationTunnelOptions)
	if err != nil {
		t.Fatalf("failed to create client stream session: %v", err)
	}
	defer clientSession.Close()

	resolved, err := clientSAM.Resolve(serverKeys.Addr().Base32())
	if err != nil {
		t.Fatalf("failed NAMING LOOKUP resolve: %v", err)
	}
	if resolved == "" {
		t.Fatal("resolved address is empty")
	}

	listener, err := serverSession.Listen()
	if err != nil {
		t.Fatalf("failed to listen on stream session: %v", err)
	}
	defer listener.Close()

	serverReceived := make(chan []byte, 1)
	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		buf := make([]byte, 256)
		n, err := conn.Read(buf)
		if err != nil {
			serverErr <- err
			return
		}
		serverReceived <- append([]byte(nil), buf[:n]...)

		_, err = conn.Write([]byte("ok-from-server"))
		serverErr <- err
	}()

	clientConn, err := clientSession.DialI2P(serverSession.Addr())
	if err != nil {
		t.Fatalf("failed to dial server over STREAM: %v", err)
	}
	defer clientConn.Close()

	payload := []byte("hello-stream")
	if _, err := clientConn.Write(payload); err != nil {
		t.Fatalf("failed to write stream payload: %v", err)
	}

	ack := make([]byte, 256)
	n, err := clientConn.Read(ack)
	if err != nil {
		t.Fatalf("failed to read stream ack: %v", err)
	}
	if string(ack[:n]) != "ok-from-server" {
		t.Fatalf("unexpected stream ack: %q", string(ack[:n]))
	}

	select {
	case got := <-serverReceived:
		if !bytes.Equal(got, payload) {
			t.Fatalf("server got %q, expected %q", string(got), string(payload))
		}
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("server stream failure: %v", err)
		}
	case <-time.After(45 * time.Second):
		t.Fatal("timeout waiting for server stream receive")
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("server stream write failed: %v", err)
	}
}

func TestSAMv33_DatagramFamilyE2E(t *testing.T) {
	ensureBridgeReady(t)

	t.Run("DATAGRAM", func(t *testing.T) {
		receiverSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("receiver SAM error: %v", err)
		}
		defer receiverSAM.Close()

		senderSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("sender SAM error: %v", err)
		}
		defer senderSAM.Close()

		receiverKeys, _ := receiverSAM.NewKeys()
		senderKeys, _ := senderSAM.NewKeys()

		receiver, err := receiverSAM.NewDatagramSession(uniqueID("dg-rx"), receiverKeys, integrationTunnelOptions, 0)
		if err != nil {
			t.Fatalf("create DATAGRAM receiver failed: %v", err)
		}
		defer receiver.Close()

		sender, err := senderSAM.NewDatagramSession(uniqueID("dg-tx"), senderKeys, integrationTunnelOptions, 0)
		if err != nil {
			t.Fatalf("create DATAGRAM sender failed: %v", err)
		}
		defer sender.Close()

		payload := []byte("datagram-e2e")
		if err := sender.SendDatagram(payload, receiver.Addr()); err != nil {
			t.Fatalf("send DATAGRAM failed: %v", err)
		}

		dg, err := receiveWithTimeout(45*time.Second, receiver.ReceiveDatagram)
		if err != nil {
			t.Fatalf("receive DATAGRAM failed: %v", err)
		}
		if !bytes.Equal(dg.Data, payload) {
			t.Fatalf("DATAGRAM payload mismatch: got %q want %q", string(dg.Data), string(payload))
		}
	})

	t.Run("RAW", func(t *testing.T) {
		receiverSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("receiver SAM error: %v", err)
		}
		defer receiverSAM.Close()

		senderSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("sender SAM error: %v", err)
		}
		defer senderSAM.Close()

		receiverKeys, _ := receiverSAM.NewKeys()
		senderKeys, _ := senderSAM.NewKeys()

		rawReceiver, err := receiverSAM.NewRawSession(uniqueID("raw-rx"), receiverKeys, integrationTunnelOptions, 0)
		if err != nil {
			t.Fatalf("create RAW receiver failed: %v", err)
		}
		defer rawReceiver.Close()

		rawSender, err := senderSAM.NewRawSession(uniqueID("raw-tx"), senderKeys, integrationTunnelOptions, 0)
		if err != nil {
			t.Fatalf("create RAW sender failed: %v", err)
		}
		defer rawSender.Close()

		payload := []byte("raw-e2e")
		if err := rawSender.SendDatagram(payload, rawReceiver.Addr()); err != nil {
			t.Fatalf("send RAW failed: %v", err)
		}

		rg, err := receiveWithTimeout(45*time.Second, rawReceiver.ReceiveDatagram)
		if err != nil {
			t.Fatalf("receive RAW failed: %v", err)
		}
		if !bytes.Equal(rg.Data, payload) {
			t.Fatalf("RAW payload mismatch: got %q want %q", string(rg.Data), string(payload))
		}
	})

	t.Run("DATAGRAM2", func(t *testing.T) {
		receiverSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("receiver SAM error: %v", err)
		}
		defer receiverSAM.Close()

		senderSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("sender SAM error: %v", err)
		}
		defer senderSAM.Close()

		receiverKeys, _ := receiverSAM.NewKeys()
		senderKeys, _ := senderSAM.NewKeys()

		rx2, err := (&datagram2.SAM{SAM: receiverSAM.SAM}).NewDatagram2Session(uniqueID("dg2-rx"), receiverKeys, integrationTunnelOptions)
		if err != nil {
			t.Fatalf("create DATAGRAM2 receiver failed: %v", err)
		}
		defer rx2.Close()

		tx2, err := (&datagram2.SAM{SAM: senderSAM.SAM}).NewDatagram2Session(uniqueID("dg2-tx"), senderKeys, integrationTunnelOptions)
		if err != nil {
			t.Fatalf("create DATAGRAM2 sender failed: %v", err)
		}
		defer tx2.Close()

		payload := []byte("datagram2-e2e")
		if err := tx2.SendDatagram(payload, rx2.Addr()); err != nil {
			t.Fatalf("send DATAGRAM2 failed: %v", err)
		}

		dg2, err := receiveWithTimeout(45*time.Second, rx2.ReceiveDatagram)
		if err != nil {
			t.Fatalf("receive DATAGRAM2 failed: %v", err)
		}
		if !bytes.Equal(dg2.Data, payload) {
			t.Fatalf("DATAGRAM2 payload mismatch: got %q want %q", string(dg2.Data), string(payload))
		}
	})

	t.Run("DATAGRAM3", func(t *testing.T) {
		receiverSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("receiver SAM error: %v", err)
		}
		defer receiverSAM.Close()

		senderSAM, err := NewSAM(integrationSAMAddr())
		if err != nil {
			t.Fatalf("sender SAM error: %v", err)
		}
		defer senderSAM.Close()

		receiverKeys, _ := receiverSAM.NewKeys()
		senderKeys, _ := senderSAM.NewKeys()

		rx3, err := (&datagram3.SAM{SAM: receiverSAM.SAM}).NewDatagram3Session(uniqueID("dg3-rx"), receiverKeys, integrationTunnelOptions)
		if err != nil {
			t.Fatalf("create DATAGRAM3 receiver failed: %v", err)
		}
		defer rx3.Close()

		tx3, err := (&datagram3.SAM{SAM: senderSAM.SAM}).NewDatagram3Session(uniqueID("dg3-tx"), senderKeys, integrationTunnelOptions)
		if err != nil {
			t.Fatalf("create DATAGRAM3 sender failed: %v", err)
		}
		defer tx3.Close()

		payload := []byte("datagram3-e2e")
		if err := tx3.NewWriter().SendDatagram(payload, rx3.Addr()); err != nil {
			t.Fatalf("send DATAGRAM3 failed: %v", err)
		}

		dg3, err := receiveWithTimeout(45*time.Second, rx3.NewReader().ReceiveDatagram)
		if err != nil {
			t.Fatalf("receive DATAGRAM3 failed: %v", err)
		}
		if !bytes.Equal(dg3.Data, payload) {
			t.Fatalf("DATAGRAM3 payload mismatch: got %q want %q", string(dg3.Data), string(payload))
		}
		if err := dg3.ResolveSource(rx3); err != nil {
			t.Fatalf("resolve DATAGRAM3 source failed: %v", err)
		}
		if dg3.Source == "" {
			t.Fatal("DATAGRAM3 resolved source is empty")
		}
	})
}

func TestSAMv33_PrimarySessionE2E(t *testing.T) {
	ensureBridgeReady(t)

	sam, err := NewSAM(integrationSAMAddr())
	if err != nil {
		t.Fatalf("failed to create SAM: %v", err)
	}
	defer sam.Close()

	keys, err := sam.NewKeys()
	if err != nil {
		t.Fatalf("failed to create keys: %v", err)
	}

	primary, err := sam.NewPrimarySession(uniqueID("primary"), keys, integrationTunnelOptions)
	if err != nil {
		t.Fatalf("failed to create PRIMARY session: %v", err)
	}
	defer primary.Close()

	streamSub, err := primary.NewStreamSubSession(uniqueID("p-stream"), []string{"LISTEN_PORT=18081"})
	if err != nil {
		t.Fatalf("failed to create STREAM subsession: %v", err)
	}
	defer streamSub.Close()

	datagramSub, err := primary.NewDatagramSubSession(uniqueID("p-dg"), nil)
	if err != nil {
		t.Fatalf("failed to create DATAGRAM subsession: %v", err)
	}
	defer datagramSub.Close()

	rawSub, err := primary.NewRawSubSession(uniqueID("p-raw"), nil)
	if err != nil {
		t.Fatalf("failed to create RAW subsession: %v", err)
	}
	defer rawSub.Close()

	datagram3Sub, err := primary.NewDatagram3SubSession(uniqueID("p-dg3"), nil)
	if err != nil {
		t.Fatalf("failed to create DATAGRAM3 subsession: %v", err)
	}
	defer datagram3Sub.Close()

	manualSubID := uniqueID("manual-addremove")
	if err := sam.AddSubSession("STREAM", manualSubID, []string{"LISTEN_PORT=18082"}); err != nil {
		t.Fatalf("failed SESSION ADD: %v", err)
	}
	if err := sam.RemoveSubSession(manualSubID); err != nil {
		t.Fatalf("failed SESSION REMOVE: %v", err)
	}
}
