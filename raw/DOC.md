# raw
--
    import "github.com/go-i2p/go-sam-go/raw"


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
Network returns the network type

#### func (*RawAddr) String

```go
func (a *RawAddr) String() string
```
String returns the string representation of the address

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
Close closes the raw connection

#### func (*RawConn) LocalAddr

```go
func (c *RawConn) LocalAddr() net.Addr
```
LocalAddr returns the local address

#### func (*RawConn) Read

```go
func (c *RawConn) Read(b []byte) (n int, err error)
```
Read implements net.Conn by wrapping ReadFrom

#### func (*RawConn) ReadFrom

```go
func (c *RawConn) ReadFrom(p []byte) (n int, addr net.Addr, err error)
```
ReadFrom reads a raw datagram from the connection

#### func (*RawConn) RemoteAddr

```go
func (c *RawConn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote address of the connection

#### func (*RawConn) SetDeadline

```go
func (c *RawConn) SetDeadline(t time.Time) error
```
SetDeadline sets the read and write deadlines

#### func (*RawConn) SetReadDeadline

```go
func (c *RawConn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future ReadFrom calls

#### func (*RawConn) SetWriteDeadline

```go
func (c *RawConn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future WriteTo calls

#### func (*RawConn) Write

```go
func (c *RawConn) Write(b []byte) (n int, err error)
```
Write implements net.Conn by wrapping WriteTo

#### func (*RawConn) WriteTo

```go
func (c *RawConn) WriteTo(p []byte, addr net.Addr) (n int, err error)
```
WriteTo writes a raw datagram to the specified address

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
Accept waits for and returns the next raw connection to the listener

#### func (*RawListener) Addr

```go
func (l *RawListener) Addr() net.Addr
```
Addr returns the listener's network address

#### func (*RawListener) Close

```go
func (l *RawListener) Close() error
```
Close closes the raw listener

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

#### func  NewRawSession

```go
func NewRawSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error)
```
NewRawSession creates a new raw session

#### func (*RawSession) Addr

```go
func (s *RawSession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this session

#### func (*RawSession) Close

```go
func (s *RawSession) Close() error
```
Close closes the raw session and all associated resources

#### func (*RawSession) Dial

```go
func (rs *RawSession) Dial(destination string) (net.PacketConn, error)
```
Dial establishes a raw connection to the specified destination

#### func (*RawSession) DialContext

```go
func (rs *RawSession) DialContext(ctx context.Context, destination string) (net.PacketConn, error)
```
DialContext establishes a raw connection with context support

#### func (*RawSession) DialI2P

```go
func (rs *RawSession) DialI2P(addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2P establishes a raw connection to an I2P address

#### func (*RawSession) DialI2PContext

```go
func (rs *RawSession) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2PContext establishes a raw connection to an I2P address with context
support

#### func (*RawSession) DialI2PTimeout

```go
func (rs *RawSession) DialI2PTimeout(addr i2pkeys.I2PAddr, timeout time.Duration) (net.PacketConn, error)
```
DialI2PTimeout establishes a raw connection to an I2P address with timeout

#### func (*RawSession) DialTimeout

```go
func (rs *RawSession) DialTimeout(destination string, timeout time.Duration) (net.PacketConn, error)
```
DialTimeout establishes a raw connection with a timeout

#### func (*RawSession) Listen

```go
func (s *RawSession) Listen() (*RawListener, error)
```

#### func (*RawSession) NewReader

```go
func (s *RawSession) NewReader() *RawReader
```
NewReader creates a RawReader for receiving raw datagrams

#### func (*RawSession) NewWriter

```go
func (s *RawSession) NewWriter() *RawWriter
```
NewWriter creates a RawWriter for sending raw datagrams

#### func (*RawSession) PacketConn

```go
func (s *RawSession) PacketConn() net.PacketConn
```
PacketConn returns a net.PacketConn interface for this session

#### func (*RawSession) ReceiveDatagram

```go
func (s *RawSession) ReceiveDatagram() (*RawDatagram, error)
```
ReceiveDatagram receives a raw datagram from any source

#### func (*RawSession) SendDatagram

```go
func (s *RawSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a raw datagram to the specified destination

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
SendDatagram sends a raw datagram to the specified destination

#### func (*RawWriter) SetTimeout

```go
func (w *RawWriter) SetTimeout(timeout time.Duration) *RawWriter
```
SetTimeout sets the timeout for raw datagram operations

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide raw-specific functionality

#### func (*SAM) NewRawSession

```go
func (s *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error)
```
NewRawSession creates a new raw session with the SAM bridge

#### func (*SAM) NewRawSessionWithPorts

```go
func (s *SAM) NewRawSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*RawSession, error)
```
NewRawSessionWithPorts creates a new raw session with port specifications

#### func (*SAM) NewRawSessionWithSignature

```go
func (s *SAM) NewRawSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*RawSession, error)
```
NewRawSessionWithSignature creates a new raw session with custom signature type
