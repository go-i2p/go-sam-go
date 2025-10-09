# stream
--
    import "github.com/go-i2p/go-sam-go/stream"

Package stream provides TCP-like reliable connections for I2P using SAMv3 STREAM
sessions.

STREAM sessions provide ordered, reliable, bidirectional byte streams over I2P
tunnels, implementing standard net.Conn and net.Listener interfaces. Ideal for
applications requiring TCP-like semantics (HTTP servers, file transfers,
persistent connections).

Key features:

    - Ordered, reliable delivery
    - Bidirectional communication
    - Standard net.Conn/net.Listener interfaces
    - Automatic connection management
    - Compatible with io.Reader/io.Writer

Session creation requires 2-5 minutes for I2P tunnel establishment. Individual
connections (Accept/Dial) require additional time for circuit building. Use
generous timeouts and exponential backoff retry logic.

Basic usage:

    sam, err := common.NewSAM("127.0.0.1:7656")
    session, err := stream.NewStreamSession(sam, "my-session", keys, []string{"inbound.length=1"})
    defer session.Close()
    listener, err := session.Listen()
    conn, err := listener.Accept()
    defer conn.Close()

See also: Package datagram (UDP-like messaging), raw (unrepliable datagrams),
primary (multi-session management).

## Usage

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide stream-specific functionality and convenience
methods. It extends the base SAM connection with streaming-specific session
creation methods, providing a more convenient API for creating streaming
sessions without requiring direct interaction with the generic session creation
methods. Example usage: sam := &SAM{SAM: commonSAM}; session, err :=
sam.NewStreamSession("id", keys, options)

#### func (*SAM) NewStreamSession

```go
func (s *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSession creates a new streaming session with the SAM bridge using
default signature. This is a convenience method that wraps the generic session
creation with streaming-specific parameters. It uses the default Ed25519
signature type and provides a simpler API for creating streaming sessions
without requiring explicit signature type specification. Example usage: session,
err := sam.NewStreamSession("my-session", keys, options)

#### func (*SAM) NewStreamSessionWithPorts

```go
func (s *SAM) NewStreamSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSessionWithPorts creates a new streaming session with port
specifications

#### func (*SAM) NewStreamSessionWithSignature

```go
func (s *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
NewStreamSessionWithSignature creates a new streaming session with custom
signature type. This method provides advanced control over the cryptographic
signature type used for the I2P destination. It supports various signature
algorithms like Ed25519, ECDSA, and DSA, allowing applications to choose the
most appropriate signature type for their needs. Example usage: session, err :=
sam.NewStreamSessionWithSignature("my-session", keys, options,
"EdDSA_SHA512_Ed25519")

#### type StreamConn

```go
type StreamConn struct {
}
```

StreamConn implements net.Conn for I2P streaming connections. It provides a
standard Go networking interface for TCP-like reliable communication over I2P
networks. The connection supports standard read/write operations with proper
timeout handling and address information for both local and remote endpoints.
Example usage: conn, err := session.Dial("destination.b32.i2p"); data, err :=
conn.Read(buffer)

#### func (*StreamConn) Close

```go
func (c *StreamConn) Close() error
```
Close closes the connection and releases all associated resources. This method
implements the net.Conn interface and is safe to call multiple times. It
properly handles concurrent access and ensures clean shutdown of the underlying
I2P streaming connection with appropriate error handling. Example usage: defer
conn.Close()

#### func (*StreamConn) LocalAddr

```go
func (c *StreamConn) LocalAddr() net.Addr
```
LocalAddr returns the local network address of the connection. This method
implements the net.Conn interface and provides the I2P address of the local
endpoint. The returned address implements the net.Addr interface and can be used
for logging or connection management. Example usage: localAddr :=
conn.LocalAddr()

#### func (*StreamConn) Read

```go
func (c *StreamConn) Read(b []byte) (int, error)
```
Read reads data from the connection into the provided buffer. This method
implements the net.Conn interface and provides thread-safe reading from the
underlying I2P streaming connection. It handles connection state checking and
proper error reporting for closed connections. Example usage: n, err :=
conn.Read(buffer)

#### func (*StreamConn) RemoteAddr

```go
func (c *StreamConn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote network address of the connection. This method
implements the net.Conn interface and provides the I2P address of the remote
endpoint. The returned address implements the net.Addr interface and can be used
for logging, authentication, or connection management. Example usage: remoteAddr
:= conn.RemoteAddr()

#### func (*StreamConn) SetDeadline

```go
func (c *StreamConn) SetDeadline(t time.Time) error
```
SetDeadline sets the read and write deadlines for the connection. This method
implements the net.Conn interface and sets both read and write deadlines to the
same time. It provides a convenient way to set overall connection timeouts for
both read and write operations. Example usage:
conn.SetDeadline(time.Now().Add(30*time.Second))

#### func (*StreamConn) SetReadDeadline

```go
func (c *StreamConn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future Read calls on the connection. This
method implements the net.Conn interface and allows setting read-specific
timeouts. A zero time value disables the deadline, and the deadline applies to
all future and pending Read calls. Example usage:
conn.SetReadDeadline(time.Now().Add(30*time.Second))

#### func (*StreamConn) SetWriteDeadline

```go
func (c *StreamConn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future Write calls on the connection.
This method implements the net.Conn interface and allows setting write-specific
timeouts. A zero time value disables the deadline, and the deadline applies to
all future and pending Write calls. Example usage:
conn.SetWriteDeadline(time.Now().Add(30*time.Second))

#### func (*StreamConn) Write

```go
func (c *StreamConn) Write(b []byte) (int, error)
```
Write writes data to the connection from the provided buffer. This method
implements the net.Conn interface and provides thread-safe writing to the
underlying I2P streaming connection. It handles connection state checking and
proper error reporting for closed connections. Example usage: n, err :=
conn.Write(data)

#### type StreamDialer

```go
type StreamDialer struct {
}
```

StreamDialer handles client-side connection establishment for I2P streaming. It
provides methods for dialing I2P destinations with configurable timeout support
and context-based cancellation. The dialer can be configured with custom
timeouts and supports both string destinations and native I2P addresses. Example
usage: dialer := session.NewDialer().SetTimeout(60*time.Second); conn, err :=
dialer.Dial("dest.b32.i2p")

#### func (*StreamDialer) Dial

```go
func (d *StreamDialer) Dial(destination string) (*StreamConn, error)
```
Dial establishes a connection to the specified destination using the default
context. This method resolves the destination string and establishes a streaming
connection using the dialer's configured timeout. It provides a simple interface
for connection establishment without requiring explicit context management.
Example usage: conn, err := dialer.Dial("destination.b32.i2p")

#### func (*StreamDialer) DialContext

```go
func (d *StreamDialer) DialContext(ctx context.Context, destination string) (*StreamConn, error)
```
DialContext establishes a connection with context support for cancellation and
timeout. This method resolves the destination string and establishes a streaming
connection with context-based cancellation support. The context can override the
dialer's default timeout and provides fine-grained control over connection
establishment. Example usage: conn, err := dialer.DialContext(ctx,
"destination.b32.i2p")

#### func (*StreamDialer) DialI2P

```go
func (d *StreamDialer) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error)
```
DialI2P establishes a connection to the specified I2P address using native
addressing. This method accepts an i2pkeys.I2PAddr directly, bypassing the need
for destination resolution. It uses the dialer's configured timeout and provides
efficient connection establishment for known I2P addresses. Example usage: conn,
err := dialer.DialI2P(addr)

#### func (*StreamDialer) DialI2PContext

```go
func (d *StreamDialer) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (*StreamConn, error)
```
DialI2PContext establishes a connection to an I2P address with context support.
This method provides the core dialing functionality with context-based
cancellation and timeout support. It handles SAM protocol communication,
connection establishment, and proper resource management for streaming
connections over I2P. Example usage: conn, err := dialer.DialI2PContext(ctx,
addr)

#### func (*StreamDialer) SetTimeout

```go
func (d *StreamDialer) SetTimeout(timeout time.Duration) *StreamDialer
```
SetTimeout sets the default timeout duration for dial operations. This method
allows customization of the connection timeout and returns the dialer for method
chaining. The timeout applies to all subsequent dial operations. Example usage:
dialer.SetTimeout(60*time.Second)

#### type StreamListener

```go
type StreamListener struct {
}
```

StreamListener implements net.Listener for I2P streaming connections. It manages
incoming connection acceptance and provides thread-safe operations for accepting
connections from remote I2P destinations. The listener runs an internal accept
loop to handle incoming connections asynchronously. A finalizer is attached to
the listener to ensure that the accept loop is cleaned up if the listener is
garbage collected without being closed. Example usage: listener, err :=
session.Listen(); conn, err := listener.Accept()

#### func (*StreamListener) Accept

```go
func (l *StreamListener) Accept() (net.Conn, error)
```
Accept waits for and returns the next connection to the listener. This method
implements the net.Listener interface and provides compatibility with standard
Go networking patterns. It returns a net.Conn interface that can be used with
any Go networking code expecting standard connections. Example usage: conn, err
:= listener.Accept()

#### func (*StreamListener) AcceptStream

```go
func (l *StreamListener) AcceptStream() (*StreamConn, error)
```
AcceptStream waits for and returns the next I2P streaming connection. This
method provides I2P-specific connection acceptance, returning a StreamConn
directly rather than the generic net.Conn interface. It offers more type safety
and I2P-specific functionality compared to the generic Accept method. Example
usage: conn, err := listener.AcceptStream()

#### func (*StreamListener) Addr

```go
func (l *StreamListener) Addr() net.Addr
```
Addr returns the listener's network address. This method implements the
net.Listener interface and provides the I2P address that the listener is bound
to. The returned address implements the net.Addr interface and can be used for
logging or connection management. Example usage: addr := listener.Addr()

#### func (*StreamListener) Close

```go
func (l *StreamListener) Close() error
```
Close closes the listener and stops accepting new connections. This method
implements the net.Listener interface and is safe to call multiple times. It
properly handles concurrent access and ensures clean shutdown of the accept loop
with appropriate resource cleanup and error handling. Example usage: defer
listener.Close()

#### type StreamSession

```go
type StreamSession struct {
	*common.BaseSession
}
```

StreamSession represents a streaming session that can create listeners and
dialers. It provides TCP-like reliable connection capabilities over the I2P
network, supporting both client and server operations. The session manages the
underlying I2P connection and provides methods for creating listeners and
dialers for stream-based communication. Example usage: session, err :=
NewStreamSession(sam, "my-session", keys, options)

#### func  NewStreamSession

```go
func NewStreamSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSession creates a new streaming session for TCP-like I2P connections.
It initializes the session with the provided SAM connection, session ID,
cryptographic keys, and configuration options. The session provides both client
and server capabilities for establishing reliable streaming connections over the
I2P network. Example usage: session, err := NewStreamSession(sam, "my-session",
keys, []string{"inbound.length=1"})

#### func  NewStreamSessionFromSubsession

```go
func NewStreamSessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSessionFromSubsession creates a StreamSession for a subsession that has
already been registered with a PRIMARY session using SESSION ADD. This
constructor skips the session creation step since the subsession is already
registered with the SAM bridge.

This function is specifically designed for use with SAMv3.3 PRIMARY sessions
where subsessions are created using SESSION ADD rather than SESSION CREATE
commands.

Parameters:

    - sam: SAM connection for data operations (separate from the primary session's control connection)
    - id: The subsession ID that was already registered with SESSION ADD
    - keys: The I2P keys from the primary session (shared across all subsessions)
    - options: Configuration options for the subsession

Returns a StreamSession ready for use without attempting to create a new SAM
session.

#### func  NewStreamSessionWithSignature

```go
func NewStreamSessionWithSignature(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
NewStreamSessionWithSignature creates a new streaming session with a custom
signature type for TCP-like I2P connections. This is the package-level function
version that allows specifying cryptographic signature algorithms. It
initializes the session with the provided SAM connection, session ID,
cryptographic keys, configuration options, and signature type. The session
provides both client and server capabilities for establishing reliable streaming
connections over the I2P network with custom cryptographic settings. Example
usage: session, err := NewStreamSessionWithSignature(sam, "my-session", keys,
[]string{"inbound.length=1"}, "EdDSA_SHA512_Ed25519")

#### func  NewStreamSessionWithSignatureAndPorts

```go
func NewStreamSessionWithSignatureAndPorts(sam *common.SAM, id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
NewStreamSessionWithSignatureAndPorts creates a new stream session with custom
signature type and port configuration. This function provides advanced control
over both cryptographic parameters and port mapping for stream sessions. The
'from' parameter specifies the local port binding, while 'to' specifies the
target port for connections. Port specifications can be single ports ("80") or
ranges ("8080-8090") depending on I2P router configuration.

This method enables complex port forwarding scenarios and integration with
existing network infrastructure that expects specific port mappings. It's
particularly useful for applications that need to maintain consistent port
assignments or work with legacy systems expecting fixed port numbers.

Example usage:

    session, err := NewStreamSessionWithSignatureAndPorts(sam, "http-proxy", "8080", "80", keys,
                       []string{"inbound.length=2"}, "EdDSA_SHA512_Ed25519")

#### func (*StreamSession) Accept

```go
func (s *StreamSession) Accept() (*StreamConn, error)
```
Accept creates a listener and accepts the next incoming connection from remote
I2P destinations. This is a convenience method that automatically creates a
listener and calls Accept() on it. It provides a simpler API for applications
that only need to accept a single connection or want to handle each connection
acceptance individually.

For applications that need to accept multiple connections or want more control
over listener lifecycle, use Listen() to get a StreamListener and call Accept()
on it directly.

Each call to Accept creates a new internal listener, so applications accepting
multiple connections should use Listen() once and then call Accept() multiple
times on the listener for better performance and resource management.

Returns a StreamConn for the accepted connection, or an error if the acceptance
fails. The error may be due to session closure, network issues, or I2P tunnel
problems.

Example usage: conn, err := session.Accept() // Simple single connection
acceptance

#### func (*StreamSession) Addr

```go
func (s *StreamSession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this session for identification purposes. This
address can be used by other I2P nodes to connect to this session. The address
is derived from the session's cryptographic keys and remains constant for the
lifetime of the session. Example usage: addr := session.Addr()

#### func (*StreamSession) Close

```go
func (s *StreamSession) Close() error
```
Close closes the streaming session and all associated resources. This method is
safe to call multiple times and will only perform cleanup once. All listeners
and connections created from this session will become invalid after closing. The
method properly handles concurrent access and resource cleanup. Example usage:
defer session.Close()

#### func (*StreamSession) Dial

```go
func (s *StreamSession) Dial(destination string) (*StreamConn, error)
```
Dial establishes a connection to the specified I2P destination using the default
timeout. This is a convenience method that creates a new dialer and establishes
a connection to the specified destination string. For custom timeout or multiple
connections, use NewDialer() for better performance. Example usage: conn, err :=
session.Dial("destination.b32.i2p")

#### func (*StreamSession) DialContext

```go
func (s *StreamSession) DialContext(ctx context.Context, destination string) (*StreamConn, error)
```
DialContext establishes a connection with context support for cancellation and
timeout. This is a convenience method that creates a new dialer and establishes
a connection to the specified destination with context-based cancellation
support. The context can be used to cancel the connection attempt or apply
custom timeouts. Example usage: conn, err := session.DialContext(ctx,
"destination.b32.i2p")

#### func (*StreamSession) DialI2P

```go
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error)
```
DialI2P establishes a connection to the specified I2P address using native
addressing. This is a convenience method that creates a new dialer and
establishes a connection to the specified I2P address using the i2pkeys.I2PAddr
type. The method uses the session's default timeout settings. Example usage:
conn, err := session.DialI2P(addr)

#### func (*StreamSession) Listen

```go
func (s *StreamSession) Listen() (*StreamListener, error)
```
Listen creates a StreamListener that accepts incoming connections from remote
I2P destinations. It initializes a listener with buffered channels for
connection handling and starts an internal accept loop to manage incoming
connections asynchronously. The listener provides thread-safe operations and
properly handles session closure and resource cleanup. A finalizer is set on the
listener to ensure that the accept loop is terminated if the listener is garbage
collected without being closed. Example usage: listener, err :=
session.Listen(); conn, err := listener.Accept()

#### func (*StreamSession) NewDialer

```go
func (s *StreamSession) NewDialer() *StreamDialer
```
NewDialer creates a StreamDialer for establishing outbound connections to I2P
destinations. It initializes a dialer with a default timeout of 30 seconds,
which can be customized using the SetTimeout method. The dialer supports both
string destinations and native I2P addresses. Example usage: dialer :=
session.NewDialer().SetTimeout(60*time.Second)
