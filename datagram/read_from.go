package datagram

import (
	"net"

	"github.com/go-i2p/logger"
)

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
	log.WithField("session_id", ds.ID()).Debug("ReadFrom: waiting for datagram")

	dg, err := ds.ReceiveDatagram()
	if err != nil {
		log.WithField("session_id", ds.ID()).WithError(err).Error("ReadFrom: failed to receive datagram")
		return 0, addr, err
	}

	bytesToCopy := len(dg.Data)
	truncated := false
	if len(p) < len(dg.Data) {
		bytesToCopy = len(p)
		truncated = true
		copy(p, dg.Data[:len(p)])
	} else {
		copy(p, dg.Data)
	}

	log.WithFields(logger.Fields{
		"session_id":    ds.ID(),
		"source":        dg.Source.String(),
		"bytes_read":    bytesToCopy,
		"datagram_size": len(dg.Data),
		"truncated":     truncated,
	}).Debug("ReadFrom: datagram received successfully")

	return len(dg.Data), dg.Source, nil
}
