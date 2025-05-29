# stream
--
    import "github.com/go-i2p/go-sam-go/stream"


## Usage

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide stream-specific functionality

#### func (*SAM) NewStreamSession

```go
func (s *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSession creates a new streaming session with the SAM bridge

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
signature type

#### type StreamConn

```go
type StreamConn struct {
}
```

StreamConn implements net.Conn for I2P streaming connections

#### func (*StreamConn) Close

```go
func (c *StreamConn) Close() error
```
Close closes the connection

#### func (*StreamConn) LocalAddr

```go
func (c *StreamConn) LocalAddr() net.Addr
```
LocalAddr returns the local network address

#### func (*StreamConn) Read

```go
func (c *StreamConn) Read(b []byte) (int, error)
```
Read reads data from the connection

#### func (*StreamConn) RemoteAddr

```go
func (c *StreamConn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote network address

#### func (*StreamConn) SetDeadline

```go
func (c *StreamConn) SetDeadline(t time.Time) error
```
SetDeadline sets the read and write deadlines

#### func (*StreamConn) SetReadDeadline

```go
func (c *StreamConn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future Read calls

#### func (*StreamConn) SetWriteDeadline

```go
func (c *StreamConn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future Write calls

#### func (*StreamConn) Write

```go
func (c *StreamConn) Write(b []byte) (int, error)
```
Write writes data to the connection

#### type StreamDialer

```go
type StreamDialer struct {
}
```

StreamDialer handles client-side connection establishment

#### func (*StreamDialer) Dial

```go
func (d *StreamDialer) Dial(destination string) (*StreamConn, error)
```
Dial establishes a connection to the specified destination

#### func (*StreamDialer) DialContext

```go
func (d *StreamDialer) DialContext(ctx context.Context, destination string) (*StreamConn, error)
```
DialContext establishes a connection with context support

#### func (*StreamDialer) DialI2P

```go
func (d *StreamDialer) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error)
```
DialI2P establishes a connection to the specified I2P address

#### func (*StreamDialer) DialI2PContext

```go
func (d *StreamDialer) DialI2PContext(ctx context.Context, addr i2pkeys.I2PAddr) (*StreamConn, error)
```
DialI2PContext establishes a connection to an I2P address with context support

#### func (*StreamDialer) SetTimeout

```go
func (d *StreamDialer) SetTimeout(timeout time.Duration) *StreamDialer
```
SetTimeout sets the default timeout for new dialers

#### type StreamListener

```go
type StreamListener struct {
}
```

StreamListener implements net.Listener for I2P streaming connections

#### func (*StreamListener) Accept

```go
func (l *StreamListener) Accept() (net.Conn, error)
```
Accept waits for and returns the next connection to the listener

#### func (*StreamListener) AcceptStream

```go
func (l *StreamListener) AcceptStream() (*StreamConn, error)
```
AcceptStream waits for and returns the next I2P streaming connection

#### func (*StreamListener) Addr

```go
func (l *StreamListener) Addr() net.Addr
```
Addr returns the listener's network address

#### func (*StreamListener) Close

```go
func (l *StreamListener) Close() error
```
Close closes the listener

#### type StreamSession

```go
type StreamSession struct {
	*common.BaseSession
}
```

StreamSession represents a streaming session that can create listeners and
dialers

#### func  NewStreamSession

```go
func NewStreamSession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSession creates a new streaming session

#### func (*StreamSession) Addr

```go
func (s *StreamSession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this session

#### func (*StreamSession) Close

```go
func (s *StreamSession) Close() error
```
Close closes the streaming session and all associated resources

#### func (*StreamSession) Dial

```go
func (s *StreamSession) Dial(destination string) (*StreamConn, error)
```
Dial establishes a connection to the specified I2P destination

#### func (*StreamSession) DialContext

```go
func (s *StreamSession) DialContext(ctx context.Context, destination string) (*StreamConn, error)
```
DialContext establishes a connection with context support

#### func (*StreamSession) DialI2P

```go
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error)
```
DialI2P establishes a connection to the specified I2P address

#### func (*StreamSession) Listen

```go
func (s *StreamSession) Listen() (*StreamListener, error)
```
Listen creates a StreamListener that accepts incoming connections

#### func (*StreamSession) NewDialer

```go
func (s *StreamSession) NewDialer() *StreamDialer
```
NewDialer creates a StreamDialer for establishing outbound connections
