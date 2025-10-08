// Package datagram provides legacy authenticated datagram sessions for I2P using SAMv3 DATAGRAM.
//
// DATAGRAM sessions provide authenticated, repliable UDP-like messaging over I2P tunnels.
// This is the legacy format without replay protection. For new applications requiring replay
// protection, use package datagram2 instead.
//
// Key features:
//   - Authenticated datagrams with signature verification
//   - Repliable (can send replies to sender)
//   - No replay protection (use datagram2 if needed)
//   - UDP-like messaging (unreliable, unordered)
//   - Maximum 31744 bytes per datagram (11 KB recommended)
//   - Implements net.PacketConn interface
//
// Session creation requires 2-5 minutes for I2P tunnel establishment. Use generous timeouts
// and exponential backoff retry logic.
//
// Basic usage:
//
//	sam, err := common.NewSAM("127.0.0.1:7656")
//	session, err := datagram.NewDatagramSession(sam, "my-session", keys, []string{"inbound.length=1"})
//	defer session.Close()
//	conn := session.PacketConn()
//	n, err := conn.WriteTo(data, destination)
//	n, addr, err := conn.ReadFrom(buffer)
//
// See also: Package datagram2 (with replay protection), datagram3 (unauthenticated),
// stream (TCP-like), raw (non-repliable), primary (multi-session management).
package datagram
