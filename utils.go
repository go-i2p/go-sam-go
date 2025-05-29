package sam3

import (
	"crypto/rand"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// RandString generates a random string
func RandString() string {
	const chars = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 12)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// PrimarySessionString returns primary session string
func PrimarySessionString() string {
	return "primary"
}

// SAMDefaultAddr returns default SAM address
func SAMDefaultAddr(fallforward string) string {
	if fallforward != "" {
		return fallforward
	}
	return SAM_HOST + ":" + SAM_PORT
}

// ExtractDest extracts destination from input
func ExtractDest(input string) string {
	parts := strings.Fields(input)
	for _, part := range parts {
		if strings.HasPrefix(part, "DEST=") {
			return part[5:]
		}
	}
	return ""
}

// ExtractPairString extracts string value from key=value pair
func ExtractPairString(input, value string) string {
	prefix := value + "="
	parts := strings.Fields(input)
	for _, part := range parts {
		if strings.HasPrefix(part, prefix) {
			return part[len(prefix):]
		}
	}
	return ""
}

// ExtractPairInt extracts integer value from key=value pair
func ExtractPairInt(input, value string) int {
	str := ExtractPairString(input, value)
	if str == "" {
		return 0
	}
	i, _ := strconv.Atoi(str)
	return i
}

// GenerateOptionString generates option string from slice
func GenerateOptionString(opts []string) string {
	return strings.Join(opts, " ")
}

// IgnorePortError ignores port-related errors
func IgnorePortError(err error) error {
	if err != nil && strings.Contains(err.Error(), "port") {
		return nil
	}
	return err
}

// Logging functions
var sam3Logger *logrus.Logger

// InitializeSAM3Logger initializes the logger
func InitializeSAM3Logger() {
	sam3Logger = logrus.New()
	sam3Logger.SetLevel(logrus.InfoLevel)
}

// GetSAM3Logger returns the initialized logger
func GetSAM3Logger() *logrus.Logger {
	if sam3Logger == nil {
		InitializeSAM3Logger()
	}
	return sam3Logger
}

// Additional utility functions that may be needed for compatibility
func ConvertOptionsToSlice(opts Options) []string {
	return opts.AsList()
}

func ConvertSliceToOptions(slice []string) Options {
	opts := make(Options)
	for _, opt := range slice {
		parts := strings.SplitN(opt, "=", 2)
		if len(parts) == 2 {
			opts[parts[0]] = parts[1]
		}
	}
	return opts
}
