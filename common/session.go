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

	// First validate primary session options for duplicates and invalid combinations
	primaryValidated := validatePrimarySessionOptions(extras)

	// Then validate signature type conflicts and clean them if needed
	cleanedExtras := validateAndCleanOptions(sam.SAMEmit.I2PConfig.SigType, primaryValidated)

	extraStr := strings.Join(cleanedExtras, " ")
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
		return oops.Errorf("writing to connection failed: %w", err)
	}
	if n != len(message) {
		log.WithFields(logrus.Fields{
			"written": n,
			"total":   len(message),
		}).Error("Incomplete write to SAM connection")
		return oops.Errorf("incomplete write to connection: wrote %d bytes, expected %d bytes", n, len(message))
	}
	return nil
}

// readSessionResponse reads the response from the SAM connection.
// Uses dynamic buffer allocation to handle large session responses with many I2CP options.
func (sam *SAM) readSessionResponse() (string, error) {
	buf, n, err := sam.readInitialBuffer()
	if err != nil {
		return "", err
	}

	// If buffer was completely filled, there might be more data
	if n == len(buf) {
		return sam.readLargeResponse(buf, n)
	}

	response := string(buf[:n])
	log.WithField("response", response).Debug("Received SAM response")
	return response, nil
}

// readInitialBuffer performs the initial read operation from the SAM connection.
func (sam *SAM) readInitialBuffer() ([]byte, int, error) {
	buf := make([]byte, 4096) // Initial buffer size for typical responses
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response")
		return nil, 0, oops.Errorf("reading from connection failed: %w", err)
	}
	return buf, n, nil
}

// readLargeResponse handles reading responses that exceed the initial buffer size.
func (sam *SAM) readLargeResponse(initialBuf []byte, initialN int) (string, error) {
	response := make([]byte, initialN, len(initialBuf)*2)
	copy(response, initialBuf[:initialN])

	additionalData, err := sam.readAdditionalData()
	if err != nil {
		return "", err
	}

	response = append(response, additionalData...)
	responseStr := string(response)
	log.WithField("response", responseStr).Debug("Received SAM response")
	return responseStr, nil
}

// readAdditionalData reads remaining data from the SAM connection in chunks.
func (sam *SAM) readAdditionalData() ([]byte, error) {
	var result []byte

	for {
		additionalBuf := make([]byte, 2048)
		additionalN, err := sam.Conn.Read(additionalBuf)
		if err != nil {
			if additionalN == 0 {
				// Connection closed or no more data
				break
			}
			log.WithError(err).Error("Failed to read additional SAM response data")
			return nil, oops.Errorf("error reading additional SAM data: %w", err)
		}

		result = append(result, additionalBuf[:additionalN]...)

		// If we didn't fill the additional buffer, we're done
		if additionalN < len(additionalBuf) {
			break
		}
	}

	return result, nil
}

// parseSessionResponse parses the SAM response and returns the appropriate session or error.
func (sam *SAM) parseSessionResponse(response, id string, keys i2pkeys.I2PKeys) (Session, error) {
	if strings.HasPrefix(response, SESSION_OK) {
		return sam.handleSuccessResponse(response, id, keys)
	}

	return nil, sam.handleErrorResponse(response)
}

// handleSuccessResponse validates and creates a session from a successful SAM response.
func (sam *SAM) handleSuccessResponse(response, id string, keys i2pkeys.I2PKeys) (Session, error) {
	expectedKeys := response[len(SESSION_OK) : len(response)-1]
	if keys.String() != expectedKeys {
		log.Error("SAM created a tunnel with different keys than requested")
		return nil, oops.Errorf("SAMv3 created a tunnel with keys other than the ones we asked it for")
	}

	log.Debug("Successfully created new session")
	return &BaseSession{
		id:   id,
		conn: sam.Conn,
		keys: keys,
		SAM:  *sam,
	}, nil
}

// handleErrorResponse processes different SAM error responses and returns appropriate errors.
func (sam *SAM) handleErrorResponse(response string) error {
	sam.Conn.Close()

	switch {
	case response == SESSION_DUPLICATE_ID:
		log.Error("Duplicate tunnel name")
		return oops.Errorf("Duplicate tunnel name")
	case response == SESSION_DUPLICATE_DEST:
		log.Error("Duplicate destination")
		return oops.Errorf("Duplicate destination")
	case response == SESSION_INVALID_KEY:
		log.Error("Invalid key for SAM session")
		return oops.Errorf("Invalid key - SAM session")
	case strings.HasPrefix(response, SESSION_I2P_ERROR):
		return sam.handleI2PError(response)
	default:
		return sam.handleUnknownResponse(response)
	}
}

// handleI2PError processes I2P-specific error responses.
func (sam *SAM) handleI2PError(response string) error {
	errorDetail := response[len(SESSION_I2P_ERROR):]
	log.WithField("error", errorDetail).Error("I2P error")
	return oops.Errorf("I2P error %v", errorDetail)
}

// handleUnknownResponse processes unrecognized SAM responses.
func (sam *SAM) handleUnknownResponse(response string) error {
	log.WithField("reply", response).Error("Unable to parse SAMv3 reply")
	return oops.Errorf("Unable to parse SAMv3 reply: %v", response)
}

// AddSubSession adds a subsession to an existing PRIMARY session using the SESSION ADD command.
// This method implements the SAMv3.3 protocol for creating subsessions that share the same
// destination and tunnels as the primary session while providing separate protocol handling.
//
// Parameters:
//   - style: Session style ("STREAM", "DATAGRAM", or "RAW")
//   - id: Unique subsession identifier within the primary session scope
//   - options: Additional SAM protocol options for the subsession
//
// The subsession inherits the destination from the primary session and uses the same
// tunnel infrastructure for enhanced efficiency. Each subsession must have a unique
// combination of style and port to enable proper routing of incoming traffic.
//
// Example usage:
//
//	err := sam.AddSubSession("STREAM", "stream-sub-1", []string{"FROM_PORT=8080"})
func (sam *SAM) AddSubSession(style, id string, options []string) error {
	log.WithFields(logrus.Fields{
		"style":   style,
		"id":      id,
		"options": options,
	}).Debug("Adding subsession to primary session")

	message, err := sam.buildSessionAddMessage(style, id, options)
	if err != nil {
		return err
	}

	if err := sam.transmitSessionMessage(message); err != nil {
		return err
	}

	response, err := sam.readSessionResponse()
	if err != nil {
		return err
	}

	return sam.parseSessionAddResponse(response, id)
}

// RemoveSubSession removes a subsession from the primary session using the SESSION REMOVE command.
// This method implements the SAMv3.3 protocol for cleanly terminating subsessions while
// keeping the primary session and other subsessions active.
//
// Parameters:
//   - id: Unique subsession identifier to remove
//
// After removal, the subsession is closed and may not be used for sending or receiving data.
// The primary session and other subsessions remain unaffected by this operation.
//
// Example usage:
//
//	err := sam.RemoveSubSession("stream-sub-1")
func (sam *SAM) RemoveSubSession(id string) error {
	log.WithField("id", id).Debug("Removing subsession from primary session")

	message := []byte("SESSION REMOVE ID=" + id + "\n")
	log.WithField("message", string(message)).Debug("Sending SESSION REMOVE message")

	if err := sam.transmitSessionMessage(message); err != nil {
		return err
	}

	response, err := sam.readSessionResponse()
	if err != nil {
		return err
	}

	return sam.parseSessionRemoveResponse(response, id)
}

// buildSessionAddMessage constructs the SESSION ADD message with style, ID, and options.
func (sam *SAM) buildSessionAddMessage(style, id string, options []string) ([]byte, error) {
	baseMsg := "SESSION ADD STYLE=" + style + " ID=" + id

	// Validate and clean options - SESSION ADD should not contain SIGNATURE_TYPE
	// Per SAMv3.3 spec: "Do not set the DESTINATION option on a SESSION ADD.
	// The subsession will use the destination specified in the primary session."
	// This also applies to SIGNATURE_TYPE since it's part of the destination.
	cleanedOptions := validateSubSessionOptions(options)

	extraStr := strings.Join(cleanedOptions, " ")
	if extraStr != "" {
		baseMsg += " " + extraStr
	}

	message := []byte(baseMsg + "\n")
	log.WithField("message", string(message)).Debug("Built SESSION ADD message")
	return message, nil
}

// parseSessionAddResponse parses the SAM response for SESSION ADD and returns appropriate errors.
func (sam *SAM) parseSessionAddResponse(response, id string) error {
	if strings.HasPrefix(response, SESSION_ADD_OK) {
		log.WithField("id", id).Debug("Successfully added subsession")
		return nil
	}

	log.WithFields(logrus.Fields{
		"id":       id,
		"response": response,
	}).Error("Failed to add subsession")

	return sam.handleErrorResponse(response)
}

// parseSessionRemoveResponse parses the SAM response for SESSION REMOVE and returns appropriate errors.
func (sam *SAM) parseSessionRemoveResponse(response, id string) error {
	if strings.HasPrefix(response, SESSION_REMOVE_OK) {
		log.WithField("id", id).Debug("Successfully removed subsession")
		return nil
	}

	log.WithFields(logrus.Fields{
		"id":       id,
		"response": response,
	}).Error("Failed to remove subsession")

	return sam.handleErrorResponse(response)
}

// NewBaseSessionFromSubsession creates a BaseSession for a subsession that has already been
// registered with a PRIMARY session using SESSION ADD. This constructor is used when the
// subsession is already registered with the SAM bridge and doesn't need a new session creation.
//
// This function is specifically designed for use with SAMv3.3 PRIMARY sessions where
// subsessions are created using SESSION ADD rather than SESSION CREATE commands.
//
// Parameters:
//   - sam: SAM connection for data operations (separate from the primary session's control connection)
//   - id: The subsession ID that was already registered with SESSION ADD
//   - keys: The I2P keys from the primary session (shared across all subsessions)
//
// Returns a BaseSession ready for use without attempting to create a new SAM session.
func NewBaseSessionFromSubsession(sam *SAM, id string, keys i2pkeys.I2PKeys) (*BaseSession, error) {
	log.WithField("id", id).Debug("Creating BaseSession from existing subsession")

	// Create a BaseSession using the provided connection and shared keys
	// The session is already registered with the SAM bridge via SESSION ADD
	baseSession := &BaseSession{
		id:   id,
		conn: sam.Conn,
		keys: keys,
		SAM:  *sam,
	}

	log.WithField("id", id).Debug("Successfully created BaseSession from subsession")
	return baseSession, nil
}
