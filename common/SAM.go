package common

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

func (sam *SAM) Keys() (k *i2pkeys.I2PKeys) {
	// TODO: copy them?
	log.Debug("Retrieving SAM keys")
	k = sam.SAMEmit.I2PConfig.DestinationKeys
	return
}

// read public/private keys from an io.Reader
func (sam *SAM) ReadKeys(r io.Reader) (err error) {
	log.Debug("Reading keys from io.Reader")
	var keys i2pkeys.I2PKeys
	keys, err = i2pkeys.LoadKeysIncompat(r)
	if err == nil {
		log.Debug("Keys loaded successfully")
		sam.SAMEmit.I2PConfig.DestinationKeys = &keys
		return
	}
	log.WithError(err).Error("Failed to load keys")
	return
}

// if keyfile fname does not exist
func (sam *SAM) EnsureKeyfile(fname string) (keys i2pkeys.I2PKeys, err error) {
	log.WithError(err).Error("Failed to load keys")
	if fname == "" {
		// transient
		keys, err = sam.NewKeys()
		if err == nil {
			sam.SAMEmit.I2PConfig.DestinationKeys = &keys
			log.WithFields(logrus.Fields{
				"keys": keys,
			}).Debug("Generated new transient keys")
		}
	} else {
		// persistent
		_, err = os.Stat(fname)
		if os.IsNotExist(err) {
			// make the keys
			keys, err = sam.NewKeys()
			if err == nil {
				sam.SAMEmit.I2PConfig.DestinationKeys = &keys
				// save keys
				var f io.WriteCloser
				f, err = os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0o600)
				if err == nil {
					err = i2pkeys.StoreKeysIncompat(keys, f)
					f.Close()
					log.Debug("Generated and saved new keys")
				}
			}
		} else if err == nil {
			// we haz key file
			var f *os.File
			f, err = os.Open(fname)
			if err == nil {
				keys, err = i2pkeys.LoadKeysIncompat(f)
				if err == nil {
					sam.SAMEmit.I2PConfig.DestinationKeys = &keys
					log.Debug("Loaded existing keys from file")
				}
			}
		}
	}
	if err != nil {
		log.WithError(err).Error("Failed to ensure keyfile")
	}
	return
}

// Creates the I2P-equivalent of an IP address, that is unique and only the one
// who has the private keys can send messages from. The public keys are the I2P
// desination (the address) that anyone can send messages to.
func (sam *SAM) NewKeys(sigType ...string) (i2pkeys.I2PKeys, error) {
	log.WithField("sigType", sigType).Debug("Generating new keys")
	sigtmp := ""
	if len(sigType) > 0 {
		sigtmp = sigType[0]
	}
	if _, err := sam.Conn.Write([]byte("DEST GENERATE " + sigtmp + "\n")); err != nil {
		log.WithError(err).Error("Failed to write DEST GENERATE command")
		return i2pkeys.I2PKeys{}, oops.Errorf("error with writing in SAM: %w", err)
	}
	buf := make([]byte, 8192)
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response for key generation")
		return i2pkeys.I2PKeys{}, oops.Errorf("error with reading in SAM: %w", err)
	}
	s := bufio.NewScanner(bytes.NewReader(buf[:n]))
	s.Split(bufio.ScanWords)

	var pub, priv string
	for s.Scan() {
		text := s.Text()
		if text == "DEST" {
			continue
		} else if text == "REPLY" {
			continue
		} else if strings.HasPrefix(text, "PUB=") {
			pub = text[4:]
		} else if strings.HasPrefix(text, "PRIV=") {
			priv = text[5:]
		} else {
			log.Error("Failed to parse keys from SAM response")
			return i2pkeys.I2PKeys{}, oops.Errorf("Failed to parse keys.")
		}
	}
	log.Debug("Successfully generated new keys")
	return i2pkeys.NewKeys(i2pkeys.I2PAddr(pub), priv), nil
}

// Performs a lookup, probably this order: 1) routers known addresses, cached
// addresses, 3) by asking peers in the I2P network.
func (sam *SAM) Lookup(name string) (i2pkeys.I2PAddr, error) {
	log.WithField("name", name).Debug("Looking up address")
	return sam.SAMResolver.Resolve(name)
}

// close this sam session
func (sam *SAM) Close() error {
	if sam.Conn != nil {
		log.Debug("Closing SAM session")
		return sam.Conn.Close()
	}
	return nil
}
