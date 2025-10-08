// Package raw provides encrypted but unauthenticated, non-repliable datagram sessions for I2P.
//
// RAW sessions send encrypted datagrams without source authentication or reply capability.
// Recipients cannot verify sender identity or send replies. Suitable for one-way broadcast
// scenarios (logging, metrics, announcements) where reply capability is not needed.
//
// Key features:
//   - Encrypted transmission (confidentiality)
//   - No source authentication (spoofable)
//   - Non-repliable (recipient cannot reply)
//   - UDP-like messaging (unreliable, unordered)
//   - Maximum 31744 bytes per datagram (11 KB recommended)
//
// Session creation requires 2-5 minutes for I2P tunnel establishment. Use generous timeouts
// and exponential backoff retry logic.
//
// Basic usage:
//
//	sam, err := common.NewSAM("127.0.0.1:7656")
//	session, err := raw.NewRawSession(sam, "my-session", keys, []string{"inbound.length=1"})
//	defer session.Close()
//	conn := session.PacketConn()
//	n, err := conn.WriteTo(data, destination)
//
// See also: Package datagram (authenticated, repliable), datagram2 (with replay protection),
// datagram3 (hash-based sources), stream (TCP-like), primary (multi-session management).
package raw
