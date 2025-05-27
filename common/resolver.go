package common

import (
	"bufio"
	"bytes"
	"errors"
	"strings"

	"github.com/go-i2p/i2pkeys"
)

func NewSAMResolver(parent *SAM) (*SAMResolver, error) {
	log.Debug("Creating new SAMResolver from existing SAM instance")
	var s SAMResolver
	s.SAM = parent
	return &s, nil
}

func NewFullSAMResolver(address string) (*SAMResolver, error) {
	log.WithField("address", address).Debug("Creating new full SAMResolver")
	var s SAMResolver
	// var err error
	sam, err := NewSAM(address)
	s.SAM = sam
	if err != nil {
		log.WithError(err).Error("Failed to create new SAM instance")
		return nil, err
	}
	return &s, nil
}

// Performs a lookup, probably this order: 1) routers known addresses, cached
// addresses, 3) by asking peers in the I2P network.
func (sam *SAMResolver) Resolve(name string) (i2pkeys.I2PAddr, error) {
	log.WithField("name", name).Debug("Resolving name")

	if _, err := sam.Conn.Write([]byte("NAMING LOOKUP NAME=" + name + "\r\n")); err != nil {
		log.WithError(err).Error("Failed to write to SAM connection")
		sam.Close()
		return i2pkeys.I2PAddr(""), err
	}
	buf := make([]byte, 4096)
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read from SAM connection")
		sam.Close()
		return i2pkeys.I2PAddr(""), err
	}
	if n <= 13 || !strings.HasPrefix(string(buf[:n]), "NAMING REPLY ") {
		log.Error("Failed to parse SAM response")
		return i2pkeys.I2PAddr(""), errors.New("Failed to parse.")
	}
	s := bufio.NewScanner(bytes.NewReader(buf[13:n]))
	s.Split(bufio.ScanWords)

	errStr := ""
	for s.Scan() {
		text := s.Text()
		log.WithField("text", text).Debug("Parsing SAM response token")
		// log.Println("SAM3", text)
		if text == SAM_RESULT_OK {
			continue
		} else if text == SAM_RESULT_INVALID_KEY {
			errStr += "Invalid key - resolver."
			log.Error("Invalid key in resolver")
		} else if text == SAM_RESULT_KEY_NOT_FOUND {
			errStr += "Unable to resolve " + name
			log.WithField("name", name).Error("Unable to resolve name")
		} else if text == "NAME="+name {
			continue
		} else if strings.HasPrefix(text, "VALUE=") {
			addr := i2pkeys.I2PAddr(text[6:])
			log.WithField("addr", addr).Debug("Name resolved successfully")
			return i2pkeys.I2PAddr(text[6:]), nil
		} else if strings.HasPrefix(text, "MESSAGE=") {
			errStr += " " + text[8:]
			log.WithField("message", text[8:]).Warn("Received message from SAM")
		} else {
			continue
		}
	}
	return i2pkeys.I2PAddr(""), errors.New(errStr)
}
