package datagram2

import (
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// Datagram2Session represents an authenticated datagram2 session with replay protection.
// This session type provides UDP-like messaging capabilities through the I2P network with
// enhanced security features compared to legacy DATAGRAM sessions. DATAGRAM2 provides
// replay protection and offline signature support, making it the recommended format for
// new applications that don't require backward compatibility.
//
// The session manages the underlying I2P connection and provides methods for creating readers
// and writers. For SAMv3 mode, it uses UDP forwarding where datagrams are received via a
// local UDP socket that the SAM bridge forwards to.
//
// I2P Timing Considerations:
//   - Session creation: 2-5 minutes for tunnel establishment
//   - Message delivery: Variable latency (network-dependent)
//   - Use generous timeouts and retry logic with exponential backoff
//
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	session, err := NewDatagram2Session(sam, "my-session", keys, options)
type Datagram2Session struct {
	*common.BaseSession
	sam        *common.SAM
	options    []string
	mu         sync.RWMutex
	closed     bool
	udpConn    *net.UDPConn // UDP connection for receiving forwarded datagrams (SAMv3 mode)
	udpEnabled bool         // Whether UDP forwarding is enabled (always true for SAMv3)
}

// Datagram2Reader handles incoming authenticated datagram2 reception from the I2P network.
// It provides asynchronous datagram reception through buffered channels, allowing applications
// to receive datagrams without blocking. The reader manages its own goroutine for continuous
// message processing and provides thread-safe access to received datagrams.
//
// All datagrams received are authenticated by the I2P router with signature verification
// performed internally. DATAGRAM2 provides replay protection, preventing replay attacks
// that are possible with legacy DATAGRAM sessions.
//
// Example usage:
//
//	reader := session.NewReader()
//	for {
//	    datagram, err := reader.ReceiveDatagram()
//	    if err != nil {
//	        // Handle error
//	    }
//	    // Process datagram.Data from datagram.Source
//	}
type Datagram2Reader struct {
	session     *Datagram2Session
	recvChan    chan *Datagram2
	errorChan   chan error
	closeChan   chan struct{}
	doneChan    chan struct{}
	closed      bool
	loopStarted bool
	mu          sync.RWMutex
	closeOnce   sync.Once
}

// Datagram2Writer handles outgoing authenticated datagram2 transmission to I2P destinations.
// It provides methods for sending datagrams with configurable timeouts and handles
// the underlying SAM protocol communication for message delivery. The writer supports
// method chaining for configuration and provides error handling for send operations.
//
// All datagrams are authenticated and provide replay protection. Maximum datagram size
// is 31744 bytes total (including headers), with 11 KB recommended for best reliability.
//
// Example usage:
//
//	writer := session.NewWriter().SetTimeout(30*time.Second)
//	err := writer.SendDatagram(data, destination)
type Datagram2Writer struct {
	session *Datagram2Session
	timeout time.Duration
}

// Datagram2 represents an authenticated I2P datagram2 message with replay protection.
// It encapsulates the payload data along with source and destination addressing details,
// providing all necessary information for processing received datagrams. The structure
// includes both the raw data bytes and I2P address information for routing.
//
// All DATAGRAM2 messages are authenticated by the I2P router, with signature verification
// performed internally. Replay protection prevents replay attacks that affect legacy DATAGRAM.
//
// Example usage:
//
//	if datagram.Source.Base32() == expectedSender {
//	    processData(datagram.Data)
//	}
type Datagram2 struct {
	Data   []byte          // Raw datagram payload (up to ~31KB)
	Source i2pkeys.I2PAddr // Authenticated source destination
	Local  i2pkeys.I2PAddr // Local destination (this session)
}

// Datagram2Addr implements net.Addr interface for I2P datagram2 addresses.
// This type provides standard Go networking address representation for I2P destinations,
// allowing seamless integration with existing Go networking code that expects net.Addr.
// The address wraps an I2P address and provides string representation and network type identification.
//
// Example usage:
//
//	addr := &Datagram2Addr{addr: destination}
//	fmt.Println(addr.Network(), addr.String())
type Datagram2Addr struct {
	addr i2pkeys.I2PAddr
}

// Network returns the network type for I2P datagram2 addresses.
// This implements the net.Addr interface by returning "datagram2" as the network type.
func (a *Datagram2Addr) Network() string {
	return "datagram2"
}

// String returns the string representation of the I2P address.
// This implements the net.Addr interface by returning the base32 address representation,
// which is suitable for display and logging purposes.
func (a *Datagram2Addr) String() string {
	return a.addr.Base32()
}

// Datagram2Conn implements net.PacketConn interface for I2P datagram2 communication.
// This type provides compatibility with standard Go networking patterns by wrapping
// datagram2 session functionality in a familiar PacketConn interface. It manages
// internal readers and writers while providing standard connection operations.
//
// The connection provides thread-safe concurrent access to I2P datagram2 operations
// and properly handles cleanup on close. All datagrams are authenticated with replay
// protection provided by the DATAGRAM2 format.
//
// Example usage:
//
//	conn := session.PacketConn()
//	n, addr, err := conn.ReadFrom(buffer)
//	n, err = conn.WriteTo(data, destination)
type Datagram2Conn struct {
	session    *Datagram2Session
	reader     *Datagram2Reader
	writer     *Datagram2Writer
	remoteAddr *i2pkeys.I2PAddr
	mu         sync.RWMutex
	closed     bool
	cleanup    runtime.Cleanup
}
