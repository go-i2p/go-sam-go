package datagram

import (
	"runtime"
	"sync"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// DatagramSession represents a datagram session that can send and receive datagrams over I2P.
// This session type provides UDP-like messaging capabilities through the I2P network, allowing
// applications to send and receive datagrams with message reliability and ordering guarantees.
// The session manages the underlying I2P connection and provides methods for creating readers and writers.
// Example usage: session, err := NewDatagramSession(sam, "my-session", keys, options)
type DatagramSession struct {
	*common.BaseSession
	sam     *common.SAM
	options []string
	mu      sync.RWMutex
	closed  bool
}

// DatagramReader handles incoming datagram reception from the I2P network.
// It provides asynchronous datagram reception through buffered channels, allowing applications
// to receive datagrams without blocking. The reader manages its own goroutine for continuous
// message processing and provides thread-safe access to received datagrams.
// Example usage: reader := session.NewReader(); datagram, err := reader.ReceiveDatagram()
type DatagramReader struct {
	session   *DatagramSession
	recvChan  chan *Datagram
	errorChan chan error
	closeChan chan struct{}
	doneChan  chan struct{}
	closed    bool
	mu        sync.RWMutex
	closeOnce sync.Once
}

// DatagramWriter handles outgoing datagram transmission to I2P destinations.
// It provides methods for sending datagrams with configurable timeouts and handles
// the underlying SAM protocol communication for message delivery. The writer supports
// method chaining for configuration and provides error handling for send operations.
// Example usage: writer := session.NewWriter().SetTimeout(30*time.Second); err := writer.SendDatagram(data, dest)
type DatagramWriter struct {
	session *DatagramSession
	timeout time.Duration
}

// Datagram represents an I2P datagram message containing data and address information.
// It encapsulates the payload data along with source and destination addressing details,
// providing all necessary information for processing received datagrams or preparing outgoing ones.
// The structure includes both the raw data bytes and I2P address information for routing.
// Example usage: if datagram.Source.Base32() == expectedSender { processData(datagram.Data) }
type Datagram struct {
	Data   []byte
	Source i2pkeys.I2PAddr
	Local  i2pkeys.I2PAddr
}

// DatagramAddr implements net.Addr interface for I2P datagram addresses.
// This type provides standard Go networking address representation for I2P destinations,
// allowing seamless integration with existing Go networking code that expects net.Addr.
// The address wraps an I2P address and provides string representation and network type identification.
// Example usage: addr := &DatagramAddr{addr: destination}; fmt.Println(addr.Network(), addr.String())
type DatagramAddr struct {
	addr i2pkeys.I2PAddr
}

// DatagramConn implements net.PacketConn interface for I2P datagram communication.
// This type provides compatibility with standard Go networking patterns by wrapping
// datagram session functionality in a familiar PacketConn interface. It manages
// internal readers and writers while providing standard connection operations.
// Example usage: conn := session.PacketConn(); n, addr, err := conn.ReadFrom(buffer)
type DatagramConn struct {
	session    *DatagramSession
	reader     *DatagramReader
	writer     *DatagramWriter
	remoteAddr *i2pkeys.I2PAddr
	mu         sync.RWMutex
	closed     bool
	cleanup    runtime.Cleanup
}
