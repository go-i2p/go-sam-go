# raw
--
    import "github.com/go-i2p/go-sam-go/raw"

Package raw provides encrypted but unauthenticated, non-repliable datagram
sessions for I2P.

RAW sessions send encrypted datagrams without source authentication or reply
capability. Recipients cannot verify sender identity or send replies. Suitable
for one-way broadcast scenarios (logging, metrics, announcements) where reply
capability is not needed.

Key features:

    - Encrypted transmission (confidentiality)
    - No source authentication (spoofable)
    - Non-repliable (recipient cannot reply)
    - UDP-like messaging (unreliable, unordered)
    - Maximum 31744 bytes per datagram (11 KB recommended)

Session creation requires 2-5 minutes for I2P tunnel establishment. Use generous
timeouts and exponential backoff retry logic.

Basic usage:

    sam, err := common.NewSAM("127.0.0.1:7656")
    session, err := raw.NewRawSession(sam, "my-session", keys, []string{"inbound.length=1"})
    defer session.Close()
    conn := session.PacketConn()
    n, err := conn.WriteTo(data, destination)

See also: Package datagram (authenticated, repliable), datagram2 (with replay
protection), datagram3 (hash-based sources), stream (TCP-like), primary
(multi-session management).

## Usage

#### type RawAddr

```go
type RawAddr struct {
}
```

RawAddr implements net.Addr for I2P raw addresses

#### func (*RawAddr) Network

```go
func (a *RawAddr) Network() string
```
Network returns the network type for this address. This method implements the
net.Addr interface and always returns "i2p-raw" to identify this as an I2P raw
datagram address type. Example usage: network := addr.Network() // returns
"i2p-raw"

#### func (*RawAddr) String

```go
func (a *RawAddr) String() string
```
String returns the string representation of the address. This method implements
the net.Addr interface and returns the Base32 encoded representation of the I2P
address for human-readable display. Example usage: addrStr := addr.String() //
returns "abcd1234...xyz.b32.i2p"

#### type RawConn

```go
type RawConn struct {
}
```

RawConn implements net.PacketConn for I2P raw datagrams

#### func (*RawConn) Close

```go
func (c *RawConn) Close() error
```
Close closes the raw connection and cleans up associated resources. This method
is safe to call multiple times and will only perform cleanup once. The
underlying session remains open and can be used by other connections. Example
usage: defer conn.Close()

#### func (*RawConn) LocalAddr

```go
func (c *RawConn) LocalAddr() net.Addr
```
LocalAddr returns the local address of the connection. This method implements
the net.PacketConn interface and returns the I2P address of the session wrapped
in a RawAddr for compatibility with net.Addr. Example usage: addr :=
conn.LocalAddr()

#### func (*RawConn) Read

```go
func (c *RawConn) Read(b []byte) (n int, err error)
```
Read implements net.Conn by wrapping ReadFrom for stream-like operations. This
method reads data and updates the remote address from the sender, providing
compatibility with net.Conn interface expectations. Example usage: n, err :=
conn.Read(buffer)

#### func (*RawConn) ReadFrom

```go
func (c *RawConn) ReadFrom(p []byte) (n int, addr net.Addr, err error)
```
ReadFrom reads a raw datagram from the connection. This method implements the
net.PacketConn interface and blocks until a datagram is received or an error
occurs, returning the data, source address, and any error. Example usage: n,
addr, err := conn.ReadFrom(buffer)

#### func (*RawConn) RemoteAddr

```go
func (c *RawConn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote address of the connection. This method implements
the net.Conn interface and returns the address of the last sender if available,
or nil if no remote address has been established. Example usage: addr :=
conn.RemoteAddr()

#### func (*RawConn) SetDeadline

```go
func (c *RawConn) SetDeadline(t time.Time) error
```
SetDeadline sets the read and write deadlines for the connection. This method
implements the net.PacketConn interface and applies the deadline to both read
and write operations through separate deadline methods. Example usage:
conn.SetDeadline(time.Now().Add(30*time.Second))

#### func (*RawConn) SetReadDeadline

```go
func (c *RawConn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future ReadFrom calls. This method
implements the net.PacketConn interface for timeout support. Currently this is a
placeholder implementation for I2P raw datagrams. Example usage:
conn.SetReadDeadline(time.Now().Add(10*time.Second))

#### func (*RawConn) SetWriteDeadline

```go
func (c *RawConn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future WriteTo calls. This method
implements the net.PacketConn interface by configuring the writer timeout based
on the deadline duration, providing timeout support for send operations. Example
usage: conn.SetWriteDeadline(time.Now().Add(5*time.Second))

#### func (*RawConn) Write

```go
func (c *RawConn) Write(b []byte) (n int, err error)
```
Write implements net.Conn by wrapping WriteTo for stream-like operations. This
method requires a remote address to be set through prior Read operations and
provides compatibility with net.Conn interface expectations. Example usage: n,
err := conn.Write(data)

#### func (*RawConn) WriteTo

```go
func (c *RawConn) WriteTo(p []byte, addr net.Addr) (n int, err error)
```
WriteTo writes a raw datagram to the specified address. This method implements
the net.PacketConn interface and sends the data to the destination address,
returning the number of bytes written and any error. Example usage: n, err :=
conn.WriteTo(data, destAddr)

#### type RawDatagram

```go
type RawDatagram struct {
	Data   []byte
	Source i2pkeys.I2PAddr
	Local  i2pkeys.I2PAddr
}
```

RawDatagram represents an I2P raw datagram message

#### type RawListener

```go
type RawListener struct {
}
```

RawListener implements net.Listener for I2P raw connections

#### func (*RawListener) Accept

```go
func (l *RawListener) Accept() (net.Conn, error)
```
Accept waits for and returns the next raw connection to the listener. This
method implements the net.Listener interface and blocks until a connection is
available or an error occurs, returning the connection or error. Example usage:
conn, err := listener.Accept()

#### func (*RawListener) Addr

```go
func (l *RawListener) Addr() net.Addr
```
Addr returns the listener's network address. This method implements the
net.Listener interface and returns the I2P address of the session wrapped in a
RawAddr for compatibility with net.Addr. Example usage: addr := listener.Addr()

#### func (*RawListener) Close

```go
func (l *RawListener) Close() error
```
Close closes the raw listener and stops accepting new connections. This method
is safe to call multiple times and will clean up all resources including the
reader and associated channels. Example usage: defer listener.Close()

#### type RawReader

```go
type RawReader struct {
}
```

RawReader handles incoming raw datagram reception

#### func (*RawReader) Close

```go
func (r *RawReader) Close() error
```
Close closes the RawReader and stops its receive loop, cleaning up all
associated resources. This method is safe to call multiple times and will not
block if the reader is already closed.

#### func (*RawReader) ReceiveDatagram

```go
func (r *RawReader) ReceiveDatagram() (*RawDatagram, error)
```
ReceiveDatagram receives a raw datagram from any source

#### type RawSession

```go
type RawSession struct {
	*common.BaseSession
}
```

RawSession represents a raw session that can send and receive raw datagrams
using SAMv3 UDP forwarding. V1/V2 TCP control socket reading is no longer
supported.

#### func  NewRawSession

```go
func NewRawSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error)
```
NewRawSession creates a new raw session for sending and receiving raw datagrams
using SAMv3 UDP forwarding. It initializes the session with the provided SAM
connection, session ID, cryptographic keys, and configuration options. It
automatically creates a UDP listener for receiving forwarded datagrams (SAMv3
requirement) and configures the session with PORT/HOST parameters. V1/V2
compatibility (reading from TCP control socket) is no longer supported. Returns
a RawSession instance that uses UDP forwarding for all raw datagram reception.
Example usage: session, err := NewRawSession(sam, "my-session", keys,
[]string{"inbound.length=1"})

#### func  NewRawSessionFromSubsession

```go
func NewRawSessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, udpConn *net.UDPConn) (*RawSession, error)
```
NewRawSessionFromSubsession creates a RawSession for a subsession that has
already been registered with a PRIMARY session using SESSION ADD. This
constructor skips the session creation step since the subsession is already
registered with the SAM bridge.

This function is specifically designed for use with SAMv3.3 PRIMARY sessions
where subsessions are created using SESSION ADD rather than SESSION CREATE
commands.

For PRIMARY raw subsessions, UDP forwarding is mandatory (SAMv3 requirement).
The UDP connection must be provided for proper raw datagram reception via UDP
forwarding.

Parameters:

    - sam: SAM connection for data operations (separate from the primary session's control connection)
    - id: The subsession ID that was already registered with SESSION ADD
    - keys: The I2P keys from the primary session (shared across all subsessions)
    - options: Configuration options for the subsession
    - udpConn: UDP connection for receiving forwarded raw datagrams (required, not nil)

Returns a RawSession ready for use without attempting to create a new SAM
session.

#### func (*RawSession) Addr

```go
func (s *RawSession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this session. This address can be used by other
I2P nodes to send datagrams to this session. The address is derived from the
session's cryptographic keys. Example usage: addr := session.Addr()

#### func (*RawSession) Close

```go
func (s *RawSession) Close() error
```
Close closes the raw session and all associated resources. This method safely
terminates the session, closes the UDP listener and underlying connection, and
cleans up any background goroutines. It's safe to call multiple times. All
readers and writers created from this session will become invalid after closing.
Example usage: defer session.Close()

#### func (*RawSession) Dial

```go
func (rs *RawSession) Dial(destination string) (net.PacketConn, error)
```
Dial establishes a raw connection to the specified I2P destination address. This
method creates a net.PacketConn interface for sending and receiving raw
datagrams with the specified destination. It uses a default timeout of 30
seconds. Dial establishes a raw connection to the specified destination

#### func (*RawSession) DialContext

```go
func (rs *RawSession) DialContext(ctx context.Context, destination string) (net.PacketConn, error)
```
DialContext establishes a raw connection with context support for cancellation.
This method provides the core dialing functionality with context-based
cancellation support, allowing for proper resource cleanup and operation
cancellation through the provided context. DialContext establishes a raw
connection with context support

#### func (*RawSession) DialI2P

```go
func (rs *RawSession) DialI2P(addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2P establishes a raw connection to an I2P address using native I2P
addressing. This method creates a net.PacketConn interface for communicating
with the specified I2P address using the native i2pkeys.I2PAddr type. It uses a
default timeout of 30 seconds. DialI2P establishes a raw connection to an I2P
address

#### func (*RawSession) DialI2PContext

```go
func (rs *RawSession) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2PContext establishes a raw connection to an I2P address with context
support. This method provides the core I2P dialing functionality with
context-based cancellation, allowing for proper resource cleanup and operation
cancellation through the provided context. DialI2PContext establishes a raw
connection to an I2P address with context support

#### func (*RawSession) DialI2PTimeout

```go
func (rs *RawSession) DialI2PTimeout(addr i2pkeys.I2PAddr, timeout time.Duration) (net.PacketConn, error)
```
DialI2PTimeout establishes a raw connection to an I2P address with timeout
support. This method provides time-bounded connection establishment using native
I2P addressing. Zero or negative timeout values disable the timeout mechanism.
DialI2PTimeout establishes a raw connection to an I2P address with timeout

#### func (*RawSession) DialTimeout

```go
func (rs *RawSession) DialTimeout(destination string, timeout time.Duration) (net.PacketConn, error)
```
DialTimeout establishes a raw connection with a specified timeout duration. This
method creates a net.PacketConn interface with timeout support, allowing for
time-bounded connection establishment. Zero or negative timeout values disable
the timeout. DialTimeout establishes a raw connection with a timeout

#### func (*RawSession) Listen

```go
func (s *RawSession) Listen() (*RawListener, error)
```
Listen creates a RawListener for accepting incoming raw connections. This method
initializes the listener with buffered channels for incoming connections and
starts the accept loop in a background goroutine to handle incoming datagrams.
Example usage: listener, err := session.Listen()

#### func (*RawSession) NewReader

```go
func (s *RawSession) NewReader() *RawReader
```
NewReader creates a RawReader for receiving raw datagrams from any source. It
initializes buffered channels for incoming datagrams and errors, returning nil
if the session is closed. The caller must start the receive loop manually by
calling receiveLoop() in a goroutine. Example usage: reader :=
session.NewReader(); go reader.receiveLoop()

#### func (*RawSession) NewWriter

```go
func (s *RawSession) NewWriter() *RawWriter
```
NewWriter creates a RawWriter for sending raw datagrams to specific
destinations. It initializes the writer with a default timeout of 30 seconds for
send operations. The timeout can be customized using the SetTimeout method on
the returned writer. Example usage: writer :=
session.NewWriter().SetTimeout(60*time.Second)

#### func (*RawSession) PacketConn

```go
func (s *RawSession) PacketConn() net.PacketConn
```
PacketConn returns a net.PacketConn interface for this session. This provides
compatibility with standard Go networking code by wrapping the session in a
RawConn that implements the PacketConn interface for datagram operations.
Example usage: conn := session.PacketConn(); n, addr, err := conn.ReadFrom(buf)

#### func (*RawSession) ReceiveDatagram

```go
func (s *RawSession) ReceiveDatagram() (*RawDatagram, error)
```
ReceiveDatagram receives a single raw datagram from any source using SAMv3 UDP
forwarding. This method performs a direct UDP read without creating a reader or
receive loop. V1/V2 TCP control socket reading is no longer supported. Example
usage: datagram, err := session.ReceiveDatagram()

#### func (*RawSession) SendDatagram

```go
func (s *RawSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a raw datagram to the specified destination address. This is
a convenience method that creates a temporary writer and sends the datagram
immediately. For multiple sends, it's more efficient to create a writer once and
reuse it. Example usage: err := session.SendDatagram(data, destAddr)

#### type RawWriter

```go
type RawWriter struct {
}
```

RawWriter handles outgoing raw datagram transmission

#### func (*RawWriter) SendDatagram

```go
func (w *RawWriter) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a raw datagram to the specified destination. This method
handles the complete send operation including data encoding, SAM protocol
communication, and response parsing for error handling. Example usage: err :=
writer.SendDatagram([]byte("hello"), destAddr)

#### func (*RawWriter) SetTimeout

```go
func (w *RawWriter) SetTimeout(timeout time.Duration) *RawWriter
```
SetTimeout sets the timeout for raw datagram operations. This method configures
the maximum time to wait for send operations to complete. It returns the writer
instance for method chaining. Example usage:
writer.SetTimeout(30*time.Second).SendDatagram(data, dest)

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide raw-specific functionality for creating and
managing raw datagram sessions. This type extends the base SAM functionality
with methods specifically designed for raw I2P datagram communication. SAM wraps
common.SAM to provide raw-specific functionality

#### func (*SAM) NewRawSession

```go
func (s *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error)
```
NewRawSession creates a new raw session with the SAM bridge using default
settings. This method establishes a new raw datagram session with the specified
ID, keys, and options. Raw sessions enable unencrypted datagram transmission
over the I2P network. NewRawSession creates a new raw session with the SAM
bridge

#### func (*SAM) NewRawSessionWithPorts

```go
func (s *SAM) NewRawSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error)
```
NewRawSessionWithPorts creates a new raw session with port specifications. This
method allows configuring specific port ranges for the session, enabling
fine-grained control over network communication ports for advanced routing
scenarios. NewRawSessionWithPorts creates a new raw session with port
specifications

#### func (*SAM) NewRawSessionWithSignature

```go
func (s *SAM) NewRawSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*RawSession, error)
```
NewRawSessionWithSignature creates a new raw session with custom signature type.
This method allows specifying a custom cryptographic signature type for the
session, enabling advanced security configurations beyond the default signature
algorithm. NewRawSessionWithSignature creates a new raw session with custom
signature type
