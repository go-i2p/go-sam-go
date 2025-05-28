package common

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultSAMHost = "127.0.0.1"
	defaultSAMPort = 7656
)

// Sam returns the SAM bridge address as a string in the format "host:port"
func (f *I2PConfig) Sam() string {
	host := f.SamHost
	if host == "" {
		host = defaultSAMHost
	}

	port := f.SamPort
	if port == 0 {
		port = defaultSAMPort
	}

	return fmt.Sprintf("%s:%d", host, port)
}

// SAMAddress returns the SAM bridge address in the format "host:port"
// This is a convenience method that uses the Sam() function to get the address.
// It is used to provide a consistent interface for retrieving the SAM address.
func (f *I2PConfig) SAMAddress() string {
	// Return the SAM address in the format "host:port"
	return f.Sam()
}

// SetSAMAddress sets the SAM bridge host and port from a combined address string.
// If no address is provided, it sets default values for the host and port.
func (f *I2PConfig) SetSAMAddress(addr string) {
	if addr == "" {
		f.SamHost = defaultSAMHost
		f.SamPort = defaultSAMPort
		return
	}

	host, portStr, err := SplitHostPort(addr)
	if err != nil {
		log.WithError(err).Warn("Failed to parse SAM address, using defaults")
		f.SamHost = defaultSAMHost
		f.SamPort = defaultSAMPort
		return
	}

	f.SamHost = host

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		log.WithField("port", portStr).Warn("Invalid port, setting to 7656")
		f.SamPort = defaultSAMPort
	} else {
		f.SamPort = port
	}
}

// ID returns the tunnel name as a formatted string. If no tunnel name is set,
// generates a random 12-character name using lowercase letters.
func (f *I2PConfig) ID() string {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	// If no tunnel name set, generate random one
	if f.TunName == "" {
		// Generate 12 random lowercase letters
		b := make([]byte, 12)
		for i := range b {
			b[i] = "abcdefghijklmnopqrstuvwxyz"[generator.Intn(26)]
		}
		f.TunName = string(b)

		// Log the generated name
		log.WithField("TunName", f.TunName).Debug("Generated random tunnel name")
	}

	// Return formatted ID string
	return fmt.Sprintf("ID=%s", f.TunName)
}

// Leasesetsettings returns the lease set configuration strings for I2P
// Returns three strings: lease set key, private key, and private signing key settings
func (f *I2PConfig) LeaseSetSettings() (string, string, string) {
	// Initialize empty strings for each setting
	var leaseSetKey, privateKey, privateSigningKey string

	// Set lease set key if configured
	if f.LeaseSetKey != "" {
		leaseSetKey = fmt.Sprintf(" i2cp.leaseSetKey=%s ", f.LeaseSetKey)
	}

	// Set lease set private key if configured
	if f.LeaseSetPrivateKey != "" {
		privateKey = fmt.Sprintf(" i2cp.leaseSetPrivateKey=%s ", f.LeaseSetPrivateKey)
	}

	// Set lease set private signing key if configured
	if f.LeaseSetPrivateSigningKey != "" {
		privateSigningKey = fmt.Sprintf(" i2cp.leaseSetPrivateSigningKey=%s ", f.LeaseSetPrivateSigningKey)
	}

	// Log the constructed settings
	log.WithFields(logrus.Fields{
		"leaseSetKey":               leaseSetKey,
		"leaseSetPrivateKey":        privateKey,
		"leaseSetPrivateSigningKey": privateSigningKey,
	}).Debug("Lease set settings constructed")

	return leaseSetKey, privateKey, privateSigningKey
}

// FromPort returns the FROM_PORT configuration string for SAM bridges >= 3.1
// Returns an empty string if SAM version < 3.1 or if fromport is "0"
func (f *I2PConfig) FromPort() string {
	// Check SAM version compatibility
	if f.SamMax == "" || f.samMax() < 3.1 {
		log.Debug("SAM version < 3.1, FromPort not applicable")
		return ""
	}

	// Return formatted FROM_PORT if fromport is set
	if f.Fromport != "0" {
		log.WithField("fromPort", f.Fromport).Debug("FromPort set")
		return fmt.Sprintf(" FROM_PORT=%s ", f.Fromport)
	}

	log.Debug("FromPort not set")
	return ""
}

// ToPort returns the TO_PORT configuration string for SAM bridges >= 3.1
// Returns an empty string if SAM version < 3.1 or if toport is "0"
func (f *I2PConfig) ToPort() string {
	// Check SAM version compatibility
	if f.samMax() < 3.1 {
		log.Debug("SAM version < 3.1, ToPort not applicable")
		return ""
	}

	// Return formatted TO_PORT if toport is set
	if f.Toport != "0" {
		log.WithField("toPort", f.Toport).Debug("ToPort set")
		return fmt.Sprintf(" TO_PORT=%s ", f.Toport)
	}

	log.Debug("ToPort not set")
	return ""
}

// SessionStyle returns the SAM session style configuration string
// If no style is set, defaults to "STREAM"
func (f *I2PConfig) SessionStyle() string {
	if f.Style != "" {
		// Log custom style setting
		log.WithField("style", f.Style).Debug("Session style set")
		return fmt.Sprintf(" STYLE=%s ", f.Style)
	}

	// Log default style
	log.Debug("Using default STREAM style")
	return " STYLE=STREAM "
}

// samMax returns the maximum SAM version supported as a float64
// If parsing fails, returns default value 3.1
func (f *I2PConfig) samMax() float64 {
	// Parse SAM max version to integer
	i, err := strconv.ParseFloat(f.SamMax, 64)
	if err != nil {
		log.WithError(err).Warn("Failed to parse SamMax, using default 3.1")
		return 3.1
	}

	// Log the parsed version and return
	log.WithField("samMax", i).Debug("SAM max version parsed")
	return i
}

// MinSAM returns the minimum SAM version supported as a string
// If no minimum version is set, returns default value "3.0"
func (f *I2PConfig) MinSAM() string {
	if f.SamMin == "" {
		log.Debug("Using default MinSAM: 3.0")
		return "3.0"
	}
	log.WithField("minSAM", f.SamMin).Debug("MinSAM set")
	return f.SamMin
}

// MaxSAM returns the maximum SAM version supported as a string
// If no maximum version is set, returns default value "3.1"
func (f *I2PConfig) MaxSAM() string {
	if f.SamMax == "" {
		log.Debug("Using default MaxSAM: 3.1")
		return "3.1"
	}
	log.WithField("maxSAM", f.SamMax).Debug("MaxSAM set")
	return f.SamMax
}

// DestinationKey returns the DESTINATION configuration string for the SAM bridge
// If destination keys are set, returns them as a string, otherwise returns "TRANSIENT"
func (f *I2PConfig) DestinationKey() string {
	// Check if destination keys are set
	if f.DestinationKeys != nil {
		// Log the destination key being used
		log.WithField("destinationKey", f.DestinationKeys.String()).Debug("Destination key set")
		return fmt.Sprintf(" DESTINATION=%s ", f.DestinationKeys.String())
	}

	// Log and return transient destination
	log.Debug("Using TRANSIENT destination")
	return " DESTINATION=TRANSIENT "
}

// SignatureType returns the SIGNATURE_TYPE configuration string for SAM bridges >= 3.1
// Returns empty string if SAM version < 3.1 or if no signature type is set
func (f *I2PConfig) SignatureType() string {
	// Check SAM version compatibility
	if f.samMax() < 3.1 {
		log.Debug("SAM version < 3.1, SignatureType not applicable")
		return ""
	}

	// Return formatted signature type if set
	if f.SigType != "" {
		log.WithField("sigType", f.SigType).Debug("Signature type set")
		return fmt.Sprintf(" SIGNATURE_TYPE=%s ", f.SigType)
	}

	log.Debug("Signature type not set")
	return ""
}

// EncryptLease returns the lease set encryption configuration string
// Returns "i2cp.encryptLeaseSet=true" if encryption is enabled, empty string otherwise
func (f *I2PConfig) EncryptLease() string {
	if f.EncryptLeaseSet {
		log.Debug("Lease set encryption enabled")
		return " i2cp.encryptLeaseSet=true "
	}
	log.Debug("Lease set encryption not enabled")
	return ""
}

// Reliability returns the message reliability configuration string for the SAM bridge
// If a reliability setting is specified, returns formatted i2cp.messageReliability setting
func (f *I2PConfig) Reliability() string {
	if f.MessageReliability != "" {
		// Log the reliability setting being used
		log.WithField("reliability", f.MessageReliability).Debug("Message reliability set")
		return fmt.Sprintf(" i2cp.messageReliability=%s ", f.MessageReliability)
	}

	// Log when reliability is not set
	log.Debug("Message reliability not set")
	return ""
}

// Reduce returns I2CP reduce-on-idle configuration settings as a string if enabled
func (f *I2PConfig) Reduce() string {
	// Return early if reduce idle is not enabled
	if !f.ReduceIdle {
		log.Debug("Reduce idle settings not applied")
		return ""
	}

	// Log and return the reduce idle configuration
	result := fmt.Sprintf("i2cp.reduceOnIdle=%t i2cp.reduceIdleTime=%d i2cp.reduceQuantity=%d",
		f.ReduceIdle, f.ReduceIdleTime, f.ReduceIdleQuantity)
	log.WithField("config", result).Debug("Reduce idle settings applied")
	return result
}

// Close returns I2CP close-on-idle configuration settings as a string if enabled
func (f *I2PConfig) Close() string {
	// Return early if close idle is not enabled
	if !f.CloseIdle {
		log.Debug("Close idle settings not applied")
		return ""
	}
	// Log and return the close idle configuration
	result := fmt.Sprintf("i2cp.closeOnIdle=%t i2cp.closeIdleTime=%d",
		f.CloseIdle, f.CloseIdleTime)
	log.WithField("config", result).Debug("Close idle settings applied")
	return result
}

// DoZero returns the zero hop and fast receive configuration string settings
func (f *I2PConfig) DoZero() string {
	// Build settings using slices for cleaner concatenation
	var settings []string
	// Add inbound zero hop setting if enabled
	if f.InAllowZeroHop {
		settings = append(settings, fmt.Sprintf("inbound.allowZeroHop=%t", f.InAllowZeroHop))
	}
	// Add outbound zero hop setting if enabled
	if f.OutAllowZeroHop {
		settings = append(settings, fmt.Sprintf("outbound.allowZeroHop=%t", f.OutAllowZeroHop))
	}
	// Add fast receive setting if enabled
	if f.FastRecieve {
		settings = append(settings, fmt.Sprintf("i2cp.fastRecieve=%t", f.FastRecieve))
	}
	// Join all settings with spaces
	result := strings.Join(settings, " ")
	// Log the final settings
	log.WithField("zeroHopSettings", result).Debug("Zero hop settings applied")

	return result
}

// formatConfigPair creates a configuration string for inbound/outbound pairs
func (f *I2PConfig) formatConfigPair(direction, property string, value interface{}) string {
	switch v := value.(type) {
	case int:
		return fmt.Sprintf("%s.%s=%d", direction, property, v)
	case string:
		return fmt.Sprintf("%s.%s=%s", direction, property, v)
	case bool:
		return fmt.Sprintf("%s.%s=%t", direction, property, v)
	default:
		return ""
	}
}

func (f *I2PConfig) InboundLength() string {
	return f.formatConfigPair("inbound", "length", f.InLength)
}

func (f *I2PConfig) OutboundLength() string {
	return f.formatConfigPair("outbound", "length", f.OutLength)
}

func (f *I2PConfig) InboundLengthVariance() string {
	return f.formatConfigPair("inbound", "lengthVariance", f.InVariance)
}

func (f *I2PConfig) OutboundLengthVariance() string {
	return f.formatConfigPair("outbound", "lengthVariance", f.OutVariance)
}

func (f *I2PConfig) InboundBackupQuantity() string {
	return f.formatConfigPair("inbound", "backupQuantity", f.InBackupQuantity)
}

func (f *I2PConfig) OutboundBackupQuantity() string {
	return f.formatConfigPair("outbound", "backupQuantity", f.OutBackupQuantity)
}

func (f *I2PConfig) InboundQuantity() string {
	return f.formatConfigPair("inbound", "quantity", f.InQuantity)
}

func (f *I2PConfig) OutboundQuantity() string {
	return f.formatConfigPair("outbound", "quantity", f.OutQuantity)
}

func (f *I2PConfig) UsingCompression() string {
	return f.formatConfigPair("i2cp", "useCompression", f.UseCompression)
}

// Print returns a slice of strings containing all the I2P configuration settings
func (f *I2PConfig) Print() []string {
	var settings []string
	// Collect tunnel configuration settings
	settings = append(settings, f.collectTunnelSettings()...)
	// Collect connection behavior settings
	settings = append(settings, f.collectConnectionSettings()...)
	// Collect lease set settings
	settings = append(settings, f.collectLeaseSetSettings()...)
	// Collect access control settings
	settings = append(settings, f.collectAccessSettings()...)
	return settings
}

// Accesslisttype returns the I2CP access list configuration string based on the AccessListType setting
func (f *I2PConfig) Accesslisttype() string {
	switch f.AccessListType {
	case ACCESS_TYPE_WHITELIST:
		log.Debug("Access list type set to allowlist")
		return "i2cp.enableAccessList=true"
	case ACCESS_TYPE_BLACKLIST:
		log.Debug("Access list type set to blocklist")
		return "i2cp.enableBlackList=true"
	default:
		log.Debug("Access list type not set")
		return ""
	}
}

// Accesslist generates the I2CP access list configuration string based on the configured access list
func (f *I2PConfig) Accesslist() string {
	// Only proceed if access list type and values are set
	if f.AccessListType != "" && len(f.AccessList) > 0 {
		// Join access list entries with commas
		accessList := strings.Join(f.AccessList, ",")
		// Log the generated access list
		log.WithField("accessList", accessList).Debug("Access list generated")
		// Return formatted access list configuration
		return fmt.Sprintf("i2cp.accessList=%s", accessList)
	}

	// Log when access list is not set
	log.Debug("Access list not set")
	return ""
}

// LeaseSetEncryptionType returns the I2CP lease set encryption type configuration string.
// If no encryption type is set, returns default value "4,0".
// Validates that all encryption types are valid integers.
func (f *I2PConfig) LeaseSetEncryptionType() string {
	// Use default encryption type if none specified
	if f.LeaseSetEncryption == "" {
		log.Debug("Using default lease set encryption type: 4,0")
		return "i2cp.leaseSetEncType=4,0"
	}

	// Validate each encryption type is a valid integer
	for _, s := range strings.Split(f.LeaseSetEncryption, ",") {
		if _, err := strconv.Atoi(s); err != nil {
			log.WithField("invalidType", s).Panic("Invalid encrypted leaseSet type")
			// panic("Invalid encrypted leaseSet type: " + s)
		}
	}

	// Log and return the configured encryption type
	log.WithField("leaseSetEncType", f.LeaseSetEncryption).Debug("Lease set encryption type set")
	return fmt.Sprintf("i2cp.leaseSetEncType=%s", f.LeaseSetEncryption)
}

// collectTunnelSettings returns all tunnel-related configuration strings
func (f *I2PConfig) collectTunnelSettings() []string {
	return []string{
		f.InboundLength(),
		f.OutboundLength(),
		f.InboundLengthVariance(),
		f.OutboundLengthVariance(),
		f.InboundBackupQuantity(),
		f.OutboundBackupQuantity(),
		f.InboundQuantity(),
		f.OutboundQuantity(),
	}
}

// collectConnectionSettings returns all connection behavior configuration strings
func (f *I2PConfig) collectConnectionSettings() []string {
	return []string{
		f.UsingCompression(),
		f.DoZero(),      // Zero hop settings
		f.Reduce(),      // Reduce idle settings
		f.Close(),       // Close idle settings
		f.Reliability(), // Message reliability
	}
}

// collectLeaseSetSettings returns all lease set configuration strings
func (f *I2PConfig) collectLeaseSetSettings() []string {
	lsk, lspk, lspsk := f.LeaseSetSettings()
	return []string{
		f.EncryptLease(), // Lease encryption
		lsk, lspk, lspsk, // Lease set keys
		f.LeaseSetEncryptionType(), // Lease set encryption type
	}
}

// collectAccessSettings returns all access control configuration strings
func (f *I2PConfig) collectAccessSettings() []string {
	return []string{
		f.Accesslisttype(), // Access list type
		f.Accesslist(),     // Access list
	}
}

func NewConfig(opts ...func(*I2PConfig) error) (*I2PConfig, error) {
	// Initialize with struct literal containing only non-zero defaults
	// Go automatically zero-initializes all other fields
	config := I2PConfig{
		SamHost:            "127.0.0.1",
		SamPort:            7656,
		SamMin:             DEFAULT_SAM_MIN,
		SamMax:             DEFAULT_SAM_MAX,
		TunType:            "server",
		Style:              SESSION_STYLE_STREAM,
		InLength:           3,
		OutLength:          3,
		InQuantity:         2,
		OutQuantity:        2,
		InVariance:         1,
		OutVariance:        1,
		InBackupQuantity:   3,
		OutBackupQuantity:  3,
		UseCompression:     true,
		ReduceIdleTime:     15,
		ReduceIdleQuantity: 4,
		CloseIdleTime:      300000,
		MessageReliability: "none",
	}

	// Apply functional options
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}
	return &config, nil
}
