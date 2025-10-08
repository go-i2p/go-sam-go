// Package stream provides TCP-like reliable connections for I2P using SAMv3 STREAM sessions.
//
// STREAM sessions provide ordered, reliable, bidirectional byte streams over I2P tunnels,
// implementing standard net.Conn and net.Listener interfaces. Ideal for applications
// requiring TCP-like semantics (HTTP servers, file transfers, persistent connections).
//
// Key features:
//   - Ordered, reliable delivery
//   - Bidirectional communication
//   - Standard net.Conn/net.Listener interfaces
//   - Automatic connection management
//   - Compatible with io.Reader/io.Writer
//
// Session creation requires 2-5 minutes for I2P tunnel establishment. Individual connections
// (Accept/Dial) require additional time for circuit building. Use generous timeouts and
// exponential backoff retry logic.
//
// Basic usage:
//
//	sam, err := common.NewSAM("127.0.0.1:7656")
//	session, err := stream.NewStreamSession(sam, "my-session", keys, []string{"inbound.length=1"})
//	defer session.Close()
//	listener, err := session.Listen()
//	conn, err := listener.Accept()
//	defer conn.Close()
//
// See also: Package datagram (UDP-like messaging), raw (unrepliable datagrams),
// primary (multi-session management).
package stream
