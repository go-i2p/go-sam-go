package common

import (
	"strconv"
	"strings"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Option is a SAMEmit Option
type Option func(*SAMEmit) error

// SetType sets the type of the forwarder server
func SetType(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if s == SESSION_STYLE_STREAM ||
			s == SESSION_STYLE_DATAGRAM ||
			s == SESSION_STYLE_RAW {
			c.Style = s
			log.WithField("style", s).Debug("Set session style")
			return nil
		}
		log.WithField("style", s).Error("Invalid session style")
		return oops.Errorf("Invalid session STYLE=%s, must be STREAM, DATAGRAM, or RAW", s)
	}
}

// SetSAMAddress sets the SAM address all-at-once
func SetSAMAddress(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		sp := strings.Split(s, ":")
		if len(sp) > 2 {
			log.WithField("address", s).Error("Invalid SAM address")
			return oops.Errorf("Invalid address string: %s", sp)
		}
		if len(sp) == 2 {
			var err error
			c.I2PConfig.SamPort, err = strconv.Atoi(sp[1])
			if err != nil {
				log.WithField("port", sp[1]).Error("Invalid SAM port")
				return oops.Errorf("Invalid SAM Port %s; non-number", sp[1])
			}
		}
		c.I2PConfig.SamHost = sp[0]
		log.WithFields(logrus.Fields{
			"host": c.I2PConfig.SamHost,
			"port": c.I2PConfig.SamPort,
		}).Debug("Set SAM address")
		return nil
	}
}

// SetSAMHost sets the host of the SAMEmit's SAM bridge
func SetSAMHost(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.SamHost = s
		log.WithField("host", s).Debug("Set SAM host")
		return nil
	}
}

// SetSAMPort sets the port of the SAMEmit's SAM bridge using a string
func SetSAMPort(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		port, err := strconv.Atoi(s)
		if err != nil {
			log.WithField("port", s).Error("Invalid SAM port: non-number")
			return oops.Errorf("Invalid SAM Port %s; non-number", s)
		}
		if port < 65536 && port > -1 {
			c.I2PConfig.SamPort = port
			log.WithField("port", s).Debug("Set SAM port")
			return nil
		}
		log.WithField("port", port).Error("Invalid SAM port")
		return oops.Errorf("Invalid port")
	}
}

// SetName sets the host of the SAMEmit's SAM bridge
func SetName(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.TunName = s
		log.WithField("name", s).Debug("Set tunnel name")
		return nil
	}
}

// SetInLength sets the number of hops inbound
func SetInLength(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if u < 7 && u >= 0 {
			c.I2PConfig.InLength = u
			log.WithField("inLength", u).Debug("Set inbound tunnel length")
			return nil
		}
		log.WithField("inLength", u).Error("Invalid inbound tunnel length")
		return oops.Errorf("Invalid inbound tunnel length")
	}
}

// SetOutLength sets the number of hops outbound
func SetOutLength(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if u < 7 && u >= 0 {
			c.I2PConfig.OutLength = u
			log.WithField("outLength", u).Debug("Set outbound tunnel length")
			return nil
		}
		log.WithField("outLength", u).Error("Invalid outbound tunnel length")
		return oops.Errorf("Invalid outbound tunnel length")
	}
}

// SetInVariance sets the variance of a number of hops inbound
func SetInVariance(i int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if i < 7 && i > -7 {
			c.I2PConfig.InVariance = i
			log.WithField("inVariance", i).Debug("Set inbound tunnel variance")
			return nil
		}
		log.WithField("inVariance", i).Error("Invalid inbound tunnel variance")
		return oops.Errorf("Invalid inbound tunnel length")
	}
}

// SetOutVariance sets the variance of a number of hops outbound
func SetOutVariance(i int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if i < 7 && i > -7 {
			c.I2PConfig.OutVariance = i
			log.WithField("outVariance", i).Debug("Set outbound tunnel variance")
			return nil
		}
		log.WithField("outVariance", i).Error("Invalid outbound tunnel variance")
		return oops.Errorf("Invalid outbound tunnel variance")
	}
}

// SetInQuantity sets the inbound tunnel quantity
func SetInQuantity(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if u <= 16 && u > 0 {
			c.I2PConfig.InQuantity = u
			log.WithField("inQuantity", u).Debug("Set inbound tunnel quantity")
			return nil
		}
		log.WithField("inQuantity", u).Error("Invalid inbound tunnel quantity")
		return oops.Errorf("Invalid inbound tunnel quantity")
	}
}

// SetOutQuantity sets the outbound tunnel quantity
func SetOutQuantity(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if u <= 16 && u > 0 {
			c.I2PConfig.OutQuantity = u
			log.WithField("outQuantity", u).Debug("Set outbound tunnel quantity")
			return nil
		}
		log.WithField("outQuantity", u).Error("Invalid outbound tunnel quantity")
		return oops.Errorf("Invalid outbound tunnel quantity")
	}
}

// SetInBackups sets the inbound tunnel backups
func SetInBackups(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if u < 6 && u >= 0 {
			c.I2PConfig.InBackupQuantity = u
			log.WithField("inBackups", u).Debug("Set inbound tunnel backups")
			return nil
		}
		log.WithField("inBackups", u).Error("Invalid inbound tunnel backup quantity")
		return oops.Errorf("Invalid inbound tunnel backup quantity")
	}
}

// SetOutBackups sets the inbound tunnel backups
func SetOutBackups(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if u < 6 && u >= 0 {
			c.I2PConfig.OutBackupQuantity = u
			log.WithField("outBackups", u).Debug("Set outbound tunnel backups")
			return nil
		}
		log.WithField("outBackups", u).Error("Invalid outbound tunnel backup quantity")
		return oops.Errorf("Invalid outbound tunnel backup quantity")
	}
}

// SetEncrypt tells the router to use an encrypted leaseset
func SetEncrypt(b bool) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if b {
			c.I2PConfig.EncryptLeaseSet = b
			return nil
		}
		c.I2PConfig.EncryptLeaseSet = b
		log.WithField("encrypt", b).Debug("Set lease set encryption")
		return nil
	}
}

// SetLeaseSetKey sets the host of the SAMEmit's SAM bridge
func SetLeaseSetKey(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.LeaseSetKey = s
		log.WithField("leaseSetKey", s).Debug("Set lease set key")
		return nil
	}
}

// SetLeaseSetPrivateKey sets the host of the SAMEmit's SAM bridge
func SetLeaseSetPrivateKey(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.LeaseSetPrivateKey = s
		log.WithField("leaseSetPrivateKey", s).Debug("Set lease set private key")
		return nil
	}
}

// SetLeaseSetPrivateSigningKey sets the host of the SAMEmit's SAM bridge
func SetLeaseSetPrivateSigningKey(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.LeaseSetPrivateSigningKey = s
		log.WithField("leaseSetPrivateSigningKey", s).Debug("Set lease set private signing key")
		return nil
	}
}

// SetMessageReliability sets the host of the SAMEmit's SAM bridge
func SetMessageReliability(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.MessageReliability = s
		log.WithField("messageReliability", s).Debug("Set message reliability")
		return nil
	}
}

// SetAllowZeroIn tells the tunnel to accept zero-hop peers
func SetAllowZeroIn(b bool) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if b {
			c.I2PConfig.InAllowZeroHop = true
			return nil
		}
		c.I2PConfig.InAllowZeroHop = false
		log.WithField("allowZeroIn", b).Debug("Set allow zero-hop inbound")
		return nil
	}
}

// SetAllowZeroOut tells the tunnel to accept zero-hop peers
func SetAllowZeroOut(b bool) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if b {
			c.I2PConfig.OutAllowZeroHop = true
			return nil
		}
		c.I2PConfig.OutAllowZeroHop = false
		log.WithField("allowZeroOut", b).Debug("Set allow zero-hop outbound")
		return nil
	}
}

// SetCompress tells clients to use compression
func SetCompress(b bool) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if b {
			c.I2PConfig.UseCompression = true
			return nil
		}
		c.I2PConfig.UseCompression = false
		log.WithField("compress", b).Debug("Set compression")
		return nil
	}
}

// SetFastRecieve tells clients to use compression
func SetFastRecieve(b bool) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if b {
			c.I2PConfig.FastRecieve = true
			return nil
		}
		c.I2PConfig.FastRecieve = false
		log.WithField("fastReceive", b).Debug("Set fast receive")
		return nil
	}
}

// SetReduceIdle tells the connection to reduce it's tunnels during extended idle time.
func SetReduceIdle(b bool) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if b {
			c.I2PConfig.ReduceIdle = true
			return nil
		}
		c.I2PConfig.ReduceIdle = false
		log.WithField("reduceIdle", b).Debug("Set reduce idle")
		return nil
	}
}

// SetReduceIdleTime sets the time to wait before reducing tunnels to idle levels
func SetReduceIdleTime(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.ReduceIdleTime = 300000
		if u >= 6 {
			idleTime := (u * 60) * 1000
			c.I2PConfig.ReduceIdleTime = idleTime
			log.WithField("reduceIdleTime", idleTime).Debug("Set reduce idle time")
			return nil
		}
		log.WithField("minutes", u).Error("Invalid reduce idle timeout")
		return oops.Errorf("Invalid reduce idle timeout(Measured in minutes) %v", u)
	}
}

// SetReduceIdleTimeMs sets the time to wait before reducing tunnels to idle levels in milliseconds
func SetReduceIdleTimeMs(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.ReduceIdleTime = 300000
		if u >= 300000 {
			c.I2PConfig.ReduceIdleTime = u
			log.WithField("reduceIdleTimeMs", u).Debug("Set reduce idle time in milliseconds")
			return nil
		}
		log.WithField("milliseconds", u).Error("Invalid reduce idle timeout")
		return oops.Errorf("Invalid reduce idle timeout(Measured in milliseconds) %v", u)
	}
}

// SetReduceIdleQuantity sets minimum number of tunnels to reduce to during idle time
func SetReduceIdleQuantity(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if u < 5 {
			c.I2PConfig.ReduceIdleQuantity = u
			log.WithField("reduceIdleQuantity", u).Debug("Set reduce idle quantity")
			return nil
		}
		log.WithField("quantity", u).Error("Invalid reduce tunnel quantity")
		return oops.Errorf("Invalid reduce tunnel quantity")
	}
}

// SetCloseIdle tells the connection to close it's tunnels during extended idle time.
func SetCloseIdle(b bool) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if b {
			c.I2PConfig.CloseIdle = true
			return nil
		}
		c.I2PConfig.CloseIdle = false
		return nil
	}
}

// SetCloseIdleTime sets the time to wait before closing tunnels to idle levels
func SetCloseIdleTime(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.CloseIdleTime = 300000
		if u >= 6 {
			idleTime := (u * 60) * 1000
			c.I2PConfig.CloseIdleTime = idleTime
			log.WithFields(logrus.Fields{
				"minutes":      u,
				"milliseconds": idleTime,
			}).Debug("Set close idle time")
			return nil
		}
		log.WithField("minutes", u).Error("Invalid close idle timeout")
		return oops.Errorf("Invalid close idle timeout(Measured in minutes) %v", u)
	}
}

// SetCloseIdleTimeMs sets the time to wait before closing tunnels to idle levels in milliseconds
func SetCloseIdleTimeMs(u int) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		c.I2PConfig.CloseIdleTime = 300000
		if u >= 300000 {
			c.I2PConfig.CloseIdleTime = u
			log.WithField("closeIdleTimeMs", u).Debug("Set close idle time in milliseconds")
			return nil
		}
		return oops.Errorf("Invalid close idle timeout(Measured in milliseconds) %v", u)
	}
}

// SetAccessListType tells the system to treat the AccessList as a whitelist
func SetAccessListType(s string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if s == ACCESS_TYPE_WHITELIST {
			c.I2PConfig.AccessListType = ACCESS_TYPE_WHITELIST
			log.Debug("Set access list type to whitelist")
			return nil
		} else if s == ACCESS_TYPE_BLACKLIST {
			c.I2PConfig.AccessListType = ACCESS_TYPE_BLACKLIST
			log.Debug("Set access list type to blacklist")
			return nil
		} else if s == ACCESS_TYPE_NONE || s == "" {
			c.I2PConfig.AccessListType = ""
			log.Debug("Set access list type to none")
			return nil
		}
		return oops.Errorf("Invalid Access list type (whitelist, blacklist, none)")
	}
}

// SetAccessList tells the system to treat the AccessList as a whitelist
func SetAccessList(s []string) func(*SAMEmit) error {
	return func(c *SAMEmit) error {
		if len(s) > 0 {
			c.I2PConfig.AccessList = append(c.I2PConfig.AccessList, s...)
			log.WithField("accessList", s).Debug("Set access list")
			return nil
		}
		log.Debug("No access list set (empty list provided)")
		return nil
	}
}
