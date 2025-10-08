// Package datagram3 provides repliable datagram sessions with hash-based source identification for I2P.
//
// DATAGRAM3 sessions provide repliable UDP-like messaging with hash-based source identification
// instead of full destinations. Sources are not cryptographically authenticated; applications
// requiring authenticated sources should use datagram2 instead.
//
// Key features:
//   - Repliable (can send replies to sender)
//   - Hash-based source identification (32-byte hash)
//   - No source authentication (spoofable)
//   - Requires NAMING LOOKUP for replies
//   - UDP-like messaging (unreliable, unordered)
//   - Maximum 31744 bytes per datagram (11 KB recommended)
//
// Session creation requires 2-5 minutes for I2P tunnel establishment. Use generous timeouts
// and exponential backoff retry logic. Hash resolution uses automatic caching to minimize
// NAMING LOOKUP overhead.
//
// Basic usage:
//
//	sam, err := common.NewSAM("127.0.0.1:7656")
//	session, err := datagram3.NewDatagram3Session(sam, "my-session", keys, []string{"inbound.length=1"})
//	defer session.Close()
//	dg, err := session.NewReader().ReceiveDatagram()
//	if err := dg.ResolveSource(session); err != nil { log.Error(err) }
//	session.NewWriter().SendDatagram([]byte("reply"), dg.Source)
//
// See also: Package datagram (legacy, authenticated), datagram2 (authenticated with replay
// protection), stream (TCP-like), raw (non-repliable), primary (multi-session management).
package datagram3
