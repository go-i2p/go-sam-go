package common

import (
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// NewSAM creates a new SAM instance by connecting to the specified address,
// performing the hello handshake, and initializing the SAM resolver.
// It returns a pointer to the SAM instance or an error if any step fails.
// This function combines connection establishment and hello handshake into a single step,
// eliminating the need for separate helper functions.
// It also initializes the SAM resolver directly after the connection is established.
// The SAM instance is ready to use for further operations like session creation or name resolution.
func NewSAM(address string) (*SAM, error) {
	return NewSAMWithAuth(address, "", "")
}

// NewSAMWithAuth creates a new SAM instance with optional authentication support.
// This function supports SAMv3.2+ USER/PASSWORD authentication for connecting to
// authenticated SAM bridges. If user and password are empty, no authentication is used.
//
// Parameters:
//   - address: SAM bridge address (e.g., "127.0.0.1:7656")
//   - user: Username for authentication (empty string for no auth)
//   - password: Password for authentication (empty string for no auth)
//
// Returns a SAM instance ready for session creation or an error if connection fails.
func NewSAMWithAuth(address, user, password string) (*SAM, error) {
	logger := log.WithFields(logrus.Fields{
		"address": address,
		"user":    user,
		"auth":    user != "" || password != "",
	})
	logger.Debug("Creating new SAM instance")

	// Use existing helper function for connection establishment
	conn, err := connectToSAM(address)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to SAM bridge")
		return nil, err // connectToSAM already wraps the error appropriately
	}

	s := &SAM{
		Conn: conn,
	}

	// Configure authentication if provided
	s.SAMEmit.I2PConfig.User = user
	s.SAMEmit.I2PConfig.Password = password

	// Use existing helper function for hello handshake with proper cleanup
	if err := sendHelloAndValidate(conn, s); err != nil {
		logger.WithError(err).Error("Failed to complete SAM handshake")
		conn.Close()
		return nil, err // sendHelloAndValidate already wraps the error appropriately
	}

	// Configure SAM instance with address
	s.SAMEmit.I2PConfig.SetSAMAddress(address)

	// Initialize resolver
	resolver, err := NewSAMResolver(s)
	if err != nil {
		logger.WithError(err).Error("Failed to create SAM resolver")
		conn.Close()
		return nil, oops.Errorf("failed to create SAM resolver: %w", err)
	}
	s.SAMResolver = *resolver
	logger.Debug("Successfully created new SAM instance")

	return s, nil
}
