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

// Keys retrieves the I2P destination keys associated with this SAM instance.
// Returns a pointer to the keys used for this SAM session's I2P identity.
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

// EnsureKeyfile ensures cryptographic keys are available, either by generating transient keys
// or by loading/creating persistent keys from the specified file.
func (sam *SAM) EnsureKeyfile(fname string) (keys i2pkeys.I2PKeys, err error) {
	if fname == "" {
		keys, err = sam.generateTransientKeys()
	} else {
		keys, err = sam.ensurePersistentKeys(fname)
	}

	if err != nil {
		log.WithError(err).Error("Failed to ensure keyfile")
	}
	return
}

// generateTransientKeys creates new temporary keys that are not saved to disk.
func (sam *SAM) generateTransientKeys() (i2pkeys.I2PKeys, error) {
	keys, err := sam.NewKeys()
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}

	sam.SAMEmit.I2PConfig.DestinationKeys = &keys
	log.WithFields(logrus.Fields{
		"keys": keys,
	}).Debug("Generated new transient keys")

	return keys, nil
}

// ensurePersistentKeys loads existing keys from file or creates new ones if file doesn't exist.
func (sam *SAM) ensurePersistentKeys(fname string) (i2pkeys.I2PKeys, error) {
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		return sam.createAndSaveKeys(fname)
	} else if err == nil {
		return sam.loadKeysFromFile(fname)
	}
	return i2pkeys.I2PKeys{}, err
}

// createAndSaveKeys generates new keys and saves them to the specified file.
func (sam *SAM) createAndSaveKeys(fname string) (i2pkeys.I2PKeys, error) {
	keys, err := sam.NewKeys()
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}

	sam.SAMEmit.I2PConfig.DestinationKeys = &keys

	if err := sam.saveKeysToFile(keys, fname); err != nil {
		return i2pkeys.I2PKeys{}, err
	}

	log.Debug("Generated and saved new keys")
	return keys, nil
}

// loadKeysFromFile loads cryptographic keys from the specified file.
func (sam *SAM) loadKeysFromFile(fname string) (i2pkeys.I2PKeys, error) {
	f, err := os.Open(fname)
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}
	defer f.Close()

	keys, err := i2pkeys.LoadKeysIncompat(f)
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}

	sam.SAMEmit.I2PConfig.DestinationKeys = &keys
	log.Debug("Loaded existing keys from file")
	return keys, nil
}

// saveKeysToFile saves cryptographic keys to the specified file with appropriate permissions.
func (sam *SAM) saveKeysToFile(keys i2pkeys.I2PKeys, fname string) error {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	return i2pkeys.StoreKeysIncompat(keys, f)
}

// Creates the I2P-equivalent of an IP address, that is unique and only the one
// who has the private keys can send messages from. The public keys are the I2P
// desination (the address) that anyone can send messages to.
func (sam *SAM) NewKeys(sigType ...string) (i2pkeys.I2PKeys, error) {
	log.WithField("sigType", sigType).Debug("Generating new keys")

	sigTypeStr := sam.prepareSigType(sigType)

	if err := sam.sendDestGenerateCommand(sigTypeStr); err != nil {
		return i2pkeys.I2PKeys{}, err
	}

	response, err := sam.readKeyGenerationResponse()
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}

	pub, priv, err := sam.parseKeyResponse(response)
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}

	log.Debug("Successfully generated new keys")
	return i2pkeys.NewKeys(i2pkeys.I2PAddr(pub), priv), nil
}

// prepareSigType extracts the signature type from the input parameters.
// It returns the first signature type if provided, otherwise returns an empty string.
func (sam *SAM) prepareSigType(sigType []string) string {
	if len(sigType) > 0 {
		return sigType[0]
	}
	return ""
}

// sendDestGenerateCommand sends the DEST GENERATE command to the SAM connection.
// It constructs and transmits the command with the specified signature type.
func (sam *SAM) sendDestGenerateCommand(sigType string) error {
	command := "DEST GENERATE " + sigType + "\n"
	if _, err := sam.Conn.Write([]byte(command)); err != nil {
		log.WithError(err).Error("Failed to write DEST GENERATE command")
		return oops.Errorf("error with writing in SAM: %w", err)
	}
	return nil
}

// readKeyGenerationResponse reads the response from the SAM connection.
// It allocates a buffer and reads the response data for key generation.
func (sam *SAM) readKeyGenerationResponse() ([]byte, error) {
	buf := make([]byte, 8192)
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response for key generation")
		return nil, oops.Errorf("error with reading in SAM: %w", err)
	}
	return buf[:n], nil
}

// parseKeyResponse parses the SAM response to extract public and private keys.
// It scans the response tokens and extracts the PUB and PRIV key values.
func (sam *SAM) parseKeyResponse(response []byte) (string, string, error) {
	s := bufio.NewScanner(bytes.NewReader(response))
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
			return "", "", oops.Errorf("Failed to parse keys.")
		}
	}
	return pub, priv, nil
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
