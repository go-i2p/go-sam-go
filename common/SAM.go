package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/go-i2p/i2pkeys"
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
		return i2pkeys.I2PKeys{}, fmt.Errorf("error with writing in SAM: %w", err)
	}
	buf := make([]byte, 8192)
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response for key generation")
		return i2pkeys.I2PKeys{}, fmt.Errorf("error with reading in SAM: %w", err)
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
			return i2pkeys.I2PKeys{}, fmt.Errorf("Failed to parse keys.")
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

// Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
// for a new I2P tunnel with name id, using the cypher keys specified, with the
// I2CP/streaminglib-options as specified. Extra arguments can be specified by
// setting extra to something else than []string{}.
// This sam3 instance is now a session
func (sam *SAM) NewGenericSession(style, id string, keys i2pkeys.I2PKeys, extras []string) (net.Conn, error) {
	log.WithFields(logrus.Fields{"style": style, "id": id}).Debug("Creating new generic session")
	return sam.NewGenericSessionWithSignature(style, id, keys, SIG_EdDSA_SHA512_Ed25519, extras)
}

func (sam *SAM) NewGenericSessionWithSignature(style, id string, keys i2pkeys.I2PKeys, sigType string, extras []string) (net.Conn, error) {
	log.WithFields(logrus.Fields{"style": style, "id": id, "sigType": sigType}).Debug("Creating new generic session with signature")
	return sam.NewGenericSessionWithSignatureAndPorts(style, id, "0", "0", keys, sigType, extras)
}

// Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
// for a new I2P tunnel with name id, using the cypher keys specified, with the
// I2CP/streaminglib-options as specified. Extra arguments can be specified by
// setting extra to something else than []string{}.
// This sam3 instance is now a session
func (sam *SAM) NewGenericSessionWithSignatureAndPorts(style, id, from, to string, keys i2pkeys.I2PKeys, sigType string, extras []string) (net.Conn, error) {
	log.WithFields(logrus.Fields{"style": style, "id": id, "from": from, "to": to, "sigType": sigType}).Debug("Creating new generic session with signature and ports")

	optStr := sam.SamOptionsString()
	extraStr := strings.Join(extras, " ")

	conn := sam.Conn
	fp := ""
	tp := ""
	if from != "0" {
		fp = " FROM_PORT=" + from
	}
	if to != "0" {
		tp = " TO_PORT=" + to
	}
	scmsg := []byte("SESSION CREATE STYLE=" + style + fp + tp + " ID=" + id + " DESTINATION=" + keys.String() + " " + optStr + extraStr + "\n")

	log.WithField("message", string(scmsg)).Debug("Sending SESSION CREATE message")

	for m, i := 0, 0; m != len(scmsg); i++ {
		if i == 15 {
			log.Error("Failed to write SESSION CREATE message after 15 attempts")
			conn.Close()
			return nil, fmt.Errorf("writing to SAM failed")
		}
		n, err := conn.Write(scmsg[m:])
		if err != nil {
			log.WithError(err).Error("Failed to write to SAM connection")
			conn.Close()
			return nil, fmt.Errorf("writing to connection failed: %w", err)
		}
		m += n
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response")
		conn.Close()
		return nil, fmt.Errorf("reading from connection failed: %w", err)
	}
	text := string(buf[:n])
	log.WithField("response", text).Debug("Received SAM response")
	if strings.HasPrefix(text, SESSION_OK) {
		if keys.String() != text[len(SESSION_OK):len(text)-1] {
			log.Error("SAM created a tunnel with different keys than requested")
			conn.Close()
			return nil, fmt.Errorf("SAMv3 created a tunnel with keys other than the ones we asked it for")
		}
		log.Debug("Successfully created new session")
		return conn, nil //&StreamSession{id, conn, keys, nil, sync.RWMutex{}, nil}, nil
	} else if text == SESSION_DUPLICATE_ID {
		log.Error("Duplicate tunnel name")
		conn.Close()
		return nil, fmt.Errorf("Duplicate tunnel name")
	} else if text == SESSION_DUPLICATE_DEST {
		log.Error("Duplicate destination")
		conn.Close()
		return nil, fmt.Errorf("Duplicate destination")
	} else if text == SESSION_INVALID_KEY {
		log.Error("Invalid key for SAM session")
		conn.Close()
		return nil, fmt.Errorf("Invalid key - SAM session")
	} else if strings.HasPrefix(text, SESSION_I2P_ERROR) {
		log.WithField("error", text[len(SESSION_I2P_ERROR):]).Error("I2P error")
		conn.Close()
		return nil, fmt.Errorf("I2P error " + text[len(SESSION_I2P_ERROR):])
	} else {
		log.WithField("reply", text).Error("Unable to parse SAMv3 reply")
		conn.Close()
		return nil, fmt.Errorf("Unable to parse SAMv3 reply: " + text)
	}
}

// close this sam session
func (sam *SAM) Close() error {
	if sam.Conn != nil {
		log.Debug("Closing SAM session")
		return sam.Conn.Close()
	}
	return nil
}
