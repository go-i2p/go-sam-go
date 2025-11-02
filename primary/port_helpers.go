package primary

import (
	"fmt"

	"github.com/samber/oops"
)

// ensureUniqueStreamPort ensures that stream subsessions have unique port assignments.
// If no port options are provided, auto-assigns a unique port to prevent SAM bridge
// "Duplicate protocol" errors. Returns the final options, assigned port (0 if none), and error.
func (p *PrimarySession) ensureUniqueStreamPort(options []string) ([]string, int, error) {
	// Check if any port options are already specified
	hasPort := false
	for _, opt := range options {
		if len(opt) >= 11 && (opt[:11] == "LISTEN_PORT" || opt[:11] == "listen_port") {
			hasPort = true
			break
		}
		if len(opt) >= 9 && (opt[:9] == "FROM_PORT" || opt[:9] == "from_port") {
			hasPort = true
			break
		}
	}

	// If port is already specified, use options as-is
	if hasPort {
		return options, 0, nil
	}

	// Auto-assign a unique port to prevent duplicate protocol errors
	port, err := p.allocatePort()
	if err != nil {
		return nil, 0, oops.Errorf("failed to allocate unique port: %w", err)
	}

	// Add the auto-assigned port as FROM_PORT
	result := make([]string, len(options)+1)
	copy(result, options)
	result[len(options)] = fmt.Sprintf("FROM_PORT=%d", port)

	log.WithField("auto_assigned_port", port).Debug("Auto-assigned unique port for stream subsession")
	return result, port, nil
}

// allocatePort finds and reserves the next available port for stream subsessions.
// Returns the allocated port number or an error if no ports are available.
func (p *PrimarySession) allocatePort() (int, error) {
	// Search for an available port starting from nextAutoPort
	maxAttempts := 1000 // Prevent infinite loops
	for attempts := 0; attempts < maxAttempts; attempts++ {
		port := p.nextAutoPort
		if !p.usedPorts[port] {
			// Found an available port, reserve it
			p.usedPorts[port] = true
			p.nextAutoPort = port + 1
			if p.nextAutoPort > 65535 {
				p.nextAutoPort = 49152 // Wrap around to start of dynamic range
			}
			return port, nil
		}
		p.nextAutoPort++
		if p.nextAutoPort > 65535 {
			p.nextAutoPort = 49152 // Wrap around to start of dynamic range
		}
	}
	return 0, oops.Errorf("no available ports after %d attempts", maxAttempts)
}

// releasePort marks a port as available for reuse by future subsessions.
func (p *PrimarySession) releasePort(port int) {
	if port > 0 {
		delete(p.usedPorts, port)
		log.WithField("released_port", port).Debug("Released port for reuse")
	}
}

// buildOptionsWithPorts constructs options with explicit FROM_PORT and TO_PORT parameters.
// Validates and reserves the specified ports to prevent conflicts. Returns the final options
// and a slice of reserved port numbers for cleanup on failure.
func (p *PrimarySession) buildOptionsWithPorts(options []string, fromPort, toPort int) ([]string, []int, error) {
	var reservedPorts []int

	// Validate and reserve fromPort if specified
	if fromPort > 0 {
		if err := p.validateAndReservePort(fromPort); err != nil {
			return nil, nil, oops.Errorf("FROM_PORT %d unavailable: %w", fromPort, err)
		}
		reservedPorts = append(reservedPorts, fromPort)
	}

	// Validate and reserve toPort if specified and different from fromPort
	if toPort > 0 && toPort != fromPort {
		if err := p.validateAndReservePort(toPort); err != nil {
			// Release fromPort if we reserved it
			if fromPort > 0 {
				p.releasePort(fromPort)
			}
			return nil, nil, oops.Errorf("TO_PORT %d unavailable: %w", toPort, err)
		}
		reservedPorts = append(reservedPorts, toPort)
	}

	// Build options by filtering out existing port options and adding new ones
	finalOptions := p.filterPortOptions(options)

	// Add FROM_PORT if specified
	if fromPort > 0 {
		finalOptions = append(finalOptions, fmt.Sprintf("FROM_PORT=%d", fromPort))
	}

	// Add TO_PORT if specified and different from fromPort
	if toPort > 0 && toPort != fromPort {
		finalOptions = append(finalOptions, fmt.Sprintf("TO_PORT=%d", toPort))
	}

	log.WithFields(map[string]interface{}{
		"from_port":      fromPort,
		"to_port":        toPort,
		"reserved_ports": reservedPorts,
	}).Debug("Built options with explicit ports")

	return finalOptions, reservedPorts, nil
}

// validateAndReservePort checks if a port is available and reserves it.
func (p *PrimarySession) validateAndReservePort(port int) error {
	if port <= 0 || port > 65535 {
		return oops.Errorf("invalid port number: %d", port)
	}

	if p.usedPorts[port] {
		return oops.Errorf("port already in use")
	}

	// Reserve the port
	p.usedPorts[port] = true
	return nil
}

// filterPortOptions removes existing port-related options from the options slice.
func (p *PrimarySession) filterPortOptions(options []string) []string {
	var filtered []string
	for _, opt := range options {
		// Skip existing port options
		if len(opt) >= 9 && (opt[:9] == "FROM_PORT" || opt[:9] == "from_port") {
			continue
		}
		if len(opt) >= 7 && (opt[:7] == "TO_PORT" || opt[:7] == "to_port") {
			continue
		}
		if len(opt) >= 11 && (opt[:11] == "LISTEN_PORT" || opt[:11] == "listen_port") {
			continue
		}
		filtered = append(filtered, opt)
	}
	return filtered
}
