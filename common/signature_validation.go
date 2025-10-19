package common

import (
	"strings"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
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
			logrus.WithFields(logrus.Fields{
				"sigType":           sigType,
				"conflictingOption": opt,
			}).Warn("Signature type conflict detected: sigType parameter takes precedence")
			// Skip this option - sigType parameter takes precedence
			continue
		}
		cleanedOptions = append(cleanedOptions, opt)
	}

	if conflictDetected {
		logrus.WithFields(logrus.Fields{
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
		logrus.WithField("duplicateSignatureTypes", signatureTypes).Error("Multiple SIGNATURE_TYPE entries found in options")
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
