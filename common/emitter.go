package common

import (
	"fmt"
	"strings"
)

// SamOptionsString generates a space-separated string of all I2P configuration options.
// Used internally to construct SAM protocol messages with tunnel and session parameters.
func (e *SAMEmit) SamOptionsString() string {
	optStr := strings.Join(e.I2PConfig.Print(), " ")
	log.WithField("optStr", optStr).Debug("Generated option string")
	return optStr
}

// Hello generates the SAM protocol HELLO command for initial handshake.
// Includes minimum and maximum supported SAM protocol versions for negotiation.
func (e *SAMEmit) Hello() string {
	hello := fmt.Sprintf("HELLO VERSION MIN=%s MAX=%s \n", e.I2PConfig.MinSAM(), e.I2PConfig.MaxSAM())
	log.WithField("hello", hello).Debug("Generated HELLO command")
	return hello
}

// HelloBytes returns the HELLO command as a byte slice for network transmission.
// Convenience method for sending the handshake command over network connections.
func (e *SAMEmit) HelloBytes() []byte {
	return []byte(e.Hello())
}

// GenerateDestination creates a SAM DEST GENERATE command for key generation.
// Uses the configured signature type to request new I2P destination keys from the router.
func (e *SAMEmit) GenerateDestination() string {
	dest := fmt.Sprintf("DEST GENERATE %s \n", e.I2PConfig.SignatureType())
	log.WithField("destination", dest).Debug("Generated DEST GENERATE command")
	return dest
}

// GenerateDestinationBytes returns the DEST GENERATE command as bytes.
// Convenience method for network transmission of key generation requests.
func (e *SAMEmit) GenerateDestinationBytes() []byte {
	return []byte(e.GenerateDestination())
}

// Lookup generates a SAM NAMING LOOKUP command for address resolution.
// Takes a human-readable name and creates a command to resolve it to an I2P destination.
func (e *SAMEmit) Lookup(name string) string {
	lookup := fmt.Sprintf("NAMING LOOKUP NAME=%s \n", name)
	log.WithField("lookup", lookup).Debug("Generated NAMING LOOKUP command")
	return lookup
}

// LookupBytes returns the NAMING LOOKUP command as bytes for transmission.
// Convenience method for sending address resolution requests over network connections.
func (e *SAMEmit) LookupBytes(name string) []byte {
	return []byte(e.Lookup(name))
}

// Create generates a SAM SESSION CREATE command for establishing new sessions.
// Combines session style, ports, ID, destination, signature type, and options into a single command.
func (e *SAMEmit) Create() string {
	create := fmt.Sprintf(
		//             //1 2 3 4 5 6 7
		"SESSION CREATE %s%s%s%s%s%s%s \n",
		e.I2PConfig.SessionStyle(),   // 1
		e.I2PConfig.FromPort(),       // 2
		e.I2PConfig.ToPort(),         // 3
		e.I2PConfig.ID(),             // 4
		e.I2PConfig.DestinationKey(), // 5
		e.I2PConfig.SignatureType(),  // 6
		e.SamOptionsString(),         // 7
	)
	log.WithField("create", create).Debug("Generated SESSION CREATE command")
	return create
}

// CreateBytes returns the SESSION CREATE command as bytes for network transmission.
// Includes debug output of the command for troubleshooting session creation issues.
func (e *SAMEmit) CreateBytes() []byte {
	fmt.Println("sam command: " + e.Create())
	return []byte(e.Create())
}

// Connect generates a SAM STREAM CONNECT command for establishing connections.
// Takes a destination address and creates a command to connect to that I2P destination.
func (e *SAMEmit) Connect(dest string) string {
	connect := fmt.Sprintf(
		"STREAM CONNECT ID=%s %s %s DESTINATION=%s \n",
		e.I2PConfig.ID(),
		e.I2PConfig.FromPort(),
		e.I2PConfig.ToPort(),
		dest,
	)
	log.WithField("connect", connect).Debug("Generated STREAM CONNECT command")
	return connect
}

// ConnectBytes returns the STREAM CONNECT command as bytes for transmission.
// Convenience method for sending connection requests over network connections.
func (e *SAMEmit) ConnectBytes(dest string) []byte {
	return []byte(e.Connect(dest))
}

// Accept generates a SAM STREAM ACCEPT command for accepting incoming connections.
// Creates a command to listen for and accept connections on the configured session.
func (e *SAMEmit) Accept() string {
	accept := fmt.Sprintf(
		"STREAM ACCEPT ID=%s %s %s",
		e.I2PConfig.ID(),
		e.I2PConfig.FromPort(),
		e.I2PConfig.ToPort(),
	)
	log.WithField("accept", accept).Debug("Generated STREAM ACCEPT command")
	return accept
}

// AcceptBytes returns the STREAM ACCEPT command as bytes for transmission.
// Convenience method for sending accept requests over network connections.
func (e *SAMEmit) AcceptBytes() []byte {
	return []byte(e.Accept())
}

// NewEmit creates a new SAMEmit instance with the specified configuration options.
// Applies functional options to configure the emitter with custom settings.
// Returns an error if any option fails to apply correctly.
func NewEmit(opts ...func(*SAMEmit) error) (*SAMEmit, error) {
	var emit SAMEmit
	for _, o := range opts {
		if err := o(&emit); err != nil {
			log.WithError(err).Error("Failed to apply option")
			return nil, err
		}
	}
	log.Debug("New SAMEmit instance created")
	return &emit, nil
}
