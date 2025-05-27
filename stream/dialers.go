package stream

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// context-aware dialer, eventually...
func (s *StreamSession) DialContext(ctx context.Context, n, addr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{"network": n, "addr": addr}).Debug("DialContext called")
	return s.DialContextI2P(ctx, n, addr)
}

// context-aware dialer, eventually...
func (s *StreamSession) DialContextI2P(ctx context.Context, n, addr string) (*StreamConn, error) {
	log.WithFields(logrus.Fields{"network": n, "addr": addr}).Debug("DialContextI2P called")
	if ctx == nil {
		log.Panic("nil context")
		panic("nil context")
	}
	deadline := s.deadline(ctx, time.Now())
	if !deadline.IsZero() {
		if d, ok := ctx.Deadline(); !ok || deadline.Before(d) {
			subCtx, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()
			ctx = subCtx
		}
	}

	i2paddr, err := i2pkeys.NewI2PAddrFromString(addr)
	if err != nil {
		log.WithError(err).Error("Failed to create I2P address from string")
		return nil, err
	}
	return s.DialI2P(i2paddr)
}

// implement net.Dialer
func (s *StreamSession) Dial(n, addr string) (c net.Conn, err error) {
	log.WithFields(logrus.Fields{"network": n, "addr": addr}).Debug("Dial called")

	var i2paddr i2pkeys.I2PAddr
	var host string
	host, _, err = net.SplitHostPort(addr)
	// log.Println("Dialing:", host)
	if err = common.IgnorePortError(err); err == nil {
		// check for name
		if strings.HasSuffix(host, ".b32.i2p") || strings.HasSuffix(host, ".i2p") {
			// name lookup
			i2paddr, err = s.Lookup(host)
			log.WithFields(logrus.Fields{"host": host, "i2paddr": i2paddr}).Debug("Looked up I2P address")
		} else {
			// probably a destination
			i2paddr, err = i2pkeys.NewI2PAddrFromBytes([]byte(host))
			// i2paddr = i2pkeys.I2PAddr(host)
			// log.Println("Destination:", i2paddr, err)
			log.WithFields(logrus.Fields{"host": host, "i2paddr": i2paddr}).Debug("Created I2P address from bytes")
		}
		if err == nil {
			return s.DialI2P(i2paddr)
		}
	}
	log.WithError(err).Error("Dial failed")
	return
}

// Dials to an I2P destination and returns a SAMConn, which implements a net.Conn.
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*StreamConn, error) {
	log.WithField("addr", addr).Debug("DialI2P called")
	sam, err := common.NewSAM(s.Sam())
	if err != nil {
		log.WithError(err).Error("Failed to create new SAM instance")
		return nil, err
	}
	conn := sam.Conn
	_, err = conn.Write([]byte("STREAM CONNECT ID=" + s.ID() + s.FromPort() + s.ToPort() + " DESTINATION=" + addr.Base64() + " SILENT=false\n"))
	if err != nil {
		log.WithError(err).Error("Failed to write STREAM CONNECT command")
		conn.Close()
		return nil, err
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		log.WithError(err).Error("Failed to write STREAM CONNECT command")
		conn.Close()
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(buf[:n]))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		switch scanner.Text() {
		case "STREAM":
			continue
		case "STATUS":
			continue
		case ResultOK:
			log.Debug("Successfully connected to I2P destination")
			return &StreamConn{s.Addr(), addr, conn}, nil
		case ResultCantReachPeer:
			log.Error("Can't reach peer")
			conn.Close()
			return nil, oops.Errorf("Can not reach peer")
		case ResultI2PError:
			log.Error("I2P internal error")
			conn.Close()
			return nil, oops.Errorf("I2P internal error")
		case ResultInvalidKey:
			log.Error("Invalid key - Stream Session")
			conn.Close()
			return nil, oops.Errorf("Invalid key - Stream Session")
		case ResultInvalidID:
			log.Error("Invalid tunnel ID")
			conn.Close()
			return nil, oops.Errorf("Invalid tunnel ID")
		case ResultTimeout:
			log.Error("Connection timeout")
			conn.Close()
			return nil, oops.Errorf("Timeout")
		default:
			log.WithField("error", scanner.Text()).Error("Unknown error")
			conn.Close()
			return nil, oops.Errorf("Unknown error: %s : %s", scanner.Text(), string(buf[:n]))
		}
	}
	log.Panic("Unexpected end of StreamSession.DialI2P()")
	panic("sam3 go library error in StreamSession.DialI2P()")
}
