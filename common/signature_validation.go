package common

import (
	"strings"

	"github.com/samber/oops"
	"github.com/go-i2p/logger"
)

// validateAndCleanOptions validates signature type options and resolves conflicts.
// If sigType parameter is specified and SIGNATURE_TYPE is also present in options,
// the sigType parameter takes precedence and conflicting options are removed.
// Returns cleaned options slice and logs warnings for any conflicts detected.
func validateAndCleanOptions(sigType string, options []string) []string {
	// If no sigType specified, return options as-is
	if sigType == "" {
		return options
	}

	var cleanedOptions []string
	var conflictDetected bool
	var conflictingEntries []string

	// Process each option, removing any SIGNATURE_TYPE entries
	for _, opt := range options {
		if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
			conflictDetected = true
			conflictingEntries = append(conflictingEntries, opt)
			logger.WithFields(logger.Fields{
				"sigType":           sigType,
				"conflictingOption": opt,
			}).Warn("Signature type conflict detected: sigType parameter takes precedence")
			// Skip this option - sigType parameter takes precedence
			continue
		}
		cleanedOptions = append(cleanedOptions, opt)
	}

	if conflictDetected {
		logger.WithFields(logger.Fields{
			"resolvedSigType":  sigType,
			"removedOptions":   conflictingEntries,
			"remainingOptions": cleanedOptions,
		}).Warn("Signature type conflicts resolved by using sigType parameter")
	}

	return cleanedOptions
}

// ValidateSignatureTypeOptions checks for duplicate SIGNATURE_TYPE entries in options.
// Returns an error if multiple SIGNATURE_TYPE entries are found, as this creates ambiguity.
func ValidateSignatureTypeOptions(options []string) error {
	var signatureTypes []string

	for _, opt := range options {
		if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
			signatureTypes = append(signatureTypes, opt)
		}
	}

	if len(signatureTypes) > 1 {
		logger.WithField("duplicateSignatureTypes", signatureTypes).Error("Multiple SIGNATURE_TYPE entries found in options")
		return oops.Errorf("multiple SIGNATURE_TYPE entries in options: %v", signatureTypes)
	}

	return nil
}

// ExtractSignatureType extracts the signature type from options and returns
// the signature type value and the remaining options without SIGNATURE_TYPE entries.
// Returns empty string if no SIGNATURE_TYPE is found in options.
func ExtractSignatureType(options []string) (string, []string) {
	var sigType string
	var remainingOptions []string

	for _, opt := range options {
		if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
			// Extract the signature type value (remove "SIGNATURE_TYPE=" prefix)
			sigType = opt[len("SIGNATURE_TYPE="):]
		} else {
			remainingOptions = append(remainingOptions, opt)
		}
	}

	return sigType, remainingOptions
}

// EnsureSignatureType ensures that only one signature type is specified.
// If sigType parameter is provided, it takes precedence and any SIGNATURE_TYPE
// entries in options are removed. If sigType is empty, options are returned as-is.
// This function provides a safe way to merge signature type specifications.
func EnsureSignatureType(sigType string, options []string) []string {
	return validateAndCleanOptions(sigType, options)
}

// validateSubSessionOptions validates and cleans options for SESSION ADD commands.
// Per SAMv3.3 spec: "Do not set the DESTINATION option on a SESSION ADD.
// The subsession will use the destination specified in the primary session."
// Also removes other options that are invalid for subsessions according to SAMv3.3:
// - SIGNATURE_TYPE (part of destination, inherited from primary)
// - i2cp.leaseSetEncType (lease set config, belongs to primary session)
// - Tunnel configuration options (inbound/outbound lengths, quantities, etc.)
// - Primary session specific options (sam.udp.host, sam.udp.port for primary)
func validateSubSessionOptions(options []string) []string {
	var cleanedOptions []string
	var removedEntries []string

	// Define options that are invalid for subsessions
	invalidPrefixes := []string{
		"SIGNATURE_TYPE=",
		"DESTINATION=",
		"i2cp.leaseSetEncType=",
		"inbound.length=",
		"outbound.length=",
		"inbound.quantity=",
		"outbound.quantity=",
		"inbound.lengthVariance=",
		"outbound.lengthVariance=",
		"inbound.backupQuantity=",
		"outbound.backupQuantity=",
	}

	for _, opt := range options {
		shouldRemove := false
		var reason string

		// Check against known invalid prefixes
		for _, prefix := range invalidPrefixes {
			if strings.HasPrefix(opt, prefix) {
				shouldRemove = true
				switch {
				case strings.HasPrefix(prefix, "SIGNATURE_TYPE="):
					reason = "subsessions inherit signature type from primary session"
				case strings.HasPrefix(prefix, "DESTINATION="):
					reason = "subsessions use destination from primary session"
				case strings.HasPrefix(prefix, "i2cp.leaseSetEncType="):
					reason = "lease set encryption is configured at primary session level"
				case strings.HasPrefix(prefix, "inbound.") || strings.HasPrefix(prefix, "outbound."):
					reason = "tunnel configuration belongs to primary session"
				default:
					reason = "option not valid for subsessions"
				}
				break
			}
		}

		// Additional check for duplicate i2cp.leaseSetEncType entries
		if strings.HasPrefix(opt, "i2cp.leaseSetEncType=") {
			// Check if we already have this option type in cleaned options
			for _, existing := range cleanedOptions {
				if strings.HasPrefix(existing, "i2cp.leaseSetEncType=") {
					logger.WithFields(logger.Fields{
						"existing":  existing,
						"duplicate": opt,
					}).Warn("Duplicate i2cp.leaseSetEncType detected in subsession options")
					shouldRemove = true
					reason = "duplicate lease set encryption type"
					break
				}
			}
		}

		if shouldRemove {
			removedEntries = append(removedEntries, opt)
			logger.WithFields(logger.Fields{
				"removedOption": opt,
				"reason":        reason,
			}).Warn("Removing invalid option from SESSION ADD per SAMv3.3 spec")
		} else {
			cleanedOptions = append(cleanedOptions, opt)
		}
	}

	if len(removedEntries) > 0 {
		logger.WithFields(logger.Fields{
			"removedOptions":   removedEntries,
			"remainingOptions": cleanedOptions,
		}).Warn("SESSION ADD invalid options removed - subsessions inherit configuration from primary session per SAMv3.3 spec")
	}

	return cleanedOptions
}

// validatePrimarySessionOptions validates options for SESSION CREATE (primary sessions).
// Detects and handles duplicate options, especially i2cp.leaseSetEncType which can
// cause ambiguity in encryption type selection.
// Returns cleaned options with duplicates resolved and logs warnings for conflicts.
func validatePrimarySessionOptions(options []string) []string {
	var cleanedOptions []string
	optionCounts := make(map[string][]string)

	// Group options by their key (prefix before '=')
	for _, opt := range options {
		parts := strings.SplitN(opt, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			optionCounts[key] = append(optionCounts[key], opt)
		} else {
			// Options without '=' are passed through as-is
			cleanedOptions = append(cleanedOptions, opt)
		}
	}

	// Process each option type
	for key, values := range optionCounts {
		if len(values) > 1 {
			// Handle duplicates
			switch key {
			case "i2cp.leaseSetEncType":
				// For encryption types, use the last specified value
				lastValue := values[len(values)-1]
				cleanedOptions = append(cleanedOptions, lastValue)
				logger.WithFields(logger.Fields{
					"duplicates": values,
					"resolvedTo": lastValue,
					"optionType": key,
				}).Warn("Multiple i2cp.leaseSetEncType entries detected - using last specified value")

			case "SIGNATURE_TYPE":
				// This should be handled by validateAndCleanOptions, but just in case
				lastValue := values[len(values)-1]
				cleanedOptions = append(cleanedOptions, lastValue)
				logger.WithFields(logger.Fields{
					"duplicates": values,
					"resolvedTo": lastValue,
				}).Warn("Multiple SIGNATURE_TYPE entries in options - using last specified value")

			default:
				// For other duplicates, use the last value and warn
				lastValue := values[len(values)-1]
				cleanedOptions = append(cleanedOptions, lastValue)
				logger.WithFields(logger.Fields{
					"duplicates": values,
					"resolvedTo": lastValue,
					"optionType": key,
				}).Warn("Duplicate option entries detected - using last specified value")
			}
		} else {
			// Single value, add as-is
			cleanedOptions = append(cleanedOptions, values[0])
		}
	}

	return cleanedOptions
}
