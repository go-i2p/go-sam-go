package common

import (
	"strings"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
// for a new I2P tunnel with name id, using the cypher keys specified, with the
// I2CP/streaminglib-options as specified. Extra arguments can be specified by
// setting extra to something else than []string{}.
// This sam3 instance is now a session
func (sam SAM) NewGenericSession(style, id string, keys i2pkeys.I2PKeys, extras []string) (Session, error) {
	log.WithFields(logrus.Fields{"style": style, "id": id}).Debug("Creating new generic session")
	return sam.NewGenericSessionWithSignature(style, id, keys, SIG_EdDSA_SHA512_Ed25519, extras)
}

func (sam SAM) NewGenericSessionWithSignature(style, id string, keys i2pkeys.I2PKeys, sigType string, extras []string) (Session, error) {
	log.WithFields(logrus.Fields{"style": style, "id": id, "sigType": sigType}).Debug("Creating new generic session with signature")
	return sam.NewGenericSessionWithSignatureAndPorts(style, id, "0", "0", keys, sigType, extras)
}

// Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
// for a new I2P tunnel with name id, using the cypher keys specified, with the
// I2CP/streaminglib-options as specified. Extra arguments can be specified by
// setting extra to something else than []string{}.
// This sam3 instance is now a session
func (sam SAM) NewGenericSessionWithSignatureAndPorts(style, id, from, to string, keys i2pkeys.I2PKeys, sigType string, extras []string) (Session, error) {
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

	n, err := conn.Write(scmsg)
	if err != nil {
		log.WithError(err).Error("Failed to write to SAM connection")
		conn.Close()
		return nil, oops.Errorf("writing to connection failed: %w", err)
	}
	if n != len(scmsg) {
		log.WithFields(logrus.Fields{
			"written": n,
			"total":   len(scmsg),
		}).Error("Incomplete write to SAM connection")
		conn.Close()
		return nil, oops.Errorf("incomplete write to connection: wrote %d bytes, expected %d bytes", n, len(scmsg))
	}
	buf := make([]byte, 4096)
	n, err = conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response")
		conn.Close()
		return nil, oops.Errorf("reading from connection failed: %w", err)
	}
	text := string(buf[:n])
	log.WithField("response", text).Debug("Received SAM response")
	if strings.HasPrefix(text, SESSION_OK) {
		if keys.String() != text[len(SESSION_OK):len(text)-1] {
			log.Error("SAM created a tunnel with different keys than requested")
			conn.Close()
			return nil, oops.Errorf("SAMv3 created a tunnel with keys other than the ones we asked it for")
		}
		log.Debug("Successfully created new session")
		return &BaseSession{
			id:   id,
			conn: conn,
			keys: keys,
			SAM:  sam,
		}, nil
	} else if text == SESSION_DUPLICATE_ID {
		log.Error("Duplicate tunnel name")
		conn.Close()
		return nil, oops.Errorf("Duplicate tunnel name")
	} else if text == SESSION_DUPLICATE_DEST {
		log.Error("Duplicate destination")
		conn.Close()
		return nil, oops.Errorf("Duplicate destination")
	} else if text == SESSION_INVALID_KEY {
		log.Error("Invalid key for SAM session")
		conn.Close()
		return nil, oops.Errorf("Invalid key - SAM session")
	} else if strings.HasPrefix(text, SESSION_I2P_ERROR) {
		log.WithField("error", text[len(SESSION_I2P_ERROR):]).Error("I2P error")
		conn.Close()
		return nil, oops.Errorf("I2P error " + text[len(SESSION_I2P_ERROR):])
	} else {
		log.WithField("reply", text).Error("Unable to parse SAMv3 reply")
		conn.Close()
		return nil, oops.Errorf("Unable to parse SAMv3 reply: " + text)
	}
}
