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

Datagram represents an I2P datagram message

#### type DatagramAddr

```go
type DatagramAddr struct {
}
```

DatagramAddr implements net.Addr for I2P datagram addresses

#### func (*DatagramAddr) Network

```go
func (a *DatagramAddr) Network() string
```
Network returns the network type

#### func (*DatagramAddr) String

```go
func (a *DatagramAddr) String() string
```
String returns the string representation of the address

#### type DatagramConn

```go
type DatagramConn struct {
}
```

DatagramConn implements net.PacketConn for I2P datagrams

#### func (*DatagramConn) Close

```go
func (c *DatagramConn) Close() error
```
Close closes the datagram connection

#### func (*DatagramConn) LocalAddr

```go
func (c *DatagramConn) LocalAddr() net.Addr
```
LocalAddr returns the local address

#### func (*DatagramConn) Read

```go
func (c *DatagramConn) Read(b []byte) (n int, err error)
```
Read implements net.Conn by wrapping ReadFrom. It reads data into the provided
byte slice and returns the number of bytes read. When reading, it also updates
the remote address of the connection. Note: This is not a typical use case for
datagrams, as they are connectionless. However, for compatibility with net.Conn,
we implement it this way.

#### func (*DatagramConn) ReadFrom

```go
func (c *DatagramConn) ReadFrom(p []byte) (n int, addr net.Addr, err error)
```
ReadFrom reads a datagram from the connection

#### func (*DatagramConn) RemoteAddr

```go
func (c *DatagramConn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote address of the connection. For datagram
connections, this returns nil as there is no single remote address.

#### func (*DatagramConn) SetDeadline

```go
func (c *DatagramConn) SetDeadline(t time.Time) error
```
SetDeadline sets the read and write deadlines

#### func (*DatagramConn) SetReadDeadline

```go
func (c *DatagramConn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future ReadFrom calls

#### func (*DatagramConn) SetWriteDeadline

```go
func (c *DatagramConn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future WriteTo calls

#### func (*DatagramConn) Write

```go
func (c *DatagramConn) Write(b []byte) (n int, err error)
```
Write implements net.Conn by wrapping WriteTo. It writes data to the remote
address and returns the number of bytes written. It uses the remote address set
by the last Read operation. If no remote address is set, it returns an error.
Note: This is not a typical use case for datagrams, as they are connectionless.
However, for compatibility with net.Conn, we implement it this way.

#### func (*DatagramConn) WriteTo

```go
func (c *DatagramConn) WriteTo(p []byte, addr net.Addr) (n int, err error)
```
WriteTo writes a datagram to the specified address

#### type DatagramListener

```go
type DatagramListener struct {
}
```

DatagramListener implements net.DatagramListener for I2P datagram connections

#### func (*DatagramListener) Accept

```go
func (l *DatagramListener) Accept() (net.Conn, error)
```
Accept waits for and returns the next packet connection to the listener

#### func (*DatagramListener) Addr

```go
func (l *DatagramListener) Addr() net.Addr
```
Addr returns the listener's network address

#### func (*DatagramListener) Close

```go
func (l *DatagramListener) Close() error
```
Close closes the packet listener

#### type DatagramReader

```go
type DatagramReader struct {
}
```

DatagramReader handles incoming datagram reception

#### func (*DatagramReader) Close

```go
func (r *DatagramReader) Close() error
```

#### func (*DatagramReader) ReceiveDatagram

```go
func (r *DatagramReader) ReceiveDatagram() (*Datagram, error)
```
ReceiveDatagram receives a datagram from any source

#### type DatagramSession

```go
type DatagramSession struct {
	*common.BaseSession
}
```

DatagramSession represents a datagram session that can send and receive
datagrams

#### func  NewDatagramSession

```go
func NewDatagramSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error)
```
NewDatagramSession creates a new datagram session

#### func (*DatagramSession) Addr

```go
func (s *DatagramSession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this session

#### func (*DatagramSession) Close

```go
func (s *DatagramSession) Close() error
```
Close closes the datagram session and all associated resources

#### func (*DatagramSession) Dial

```go
func (ds *DatagramSession) Dial(destination string) (net.PacketConn, error)
```
Dial establishes a datagram connection to the specified destination

#### func (*DatagramSession) DialContext

```go
func (ds *DatagramSession) DialContext(ctx context.Context, destination string) (net.PacketConn, error)
```
DialContext establishes a datagram connection with context support

#### func (*DatagramSession) DialI2P

```go
func (ds *DatagramSession) DialI2P(addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2P establishes a datagram connection to an I2P address

#### func (*DatagramSession) DialI2PContext

```go
func (ds *DatagramSession) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (net.PacketConn, error)
```
DialI2PContext establishes a datagram connection to an I2P address with context
support

#### func (*DatagramSession) DialI2PTimeout

```go
func (ds *DatagramSession) DialI2PTimeout(addr i2pkeys.I2PAddr, timeout time.Duration) (net.PacketConn, error)
```
DialI2PTimeout establishes a datagram connection to an I2P address with timeout

#### func (*DatagramSession) DialTimeout

```go
func (ds *DatagramSession) DialTimeout(destination string, timeout time.Duration) (net.PacketConn, error)
```
DialTimeout establishes a datagram connection with a timeout

#### func (*DatagramSession) Listen

```go
func (s *DatagramSession) Listen() (*DatagramListener, error)
```

#### func (*DatagramSession) NewReader

```go
func (s *DatagramSession) NewReader() *DatagramReader
```
NewReader creates a DatagramReader for receiving datagrams

#### func (*DatagramSession) NewWriter

```go
func (s *DatagramSession) NewWriter() *DatagramWriter
```
NewWriter creates a DatagramWriter for sending datagrams

#### func (*DatagramSession) PacketConn

```go
func (s *DatagramSession) PacketConn() net.PacketConn
```
PacketConn returns a net.PacketConn interface for this session

#### func (*DatagramSession) ReceiveDatagram

```go
func (s *DatagramSession) ReceiveDatagram() (*Datagram, error)
```
ReceiveDatagram receives a datagram from any source

#### func (*DatagramSession) SendDatagram

```go
func (s *DatagramSession) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a datagram to the specified destination

#### type DatagramWriter

```go
type DatagramWriter struct {
}
```

DatagramWriter handles outgoing datagram transmission

#### func (*DatagramWriter) SendDatagram

```go
func (w *DatagramWriter) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a datagram to the specified destination

#### func (*DatagramWriter) SetTimeout

```go
func (w *DatagramWriter) SetTimeout(timeout time.Duration) *DatagramWriter
```
SetTimeout sets the timeout for datagram operations

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide datagram-specific functionality

#### func (*SAM) NewDatagramSession

```go
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error)
```
NewDatagramSession creates a new datagram session with the SAM bridge

#### func (*SAM) NewDatagramSessionWithPorts

```go
func (s *SAM) NewDatagramSessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*DatagramSession, error)
```
NewDatagramSessionWithPorts creates a new datagram session with port
specifications

#### func (*SAM) NewDatagramSessionWithSignature

```go
func (s *SAM) NewDatagramSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*DatagramSession, error)
```
NewDatagramSessionWithSignature creates a new datagram session with custom
signature type
