package datagram

import (
	"net"

	"github.com/go-i2p/i2pkeys"
)

// WriteTo sends a datagram message to the specified I2P address.
// It returns the number of bytes written (n) and any error encountered (err).
//
// The returned n may be less than len(p) if the message was truncated to fit the maximum datagram size limit.
//
// The error return value may indicate:
//   - SAM protocol errors (e.g., malformed commands, protocol violations, or non-OK responses)
//   - I2P network errors (e.g., tunnel not established, network unreachable, or timeouts)
//   - Application errors (e.g., nil session, invalid address type, or session not connected)
//
// The method may block until the message is sent or an error occurs.
// If the message is larger than the maximum allowed datagram size,
// it will be truncated to fit within the limit.
//
// Example usage:
//
//	n, err := ds.WriteTo([]byte("Hello, I2P!"), addr)
//	if err != nil {
//	    // handle error
//	}
//	fmt.Printf("Sent %d bytes to %s\n", n, addr.Base32())
func (ds *DatagramSession) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	switch addr := addr.(type) {
	case *i2pkeys.I2PAddr:
		// Valid I2P address type, proceed to send datagram
	case i2pkeys.I2PAddr:
		// Valid I2P address type, proceed to send datagram
	default:
		return 0, &net.OpError{
			Op:   "write",
			Net:  "i2p-datagram",
			Addr: addr,
			Err:  &net.AddrError{Err: "invalid address type", Addr: addr.String()},
		}
	}
	err = ds.SendDatagram(p, addr.(i2pkeys.I2PAddr))
	return len(p), err
}
