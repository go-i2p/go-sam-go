package primary

import (
	"net"
	"strings"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

func (sam *PrimarySession) Dial(network, addr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{"network": network, "addr": addr}).Debug("Dial() called")
	if network == "udp" || network == "udp4" || network == "udp6" {
		// return sam.DialUDPI2P(network, network+addr[0:4], addr)
		return sam.DialUDPI2P(network, network+addr[0:4], addr)
	}
	if network == "tcp" || network == "tcp4" || network == "tcp6" {
		// return sam.DialTCPI2P(network, network+addr[0:4], addr)
		return sam.DialTCPI2P(network, network+addr[0:4], addr)
	}
	log.WithField("network", network).Error("Invalid network type")
	return nil, oops.Errorf("Error: Must specify a valid network type")
}

// DialTCP implements x/dialer
func (sam *PrimarySession) DialTCP(network string, laddr, raddr net.Addr) (net.Conn, error) {
	log.WithFields(logrus.Fields{"network": network, "laddr": laddr, "raddr": raddr}).Debug("DialTCP() called")
	ts, ok := sam.stsess[network+raddr.String()[0:4]]
	var err error
	if !ok {
		ts, err = sam.NewUniqueStreamSubSession(network + raddr.String()[0:4])
		if err != nil {
			log.WithError(err).Error("Failed to create new unique stream sub-session")
			return nil, err
		}
		sam.stsess[network+raddr.String()[0:4]] = ts
		ts, _ = sam.stsess[network+raddr.String()[0:4]]
	}
	return ts.Dial(network, raddr.String())
}

func (sam *PrimarySession) DialTCPI2P(network, laddr, raddr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{"network": network, "laddr": laddr, "raddr": raddr}).Debug("DialTCPI2P() called")
	ts, ok := sam.stsess[network+raddr[0:4]]
	var err error
	if !ok {
		ts, err = sam.NewUniqueStreamSubSession(network + laddr)
		if err != nil {
			log.WithError(err).Error("Failed to create new unique stream sub-session")
			return nil, err
		}
		sam.stsess[network+raddr[0:4]] = ts
		ts, _ = sam.stsess[network+raddr[0:4]]
	}
	return ts.Dial(network, raddr)
}

// DialUDP implements x/dialer
func (sam *PrimarySession) DialUDP(network string, laddr, raddr net.Addr) (net.PacketConn, error) {
	log.WithFields(logrus.Fields{"network": network, "laddr": laddr, "raddr": raddr}).Debug("DialUDP() called")
	ds, ok := sam.dgsess[network+raddr.String()[0:4]]
	var err error
	if !ok {
		ds, err = sam.NewDatagramSubSession(network+raddr.String()[0:4], 0)
		if err != nil {
			log.WithError(err).Error("Failed to create new datagram sub-session")
			return nil, err
		}
		sam.dgsess[network+raddr.String()[0:4]] = ds
		ds, _ = sam.dgsess[network+raddr.String()[0:4]]
	}
	return ds.Dial(network, raddr.String())
}

func (sam *PrimarySession) DialUDPI2P(network, laddr, raddr string) (*datagram.DatagramSession, error) {
	log.WithFields(logrus.Fields{"network": network, "laddr": laddr, "raddr": raddr}).Debug("DialUDPI2P() called")
	ds, ok := sam.dgsess[network+raddr[0:4]]
	var err error
	if !ok {
		ds, err = sam.NewDatagramSubSession(network+laddr, 0)
		if err != nil {
			log.WithError(err).Error("Failed to create new datagram sub-session")
			return nil, err
		}
		sam.dgsess[network+raddr[0:4]] = ds
		ds, _ = sam.dgsess[network+raddr[0:4]]
	}
	return ds.Dial(network, raddr)
}

func (s *PrimarySession) Lookup(name string) (a net.Addr, err error) {
	log.WithField("name", name).Debug("Lookup() called")
	var sam *common.SAM
	name = strings.Split(name, ":")[0]
	sam, err = common.NewSAM(s.samAddr)
	if err == nil {
		log.WithField("addr", a).Debug("Lookup successful")
		defer sam.Close()
		a, err = sam.Lookup(name)
	}
	log.WithError(err).Error("Lookup failed")
	return
}
