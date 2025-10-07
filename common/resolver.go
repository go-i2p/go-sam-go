package common

import (
	"bufio"
	"bytes"
	"errors"
	"strings"

	"github.com/go-i2p/i2pkeys"
	"github.com/sirupsen/logrus"
)

// NewSAMResolver creates a new SAMResolver using an existing SAM instance.
// This allows sharing a single SAM connection for both session management and address resolution.
// Returns a configured resolver ready for performing I2P address lookups.
func NewSAMResolver(parent *SAM) (*SAMResolver, error) {
	log.Debug("Creating new SAMResolver from existing SAM instance")
	if parent == nil {
		return nil, errors.New("parent SAM instance cannot be nil")
	}
	var s SAMResolver
	s.SAM = parent
	return &s, nil
}

// NewFullSAMResolver creates a complete SAMResolver with its own SAM connection.
// Establishes a new connection to the specified SAM bridge address for address resolution.
// Returns a fully configured resolver or an error if connection fails.
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

	if err := sam.sendLookupRequest(name, false); err != nil {
		return i2pkeys.I2PAddr(""), err
	}

	response, err := sam.readLookupResponse()
	if err != nil {
		return i2pkeys.I2PAddr(""), err
	}

	scanner, err := sam.prepareLookupScanner(response)
	if err != nil {
		return i2pkeys.I2PAddr(""), err
	}

	addr, _, err := sam.processLookupResponse(scanner, name)
	return addr, err
}

// ResolveWithOptions performs a name lookup with SAMv3.2+ service discovery support.
// When options is true, the SAM bridge may return additional service metadata along with the address.
//
// This function enables SAMv3.2+ service discovery by requesting OPTIONS=true in the NAMING LOOKUP
// command. Service operators can publish metadata about their services to help clients automatically
// discover connection parameters, supported protocols, and service capabilities.
//
// Parameters:
//   - name: I2P hostname to resolve (e.g., "service.i2p", "forum.i2p")
//   - options: Set to true to request service metadata, false for address-only lookup
//
// Returns:
//   - i2pkeys.I2PAddr: The resolved I2P destination address
//   - map[string]string: Service options/metadata (empty if options=false or no metadata available)
//   - error: Any error that occurred during resolution
//
// Service Discovery Examples:
//
//	// Basic address resolution (compatible with older SAM versions)
//	addr, _, err := resolver.ResolveWithOptions("mysite.i2p", false)
//
//	// Service discovery with metadata (SAMv3.2+ feature)
//	addr, metadata, err := resolver.ResolveWithOptions("api.i2p", true)
//	if err != nil {
//	    return err
//	}
//
//	// Extract service information from metadata
//	if port := metadata["port"]; port != "" {
//	    fmt.Printf("Service port: %s\n", port)
//	}
//	if protocol := metadata["protocol"]; protocol != "" {
//	    fmt.Printf("Protocol: %s\n", protocol)
//	}
//	if path := metadata["path"]; path != "" {
//	    fmt.Printf("API path: %s\n", path)
//	}
//
// Common metadata keys include: port, protocol, path, description, version, contact.
// Metadata availability depends on service operator configuration and SAM bridge version.
func (sam *SAMResolver) ResolveWithOptions(name string, options bool) (i2pkeys.I2PAddr, map[string]string, error) {
	log.WithFields(logrus.Fields{
		"name":    name,
		"options": options,
	}).Debug("Resolving name with options")

	if err := sam.sendLookupRequest(name, options); err != nil {
		return i2pkeys.I2PAddr(""), nil, err
	}

	response, err := sam.readLookupResponse()
	if err != nil {
		return i2pkeys.I2PAddr(""), nil, err
	}

	scanner, err := sam.prepareLookupScanner(response)
	if err != nil {
		return i2pkeys.I2PAddr(""), nil, err
	}

	return sam.processLookupResponse(scanner, name)
}

// sendLookupRequest sends a NAMING LOOKUP request to the SAM connection.
// It writes the lookup command with optional OPTIONS=true parameter and handles any connection errors.
func (sam *SAMResolver) sendLookupRequest(name string, options bool) error {
	cmd := "NAMING LOOKUP NAME=" + name
	if options {
		cmd += " OPTIONS=true"
	}
	cmd += "\r\n"

	log.WithFields(logrus.Fields{
		"name":    name,
		"options": options,
		"command": strings.TrimSpace(cmd),
	}).Debug("Sending lookup request")

	if _, err := sam.Conn.Write([]byte(cmd)); err != nil {
		log.WithError(err).Error("Failed to write to SAM connection")
		sam.Close()
		return err
	}
	return nil
}

// readLookupResponse reads the response from the SAM connection.
// Uses dynamic buffer allocation to handle large naming responses with service options.
// It handles reading errors and connection cleanup on failure.
func (sam *SAMResolver) readLookupResponse() ([]byte, error) {
	buf := make([]byte, 4096) // Initial buffer size for typical responses
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read from SAM connection")
		sam.Close()
		return nil, err
	}

	// If buffer was completely filled, there might be more data
	if n == len(buf) {
		// Use a growing buffer to read remaining data
		response := make([]byte, n, len(buf)*2)
		copy(response, buf[:n])

		for {
			additionalBuf := make([]byte, 2048)
			additionalN, err := sam.Conn.Read(additionalBuf)
			if err != nil {
				if additionalN == 0 {
					// Connection closed or no more data
					break
				}
				log.WithError(err).Error("Failed to read additional SAM response data")
				sam.Close()
				return nil, err
			}

			response = append(response, additionalBuf[:additionalN]...)

			// If we didn't fill the additional buffer, we're done
			if additionalN < len(additionalBuf) {
				break
			}
		}

		return response, nil
	}

	return buf[:n], nil
}

// prepareLookupScanner validates the response format and creates a scanner.
// It ensures the response has the correct "NAMING REPLY" prefix and length.
func (sam *SAMResolver) prepareLookupScanner(response []byte) (*bufio.Scanner, error) {
	if len(response) <= 13 || !strings.HasPrefix(string(response), "NAMING REPLY ") {
		log.Error("Failed to parse SAM response")
		return nil, errors.New("failed to parse SAM response")
	}

	scanner := bufio.NewScanner(bytes.NewReader(response[13:]))
	scanner.Split(bufio.ScanWords)
	return scanner, nil
}

// processLookupResponse processes the scanner tokens and returns the resolved address and options.
// It handles different response types and accumulates error messages and service metadata.
func (sam *SAMResolver) processLookupResponse(scanner *bufio.Scanner, name string) (i2pkeys.I2PAddr, map[string]string, error) {
	errStr := ""
	options := make(map[string]string)

	for scanner.Scan() {
		text := scanner.Text()
		log.WithField("text", text).Debug("Parsing SAM response token")

		if resolved, found := sam.handleValueResponse(text); found {
			return resolved, options, nil
		}

		// Handle service options (key=value pairs from OPTIONS=true)
		if sam.handleOptionsResponse(text, options) {
			continue
		}

		if sam.shouldSkipToken(text, name) {
			continue
		}

		errStr = sam.handleErrorResponse(text, name, errStr)
	}
	return i2pkeys.I2PAddr(""), options, errors.New(errStr)
}

// handleValueResponse processes VALUE= responses and returns the resolved address.
// Returns the address and true if this was a VALUE response, empty address and false otherwise.
func (sam *SAMResolver) handleValueResponse(text string) (i2pkeys.I2PAddr, bool) {
	if strings.HasPrefix(text, "VALUE=") {
		addr := i2pkeys.I2PAddr(text[6:])
		log.WithField("addr", addr).Debug("Name resolved successfully")
		return addr, true
	}
	return i2pkeys.I2PAddr(""), false
}

// handleOptionsResponse processes service option responses from SAMv3.2+ OPTIONS=true lookups.
// Returns true if this was an option response that was handled, false otherwise.
// Options are in format "key=value" and get stored in the options map.
//
// SAMv3.2+ Service Discovery Examples:
//
// When performing a lookup with OPTIONS=true, you can receive service metadata:
//
//	resolver.LookupWithOptions("myservice.i2p", true)
//
// Common service option keys include:
//   - "port": TCP port number for the service (e.g., "port=8080")
//   - "protocol": Service protocol type (e.g., "protocol=http", "protocol=https")
//   - "path": Default path for HTTP services (e.g., "path=/api/v1")
//   - "description": Human-readable service description
//   - "version": Service version information
//   - "contact": Service operator contact information
//
// Usage Pattern:
//
//	addr, options, err := resolver.LookupWithOptions("service.i2p", true)
//	if err != nil {
//	    log.Fatal("Lookup failed:", err)
//	}
//
//	// Check for HTTP service with specific port
//	if port := options["port"]; port != "" {
//	    if protocol := options["protocol"]; protocol == "http" {
//	        fmt.Printf("HTTP service available at %s:%s\n", addr, port)
//	    }
//	}
//
//	// Check for API path
//	if path := options["path"]; path != "" {
//	    fmt.Printf("API endpoint: %s%s\n", addr, path)
//	}
//
// Service operators can publish this metadata in their I2P router configuration
// to help clients discover service capabilities and connection parameters.
func (sam *SAMResolver) handleOptionsResponse(text string, options map[string]string) bool {
	// Skip standard SAM response tokens
	if strings.HasPrefix(text, "RESULT=") || strings.HasPrefix(text, "NAME=") ||
		strings.HasPrefix(text, "VALUE=") || strings.HasPrefix(text, "MESSAGE=") {
		return false
	}

	// Check if this looks like a service option (contains = but not a known SAM token)
	if strings.Contains(text, "=") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				options[key] = value
				log.WithFields(logrus.Fields{
					"key":   key,
					"value": value,
				}).Debug("Added service option")
				return true
			}
		}
	}

	return false
}

// shouldSkipToken determines if a token should be skipped without processing.
// Returns true for OK results and NAME= tokens that match the requested name.
func (sam *SAMResolver) shouldSkipToken(text, name string) bool {
	return text == SAM_RESULT_OK || text == "NAME="+name
}

// handleErrorResponse processes error responses and accumulates error messages.
// Returns the updated error string with any new error information.
func (sam *SAMResolver) handleErrorResponse(text, name, errStr string) string {
	if text == SAM_RESULT_INVALID_KEY {
		errStr += "Invalid key - resolver."
		log.Error("Invalid key in resolver")
	} else if text == SAM_RESULT_KEY_NOT_FOUND {
		errStr += "Unable to resolve " + name
		log.WithField("name", name).Error("Unable to resolve name")
	} else if strings.HasPrefix(text, "MESSAGE=") {
		errStr += " " + text[8:]
		log.WithField("message", text[8:]).Warn("Received message from SAM")
	}
	return errStr
}
