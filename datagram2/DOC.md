# datagram2
--
    import "github.com/go-i2p/go-sam-go/datagram2"

Package datagram2 provides authenticated datagram sessions with replay
protection for I2P.

DATAGRAM2 sessions provide authenticated, repliable UDP-like messaging over I2P
tunnels with replay attack protection. This is the recommended datagram format
for applications requiring both source authentication and replay protection.

Key features:

    - Authenticated datagrams with signature verification
    - Replay protection (not available in legacy DATAGRAM)
    - Repliable (can send replies to sender)
    - UDP-like messaging (unreliable, unordered)
    - Maximum 31744 bytes per datagram (11 KB recommended)
    - Implements net.PacketConn interface

Session creation requires 2-5 minutes for I2P tunnel establishment. Use generous
timeouts and exponential backoff retry logic.

Basic usage:

    sam, err := common.NewSAM("127.0.0.1:7656")
    session, err := datagram2.NewDatagram2Session(sam, "my-session", keys, []string{"inbound.length=1"})
    defer session.Close()
    conn := session.PacketConn()
    n, err := conn.WriteTo(data, destination)
    n, addr, err := conn.ReadFrom(buffer)

See also: Package datagram (legacy, no replay protection), datagram3
(unauthenticated), stream (TCP-like), raw (non-repliable), primary
(multi-session management).

## Usage

#### type Datagram2

```go
type Datagram2 struct {
	Data   []byte          // Raw datagram payload (up to ~31KB)
	Source i2pkeys.I2PAddr // Authenticated source destination
	Local  i2pkeys.I2PAddr // Local destination (this session)
}
```

Datagram2 represents an authenticated I2P datagram2 message with replay
protection. It encapsulates the payload data along with source and destination
addressing details, providing all necessary information for processing received
datagrams. The structure includes both the raw data bytes and I2P address
information for routing.

All DATAGRAM2 messages are authenticated by the I2P router, with signature
verification performed internally. Replay protection prevents replay attacks
that affect legacy DATAGRAM.

Example usage:

    if datagram.Source.Base32() == expectedSender {
        processData(datagram.Data)
    }

#### type Datagram2Addr

```go
type Datagram2Addr struct {
}
```

Datagram2Addr implements net.Addr interface for I2P datagram2 addresses. This
type provides standard Go networking address representation for I2P
destinations, allowing seamless integration with existing Go networking code
that expects net.Addr. The address wraps an I2P address and provides string
representation and network type identification.

Example usage:

    addr := &Datagram2Addr{addr: destination}
    fmt.Println(addr.Network(), addr.String())

#### func (*Datagram2Addr) Network

```go
func (a *Datagram2Addr) Network() string
```
Network returns the network type for I2P datagram2 addresses. This implements
the net.Addr interface by returning "datagram2" as the network type.

#### func (*Datagram2Addr) String

```go
func (a *Datagram2Addr) String() string
```
String returns the string representation of the I2P address. This implements the
net.Addr interface by returning the base32 address representation, which is
suitable for display and logging purposes.

#### type Datagram2Conn

```go
type Datagram2Conn struct {
}
```

Datagram2Conn implements net.PacketConn interface for I2P datagram2
communication. This type provides compatibility with standard Go networking
patterns by wrapping datagram2 session functionality in a familiar PacketConn
interface. It manages internal readers and writers while providing standard
connection operations.

The connection provides thread-safe concurrent access to I2P datagram2
operations and properly handles cleanup on close. All datagrams are
authenticated with replay protection provided by the DATAGRAM2 format.

Example usage:

    conn := session.PacketConn()
    n, addr, err := conn.ReadFrom(buffer)
    n, err = conn.WriteTo(data, destination)

#### func (*Datagram2Conn) Close

```go
func (c *Datagram2Conn) Close() error
```
Close closes the datagram2 connection and releases associated resources. This
method implements the net.Conn interface. It closes the reader and writer but
does not close the underlying session, which may be shared by other connections.
Multiple calls to Close are safe and will return nil after the first call.

#### func (*Datagram2Conn) LocalAddr

```go
func (c *Datagram2Conn) LocalAddr() net.Addr
```
LocalAddr returns the local network address as a Datagram2Addr containing the
I2P destination address of this connection's session. This method implements the
net.Conn interface and provides access to the local I2P destination.

#### func (*Datagram2Conn) Read

```go
func (c *Datagram2Conn) Read(b []byte) (n int, err error)
```
Read implements net.Conn by wrapping ReadFrom for stream-like usage. It reads
data into the provided byte slice and returns the number of bytes read. When
reading, it also updates the remote address of the connection for subsequent
Write calls. Note: This is not typical for datagrams which are connectionless,
but provides compatibility with the net.Conn interface.

#### func (*Datagram2Conn) ReadFrom

```go
func (c *Datagram2Conn) ReadFrom(p []byte) (n int, addr net.Addr, err error)
```
ReadFrom reads an authenticated datagram with replay protection from the
connection. This method implements the net.PacketConn interface. It starts the
receive loop if not already started and blocks until a datagram is received. The
data is copied to the provided buffer p, and the authenticated source address is
returned as a Datagram2Addr.

All datagrams are authenticated by the I2P router with DATAGRAM2 replay
protection.

#### func (*Datagram2Conn) RemoteAddr

```go
func (c *Datagram2Conn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote network address of the connection. This method
implements the net.Conn interface. For datagram2 connections, this returns the
authenticated address of the last peer that sent data (set by Read), or nil if
no data has been received yet.

#### func (*Datagram2Conn) SetDeadline

```go
func (c *Datagram2Conn) SetDeadline(t time.Time) error
```
SetDeadline sets both read and write deadlines for the connection. This method
implements the net.Conn interface by calling both SetReadDeadline and
SetWriteDeadline with the same time value. If either deadline cannot be set, the
first error encountered is returned.

#### func (*Datagram2Conn) SetReadDeadline

```go
func (c *Datagram2Conn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future ReadFrom calls. This method
implements the net.Conn interface. For datagram2 connections, this is currently
a placeholder implementation that always returns nil. Timeout handling is
managed differently for datagram operations.

#### func (*Datagram2Conn) SetWriteDeadline

```go
func (c *Datagram2Conn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future WriteTo calls. This method
implements the net.Conn interface. If the deadline is not zero, it calculates
the timeout duration and sets it on the writer for subsequent write operations.

#### func (*Datagram2Conn) Write

```go
func (c *Datagram2Conn) Write(b []byte) (n int, err error)
```
Write implements net.Conn by wrapping WriteTo for stream-like usage. It writes
data to the remote address set by the last Read operation and returns the number
of bytes written. If no remote address has been set, it returns an error. Note:
This is not typical for datagrams which are connectionless, but provides
compatibility with the net.Conn interface.

#### func (*Datagram2Conn) WriteTo

```go
func (c *Datagram2Conn) WriteTo(p []byte, addr net.Addr) (n int, err error)
```
WriteTo writes an authenticated datagram with replay protection to the specified
address. This method implements the net.PacketConn interface. The address must
be a Datagram2Addr containing a valid I2P destination. The entire byte slice p
is sent as a single authenticated datagram message with replay protection.

#### type Datagram2Reader

```go
type Datagram2Reader struct {
}
```

Datagram2Reader handles incoming authenticated datagram2 reception from the I2P
network. It provides asynchronous datagram reception through buffered channels,
allowing applications to receive datagrams without blocking. The reader manages
its own goroutine for continuous message processing and provides thread-safe
access to received datagrams.

All datagrams received are authenticated by the I2P router with signature
verification performed internally. DATAGRAM2 provides replay protection,
preventing replay attacks that are possible with legacy DATAGRAM sessions.

Example usage:

    reader := session.NewReader()
    for {
        datagram, err := reader.ReceiveDatagram()
        if err != nil {
            // Handle error
        }
        // Process datagram.Data from datagram.Source
    }

#### func (*Datagram2Reader) Close

```go
func (r *Datagram2Reader) Close() error
```
Close closes the Datagram2Reader and stops its receive loop. This method safely
terminates the reader, cleans up all associated resources, and signals any
waiting goroutines to stop. It's safe to call multiple times and will not block
if the reader is already closed.

Example usage:

    defer reader.Close()

#### func (*Datagram2Reader) ReceiveDatagram

```go
func (r *Datagram2Reader) ReceiveDatagram() (*Datagram2, error)
```
ReceiveDatagram receives a single authenticated datagram with replay protection
from the I2P network. This method blocks until a datagram is received or an
error occurs, returning the received datagram with its data and authenticated
addressing information. It handles concurrent access safely and provides proper
error handling for network issues.

All datagrams are authenticated by the I2P router with DATAGRAM2 replay
protection.

Example usage:

    datagram, err := reader.ReceiveDatagram()
    if err != nil {
        // Handle error
    }
    // Process datagram.Data from authenticated datagram.Source

#### type Datagram2Session

```go
type Datagram2Session struct {
	*common.BaseSession
}
```

Datagram2Session represents an authenticated datagram2 session with replay
protection. This session type provides UDP-like messaging capabilities through
the I2P network with enhanced security features compared to legacy DATAGRAM
sessions. DATAGRAM2 provides replay protection and offline signature support,
making it the recommended format for new applications that don't require
backward compatibility.

The session manages the underlying I2P connection and provides methods for
creating readers and writers. For SAMv3 mode, it uses UDP forwarding where
datagrams are received via a local UDP socket that the SAM bridge forwards to.

I2P Timing Considerations:

    - Session creation: 2-5 minutes for tunnel establishment
    - Message delivery: Variable latency (network-dependent)
    - Use generous timeouts and retry logic with exponential backoff

Example usage:

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    session, err := NewDatagram2Session(sam, "my-session", keys, options)

#### func  NewDatagram2Session

```go
func NewDatagram2Session(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*Datagram2Session, error)
```
NewDatagram2Session creates a new datagram2 session with replay protection for
UDP-like I2P messaging. It initializes the session with the provided SAM
connection, session ID, cryptographic keys, and configuration options. The
session automatically creates a UDP listener for receiving forwarded datagrams
per SAMv3 requirements. Example usage: session, err := NewDatagram2Session(sam,
"my-session", keys, []string{"inbound.length=1"})

#### func  NewDatagram2SessionFromSubsession

```go
func NewDatagram2SessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, udpConn *net.UDPConn) (*Datagram2Session, error)
```
NewDatagram2SessionFromSubsession creates a Datagram2Session for a subsession
that has already been registered with a PRIMARY session using SESSION ADD. This
constructor skips the session creation step since the subsession is already
registered with the SAM bridge.

For PRIMARY datagram2 subsessions, UDP forwarding is mandatory (SAMv3
requirement). The UDP connection must be provided for proper datagram reception.

Example usage: session, err := NewDatagram2SessionFromSubsession(sam, "sub1",
keys, options, udpConn)

#### func (*Datagram2Session) Addr

```go
func (s *Datagram2Session) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this datagram2 session. This address represents
the session's identity on the I2P network and can be used by other nodes to send
authenticated datagrams with replay protection to this session. The address is
derived from the session's cryptographic keys. Example usage: myAddr :=
session.Addr(); fmt.Println("My I2P address:", myAddr.Base32())

#### func (*Datagram2Session) Close

```go
func (s *Datagram2Session) Close() error
```
Close closes the datagram2 session and all associated resources. This method
safely terminates the session, closes the UDP listener and underlying
connection, and cleans up any background goroutines. It's safe to call multiple
times. Example usage: defer session.Close()

#### func (*Datagram2Session) NewReader

```go
func (s *Datagram2Session) NewReader() *Datagram2Reader
```
NewReader creates a Datagram2Reader for receiving authenticated datagrams with
replay protection. This method initializes a new reader with buffered channels
for asynchronous datagram reception. The reader must be started manually with
receiveLoop() for continuous operation. Example usage: reader :=
session.NewReader(); go reader.receiveLoop(); datagram, err :=
reader.ReceiveDatagram()

#### func (*Datagram2Session) NewWriter

```go
func (s *Datagram2Session) NewWriter() *Datagram2Writer
```
NewWriter creates a Datagram2Writer for sending authenticated datagrams with
replay protection. This method initializes a new writer with a default timeout
of 30 seconds for send operations. The timeout can be customized using the
SetTimeout method on the returned writer. Example usage: writer :=
session.NewWriter().SetTimeout(60*time.Second); err := writer.SendDatagram(data,
dest)

#### func (*Datagram2Session) PacketConn

```go
func (s *Datagram2Session) PacketConn() net.PacketConn
```
PacketConn returns a net.PacketConn interface for this datagram2 session. This
method provides compatibility with standard Go networking code by wrapping the
datagram2 session in a connection that implements the PacketConn interface. The
connection provides authenticated datagrams with replay protection.

Example usage:

    conn := session.PacketConn()
    n, addr, err := conn.ReadFrom(buffer)
    n, err = conn.WriteTo(data, destination)

#### func (*Datagram2Session) ReceiveDatagram

```go
func (s *Datagram2Session) ReceiveDatagram() (*Datagram2, error)
```
ReceiveDatagram receives a single authenticated datagram from the I2P network.
This method is a convenience wrapper that performs a direct single read
operation without starting a continuous receive loop. For continuous reception,
use NewReader() and manage the reader lifecycle manually. Example usage:
datagram, err := session.ReceiveDatagram()

#### func (*Datagram2Session) SendDatagram

```go
func (s *Datagram2Session) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends an authenticated datagram with replay protection to the
specified destination. This is a convenience method that creates a temporary
writer and sends the datagram immediately. For multiple sends, it's more
efficient to create a writer once and reuse it. Example usage: err :=
session.SendDatagram([]byte("hello"), destinationAddr)

#### type Datagram2Writer

```go
type Datagram2Writer struct {
}
```

Datagram2Writer handles outgoing authenticated datagram2 transmission to I2P
destinations. It provides methods for sending datagrams with configurable
timeouts and handles the underlying SAM protocol communication for message
delivery. The writer supports method chaining for configuration and provides
error handling for send operations.

All datagrams are authenticated and provide replay protection. Maximum datagram
size is 31744 bytes total (including headers), with 11 KB recommended for best
reliability.

Example usage:

    writer := session.NewWriter().SetTimeout(30*time.Second)
    err := writer.SendDatagram(data, destination)

#### func (*Datagram2Writer) SendDatagram

```go
func (w *Datagram2Writer) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends an authenticated datagram with replay protection to the
specified I2P destination. It uses the SAMv3 UDP approach by sending to port
7655 with DATAGRAM2 format. Maximum datagram size is 31744 bytes (11 KB
recommended for reliability). Example usage: err :=
writer.SendDatagram([]byte("hello world"), destinationAddr)

#### func (*Datagram2Writer) SetTimeout

```go
func (w *Datagram2Writer) SetTimeout(timeout time.Duration) *Datagram2Writer
```
SetTimeout sets the timeout for datagram2 write operations. This method
configures the maximum time to wait for authenticated datagram send operations
to complete. Returns the writer instance for method chaining convenience.
Example usage: writer.SetTimeout(30*time.Second).SendDatagram(data, destination)

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide datagram2-specific functionality for I2P
messaging. This type extends the base SAM functionality with methods
specifically designed for DATAGRAM2 communication, providing authenticated
datagrams with replay protection. DATAGRAM2 is the recommended format for new
applications, offering enhanced security over legacy DATAGRAM sessions through
replay attack prevention and offline signature support. Example usage: sam :=
&SAM{SAM: baseSAM}; session, err := sam.NewDatagram2Session(id, keys, options)

#### func (*SAM) NewDatagram2Session

```go
func (s *SAM) NewDatagram2Session(id string, keys i2pkeys.I2PKeys, options []string) (*Datagram2Session, error)
```
NewDatagram2Session creates a new datagram2 session with the SAM bridge using
default settings. This method establishes a new DATAGRAM2 session for
authenticated UDP-like messaging over I2P with replay protection. It uses
default signature settings (Ed25519) and automatically configures UDP forwarding
for SAMv3 compatibility. Session creation can take 2-5 minutes due to I2P tunnel
establishment, so generous timeouts are recommended.

DATAGRAM2 provides enhanced security compared to legacy DATAGRAM:

    - Replay protection prevents replay attacks
    - Offline signature support for advanced key management
    - Identical SAM API for easy migration

Example usage:

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    session, err := sam.NewDatagram2Session("my-session", keys, []string{"inbound.length=1"})

#### func (*SAM) NewDatagram2SessionWithPorts

```go
func (s *SAM) NewDatagram2SessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*Datagram2Session, error)
```
NewDatagram2SessionWithPorts creates a new datagram2 session with port
specifications. This method allows configuring specific I2CP port ranges for the
session, enabling fine-grained control over network communication ports for
advanced routing scenarios. Port configuration is useful for applications
requiring specific port mappings or PRIMARY session subsessions. This function
automatically creates a UDP listener for SAMv3 UDP forwarding (required for v3
mode).

The FROM_PORT and TO_PORT parameters specify I2CP ports for protocol-level
communication, distinct from the UDP forwarding port which is auto-assigned by
the OS.

Example usage:

    session, err := sam.NewDatagram2SessionWithPorts(id, "8080", "8081", keys, options)

#### func (*SAM) NewDatagram2SessionWithSignature

```go
func (s *SAM) NewDatagram2SessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*Datagram2Session, error)
```
NewDatagram2SessionWithSignature creates a new datagram2 session with custom
signature type. This method allows specifying a custom cryptographic signature
type for the session, enabling advanced security configurations beyond the
default Ed25519 algorithm. DATAGRAM2 supports offline signatures, allowing
pre-signed destinations for enhanced privacy and key management flexibility.

Different signature types provide various security levels and compatibility
options:

    - Ed25519 (type 7) - Recommended for most applications
    - ECDSA (types 1-3) - Legacy compatibility
    - RedDSA (type 11) - Advanced privacy features

Example usage:

    session, err := sam.NewDatagram2SessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
