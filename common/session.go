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

// Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
// for a new I2P tunnel with name id, using the cypher keys specified, with the
// I2CP/streaminglib-options as specified. Extra arguments can be specified by
// setting extra to something else than []string{}.
// This sam3 instance is now a session
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

	if err := sam.configureSessionParameters(style, id, from, to, keys, sigType); err != nil {
		return nil, err
	}

	message, err := sam.buildSessionCreateMessage(extras)
	if err != nil {
		return nil, err
	}

	if err := sam.transmitSessionMessage(message); err != nil {
		return nil, err
	}

	response, err := sam.readSessionResponse()
	if err != nil {
		return nil, err
	}

	return sam.parseSessionResponse(response, id, keys)
}

// configureSessionParameters sets up the SAMEmit configuration with session parameters.
func (sam *SAM) configureSessionParameters(style, id, from, to string, keys i2pkeys.I2PKeys, sigType string) error {
	sam.SAMEmit.I2PConfig.Style = style
	sam.SAMEmit.I2PConfig.TunName = id
	sam.SAMEmit.I2PConfig.DestinationKeys = &keys
	sam.SAMEmit.I2PConfig.SigType = sigType
	sam.SAMEmit.I2PConfig.Fromport = from
	sam.SAMEmit.I2PConfig.Toport = to
	return nil
}

// buildSessionCreateMessage constructs the SESSION CREATE message with optional extras.
func (sam *SAM) buildSessionCreateMessage(extras []string) ([]byte, error) {
	baseMsg := strings.TrimSuffix(sam.SAMEmit.Create(), " \n")

	extraStr := strings.Join(extras, " ")
	if extraStr != "" {
		baseMsg += " " + extraStr
	}

	message := []byte(baseMsg + "\n")
	log.WithField("message", string(message)).Debug("Sending SESSION CREATE message " + string(message))
	return message, nil
}

// transmitSessionMessage sends the SESSION CREATE message to the SAM connection.
func (sam *SAM) transmitSessionMessage(message []byte) error {
	conn := sam.Conn
	n, err := conn.Write(message)
	if err != nil {
		log.WithError(err).Error("Failed to write to SAM connection")
		conn.Close()
		return oops.Errorf("writing to connection failed: %w", err)
	}
	if n != len(message) {
		log.WithFields(logrus.Fields{
			"written": n,
			"total":   len(message),
		}).Error("Incomplete write to SAM connection")
		conn.Close()
		return oops.Errorf("incomplete write to connection: wrote %d bytes, expected %d bytes", n, len(message))
	}
	return nil
}

// readSessionResponse reads the response from the SAM connection.
func (sam *SAM) readSessionResponse() (string, error) {
	buf := make([]byte, 4096)
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response")
		sam.Conn.Close()
		return "", oops.Errorf("reading from connection failed: %w", err)
	}
	response := string(buf[:n])
	log.WithField("response", response).Debug("Received SAM response")
	return response, nil
}

// parseSessionResponse parses the SAM response and returns the appropriate session or error.
func (sam *SAM) parseSessionResponse(response, id string, keys i2pkeys.I2PKeys) (Session, error) {
	conn := sam.Conn

	if strings.HasPrefix(response, SESSION_OK) {
		if keys.String() != response[len(SESSION_OK):len(response)-1] {
			log.Error("SAM created a tunnel with different keys than requested")
			conn.Close()
			return nil, oops.Errorf("SAMv3 created a tunnel with keys other than the ones we asked it for")
		}
		log.Debug("Successfully created new session")
		return &BaseSession{
			id:   id,
			conn: conn,
			keys: keys,
			SAM:  *sam,
		}, nil
	} else if response == SESSION_DUPLICATE_ID {
		log.Error("Duplicate tunnel name")
		conn.Close()
		return nil, oops.Errorf("Duplicate tunnel name")
	} else if response == SESSION_DUPLICATE_DEST {
		log.Error("Duplicate destination")
		conn.Close()
		return nil, oops.Errorf("Duplicate destination")
	} else if response == SESSION_INVALID_KEY {
		log.Error("Invalid key for SAM session")
		conn.Close()
		return nil, oops.Errorf("Invalid key - SAM session")
	} else if strings.HasPrefix(response, SESSION_I2P_ERROR) {
		log.WithField("error", response[len(SESSION_I2P_ERROR):]).Error("I2P error")
		conn.Close()
		return nil, oops.Errorf("I2P error %v", response[len(SESSION_I2P_ERROR):])
	} else {
		log.WithField("reply", response).Error("Unable to parse SAMv3 reply")
		conn.Close()
		return nil, oops.Errorf("Unable to parse SAMv3 reply: %v", response)
	}
}
