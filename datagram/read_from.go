package datagram

import "net"

// ReadFrom reads a datagram message from the session, storing it in p.
// It returns the number of bytes read (n), the sender's I2P address (addr),
// and any error encountered (err). The address can be used to reply to the sender.
//
// The method blocks until a datagram is available or an error occurs.
// If the provided buffer p is too small to hold the incoming datagram,
// the excess data will be discarded, and n will be set to the buffer size.
//
// Example usage:
//
//	n, addr, err := ds.ReadFrom(buf)
//	if err != nil {
//	    // handle error
//	}
//	fmt.Printf("Received %d bytes from %s\n", n, addr.Base32())
func (ds *DatagramSession) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	dg, err := ds.ReceiveDatagram()
	if err != nil {
		return 0, addr, err
	}
	if len(p) < len(dg.Data) {
		copy(p, dg.Data[:len(p)])
	} else {
		copy(p, dg.Data)
	}
	return len(dg.Data), dg.Source, nil
}
