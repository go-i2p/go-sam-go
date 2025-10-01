# datagram
--
    import "github.com/go-i2p/go-sam-go/datagram"


## Usage

#### type Datagram

```go
type Datagram struct {
	Data   []byte
	Source i2pkeys.I2PAddr
	Local  i2pkeys.I2PAddr
}
```

Datagram represents an I2P datagram message containing data and address
information. It encapsulates the payload data along with source and destination
addressing details, providing all necessary information for processing received
datagrams or preparing outgoing ones. The structure includes both the raw data
bytes and I2P address information for routing. Example usage: if
datagram.Source.Base32() == expectedSender { processData(datagram.Data) }

#### type DatagramAddr

```go
type DatagramAddr struct {
}
```

DatagramAddr implements net.Addr interface for I2P datagram addresses. This type
provides standard Go networking address representation for I2P destinations,
allowing seamless integration with existing Go networking code that expects
net.Addr. The address wraps an I2P address and provides string representation
and network type identification. Example usage: addr := &DatagramAddr{addr:
destination}; fmt.Println(addr.Network(), addr.String())

#### func (*DatagramAddr) Network

```go
func (a *DatagramAddr) Network() string
```
Network returns the network type for this address. This method implements the
net.Addr interface and always returns "i2p-datagram" to identify this as an I2P
datagram address type for networking compatibility. Example usage: network :=
addr.Network() // returns "i2p-datagram"

#### func (*DatagramAddr) String

```go
func (a *DatagramAddr) String() string
```
String returns the string representation of the address. This method implements
the net.Addr interface and returns the Base32 encoded representation of the I2P
address for human-readable display and logging. Example usage: addrStr :=
addr.String() // returns "abcd1234...xyz.b32.i2p"

#### type DatagramConn

```go
type DatagramConn struct {
}
```

DatagramConn implements net.PacketConn interface for I2P datagram communication.
This type provides compatibility with standard Go networking patterns by
wrapping datagram session functionality in a familiar PacketConn interface. It
manages internal readers and writers while providing standard connection
operations. Example usage: conn := session.PacketConn(); n, addr, err :=
conn.ReadFrom(buffer)

#### func (*DatagramConn) Close

```go
func (c *DatagramConn) Close() error
```
Close closes the datagram connection and releases associated resources. This
method implements the net.Conn interface. It closes the reader and writer but
does not close the underlying session, which may be shared by other connections.
Multiple calls to Close are safe and will return nil after the first call.

#### func (*DatagramConn) LocalAddr

```go
func (c *DatagramConn) LocalAddr() net.Addr
```
LocalAddr returns the local network address as a DatagramAddr containing the I2P
destination address of this connection's session. This method implements the
net.Conn interface and provides access to the local I2P destination.

#### func (*DatagramConn) Read

```go
func (c *DatagramConn) Read(b []byte) (n int, err error)
```
Read implements net.Conn by wrapping ReadFrom for stream-like usage. It reads
data into the provided byte slice and returns the number of bytes read. When
reading, it also updates the remote address of the connection for subsequent
Write calls. Note: This is not typical for datagrams which are connectionless,
but provides compatibility with the net.Conn interface.

#### func (*DatagramConn) ReadFrom

```go
func (c *DatagramConn) ReadFrom(p []byte) (n int, addr net.Addr, err error)
```
ReadFrom reads a datagram from the connection and returns the number of bytes
read, the source address, and any error encountered. This method implements the
net.PacketConn interface. It starts the receive loop if not already started and
blocks until a datagram is received. The data is copied to the provided buffer
p, and the source address is returned as a DatagramAddr.

#### func (*DatagramConn) RemoteAddr

```go
func (c *DatagramConn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote network address of the connection. This method
implements the net.Conn interface. For datagram connections, this returns the
address of the last peer that sent data (set by Read), or nil if no data has
been received yet.

#### func (*DatagramConn) SetDeadline

```go
func (c *DatagramConn) SetDeadline(t time.Time) error
```
SetDeadline sets both read and write deadlines for the connection. This method
implements the net.Conn interface by calling both SetReadDeadline and
SetWriteDeadline with the same time value. If either deadline cannot be set, the
first error encountered is returned.

#### func (*DatagramConn) SetReadDeadline

```go
func (c *DatagramConn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future ReadFrom calls. This method
implements the net.Conn interface. For datagram connections, this is currently a
placeholder implementation that always returns nil. Timeout handling is managed
differently for datagram operations.

#### func (*DatagramConn) SetWriteDeadline

```go
func (c *DatagramConn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future WriteTo calls. This method
implements the net.Conn interface. If the deadline is not zero, it calculates
the timeout duration and sets it on the writer for subsequent write operations.

#### func (*DatagramConn) Write

```go
func (c *DatagramConn) Write(b []byte) (n int, err error)
```
Write implements net.Conn by wrapping WriteTo for stream-like usage. It writes
data to the remote address set by the last Read operation and returns the number
of bytes written. If no remote address has been set, it returns an error. Note:
This is not typical for datagrams which are connectionless, but provides
compatibility with the net.Conn interface.

#### func (*DatagramConn) WriteTo

```go
func (c *DatagramConn) WriteTo(p []byte, addr net.Addr) (n int, err error)
```
WriteTo writes a datagram to the specified address and returns the number of
bytes written and any error encountered. This method implements the
net.PacketConn interface. The address must be a DatagramAddr containing a valid
I2P destination. The entire byte slice p is sent as a single datagram message.

#### type DatagramListener

```go
type DatagramListener struct {
}
```

DatagramListener implements net.Listener for I2P datagram connections. It
provides a way to accept incoming datagram connections in a stream-like manner,
where each accepted connection represents a new DatagramConn that can be used
for bidirectional communication with remote I2P destinations.

#### func (*DatagramListener) Accept

```go
func (l *DatagramListener) Accept() (net.Conn, error)
```
Accept waits for and returns the next datagram connection to the listener. This
method implements the net.Listener interface. It blocks until a new connection
is available or an error occurs. Each accepted connection is a new DatagramConn
that shares the underlying session but has its own reader/writer.

#### func (*DatagramListener) Addr

```go
func (l *DatagramListener) Addr() net.Addr
```
Addr returns the listener's network address as a DatagramAddr. This method
implements the net.Listener interface and provides access to the I2P destination
address that this listener is bound to.

#### func (*DatagramListener) Close

```go
func (l *DatagramListener) Close() error
```
Close closes the datagram listener and releases associated resources. This
method implements the net.Listener interface. It stops accepting new connections
and closes the reader. The underlying session is not closed as it may be shared
by other components. Multiple calls to Close are safe.

#### type DatagramReader

```go
type DatagramReader struct {
}
```

DatagramReader handles incoming datagram reception from the I2P network. It
provides asynchronous datagram reception through buffered channels, allowing
applications to receive datagrams without blocking. The reader manages its own
goroutine for continuous message processing and provides thread-safe access to
received datagrams. Example usage: reader := session.NewReader(); datagram, err
:= reader.ReceiveDatagram()

#### func (*DatagramReader) Close

```go
func (r *DatagramReader) Close() error
```
Close closes the DatagramReader and stops its receive loop. This method safely
terminates the reader, cleans up all associated resources, and signals any
waiting goroutines to stop. It's safe to call multiple times and will not block
if the reader is already closed. Example usage: defer reader.Close()

#### func (*DatagramReader) ReceiveDatagram

```go
func (r *DatagramReader) ReceiveDatagram() (*Datagram, error)
```
ReceiveDatagram receives a single datagram from the I2P network. This method
blocks until a datagram is received or an error occurs, returning the received
datagram with its data and addressing information. It handles concurrent access
safely and provides proper error handling for network issues. Example usage:
datagram, err := reader.ReceiveDatagram()

#### type DatagramSession

```go
type DatagramSession struct {
	*common.BaseSession
}
```

DatagramSession represents a datagram session that can send and receive
datagrams over I2P. This session type provides UDP-like messaging capabilities
through the I2P network, allowing applications to send and receive datagrams
with message reliability and ordering guarantees. The session manages the
underlying I2P connection and provides methods for creating readers and writers.
Example usage: session, err := NewDatagramSession(sam, "my-session", keys,
options)

#### func  NewDatagramSession

```go
func NewDatagramSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error)
```
NewDatagramSession creates a new datagram session for UDP-like I2P messaging.
This function establishes a new datagram session with the provided SAM
connection, session ID, cryptographic keys, and configuration options. It
returns a DatagramSession instance that can be used for sending and receiving
datagrams over the I2P network. Example usage: session, err :=
NewDatagramSession(sam, "my-session", keys, []string{"inbound.length=1"})

#### func (*DatagramSession) Addr

```go
func (s *DatagramSession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this session. This address represents the
session's identity on the I2P network and can be used by other nodes to send
datagrams to this session. The address is derived from the session's
cryptographic keys. Example usage: myAddr := session.Addr(); fmt.Println("My I2P
address:", myAddr.Base32())

#### func (*DatagramSession) Close

```go
func (s *DatagramSession) Close() error
```
Close closes the datagram session and all associated resources. This method
safely terminates the session, closes the underlying connection, and cleans up
any background goroutines. It's safe to call multiple times. Example usage:
defer session.Close()

#### func (*DatagramSession) Dial

```go
func (ds *DatagramSession) Dial(destination string) (net.PacketConn, error)
```
Dial establishes a datagram connection to the specified I2P destination. This
method creates a net.PacketConn interface for sending and receiving datagrams
with the specified destination. It uses a default timeout of 30 seconds for
connection establishment and provides UDP-like communication over I2P networks.
Example usage: conn, err := session.Dial("destination.b32.i2p")

#### func (*DatagramSession) DialContext

```go
func (ds *DatagramSession) DialContext(ctx context.Context, destination string) (net.PacketConn, error)
```
DialContext establishes a datagram connection with context support for
cancellation. This method provides the core dialing functionality with
context-based cancellation support, allowing for proper resource cleanup and
operation cancellation through the provided context. It validates the
destination and session state before attempting connection establishment.
Example usage: conn, err := session.DialContext(ctx, "destination.b32.i2p")

#### func (*DatagramSession) DialI2P

```go
func (ds *DatagramSession) DialI2P(addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2P establishes a datagram connection to an I2P address using native
addressing. This method creates a net.PacketConn interface for communicating
with the specified I2P address using the native i2pkeys.I2PAddr type. It uses a
default timeout of 30 seconds and provides type-safe addressing for I2P
destinations. Example usage: conn, err := session.DialI2P(i2pAddress)

#### func (*DatagramSession) DialI2PContext

```go
func (ds *DatagramSession) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2PContext establishes a datagram connection to an I2P address with context
support. This method provides the core I2P dialing functionality with
context-based cancellation, allowing for proper resource cleanup and operation
cancellation through the provided context. It validates the session state and
creates a connection with integrated reader and writer. Example usage: conn, err
:= session.DialI2PContext(ctx, i2pAddress)

#### func (*DatagramSession) DialI2PTimeout

```go
func (ds *DatagramSession) DialI2PTimeout(addr i2pkeys.I2PAddr, timeout time.Duration) (net.PacketConn, error)
```
DialI2PTimeout establishes a datagram connection to an I2P address with timeout.
This method provides time-bounded connection establishment using native I2P
addressing. Zero or negative timeout values disable the timeout mechanism. The
timeout only applies to the initial connection setup, not to subsequent datagram
operations. Example usage: conn, err := session.DialI2PTimeout(i2pAddress,
60*time.Second)

#### func (*DatagramSession) DialTimeout

```go
func (ds *DatagramSession) DialTimeout(destination string, timeout time.Duration) (net.PacketConn, error)
```
DialTimeout establishes a datagram connection with specified timeout duration.
This method creates a net.PacketConn interface with timeout support for
connection establishment. Zero or negative timeout values disable the timeout
mechanism. The timeout only applies to the initial connection setup, not to
subsequent operations. Example usage: conn, err :=
session.DialTimeout("destination.b32.i2p", 60*time.Second)

#### func (*DatagramSession) Listen

```go
func (s *DatagramSession) Listen() (*DatagramListener, error)
```
Listen creates a new DatagramListener for accepting incoming connections. This
method creates a listener that can accept multiple concurrent datagram
connections on the same session. Each accepted connection will be a separate
DatagramConn that shares the underlying session but has its own reader/writer.
The listener starts an accept loop in a goroutine to handle incoming
connections.

#### func (*DatagramSession) NewReader

```go
func (s *DatagramSession) NewReader() *DatagramReader
```
NewReader creates a DatagramReader for receiving datagrams from any source. This
method initializes a new reader with buffered channels for asynchronous datagram
reception. The reader must be started manually with receiveLoop() for continuous
operation. Example usage: reader := session.NewReader(); go
reader.receiveLoop(); datagram, err := reader.ReceiveDatagram()

#### func (*DatagramSession) NewWriter

```go
func (s *DatagramSession) NewWriter() *DatagramWriter
```
NewWriter creates a DatagramWriter for sending datagrams to specific
destinations. This method initializes a new writer with a default timeout of 30
seconds for send operations. The timeout can be customized using the SetTimeout
method on the returned writer. Example usage: writer :=
session.NewWriter().SetTimeout(60*time.Second); err := writer.SendDatagram(data,
dest)

#### func (*DatagramSession) PacketConn

```go
func (s *DatagramSession) PacketConn() net.PacketConn
```
PacketConn returns a net.PacketConn interface for this session. This method
provides compatibility with standard Go networking code by wrapping the datagram
session in a connection that implements the PacketConn interface. Example usage:
conn := session.PacketConn(); n, addr, err := conn.ReadFrom(buffer)

#### func (*DatagramSession) ReceiveDatagram

```go
func (s *DatagramSession) ReceiveDatagram() (*Datagram, error)
```
ReceiveDatagram receives a single datagram from any source. This is a
convenience method that creates a temporary reader, starts the receive loop,
gets one datagram, and cleans up resources automatically. For continuous
reception, use NewReader() and manage the reader lifecycle manually. Example
usage: datagram, err := session.ReceiveDatagram()

#### func (*DatagramSession) SendDatagram

```go
func (s *DatagramSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a datagram to the specified destination address. This is a
convenience method that creates a temporary writer and sends the datagram
immediately. For multiple sends, it's more efficient to create a writer once and
reuse it. Example usage: err := session.SendDatagram([]byte("hello"),
destinationAddr)

#### type DatagramWriter

```go
type DatagramWriter struct {
}
```

DatagramWriter handles outgoing datagram transmission to I2P destinations. It
provides methods for sending datagrams with configurable timeouts and handles
the underlying SAM protocol communication for message delivery. The writer
supports method chaining for configuration and provides error handling for send
operations. Example usage: writer :=
session.NewWriter().SetTimeout(30*time.Second); err := writer.SendDatagram(data,
dest)

#### func (*DatagramWriter) SendDatagram

```go
func (w *DatagramWriter) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a datagram to the specified I2P destination address. This
method handles the complete datagram transmission process including data
encoding, SAM protocol communication, and response validation. It blocks until
the datagram is sent or an error occurs, respecting the configured timeout
duration. Example usage: err := writer.SendDatagram([]byte("hello world"),
destinationAddr)

#### func (*DatagramWriter) SetTimeout

```go
func (w *DatagramWriter) SetTimeout(timeout time.Duration) *DatagramWriter
```
SetTimeout sets the timeout for datagram write operations. This method
configures the maximum time to wait for datagram send operations to complete.
The timeout prevents indefinite blocking during network congestion or connection
issues. Returns the writer instance for method chaining convenience. Example
usage: writer.SetTimeout(30*time.Second).SendDatagram(data, destination)

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide datagram-specific functionality for I2P
messaging. This type extends the base SAM functionality with methods
specifically designed for datagram communication, including session creation
with various configuration options and signature types for enhanced security and
routing control. Example usage: sam := &SAM{SAM: baseSAM}; session, err :=
sam.NewDatagramSession(id, keys, options)

#### func (*SAM) NewDatagramSession

```go
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error)
```
NewDatagramSession creates a new datagram session with the SAM bridge using
default settings. This method establishes a new datagram session for UDP-like
messaging over I2P with the specified session ID, cryptographic keys, and
configuration options. It uses default signature settings and provides a simple
interface for basic datagram communication needs. Example usage: session, err :=
sam.NewDatagramSession("my-session", keys, []string{"inbound.length=1"})

#### func (*SAM) NewDatagramSessionWithPorts

```go
func (s *SAM) NewDatagramSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error)
```
NewDatagramSessionWithPorts creates a new datagram session with port
specifications. This method allows configuring specific port ranges for the
session, enabling fine-grained control over network communication ports for
advanced routing scenarios. Port configuration is useful for applications
requiring specific port mappings or firewall compatibility. Example usage:
session, err := sam.NewDatagramSessionWithPorts(id, "8080", "8081", keys,
options)

#### func (*SAM) NewDatagramSessionWithSignature

```go
func (s *SAM) NewDatagramSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*DatagramSession, error)
```
NewDatagramSessionWithSignature creates a new datagram session with custom
signature type. This method allows specifying a custom cryptographic signature
type for the session, enabling advanced security configurations beyond the
default signature algorithm. Different signature types provide various security
levels and compatibility options. Example usage: session, err :=
sam.NewDatagramSessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
