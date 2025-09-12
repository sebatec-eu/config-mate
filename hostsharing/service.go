// Package hostsharing provides utilities for Hostsharing environments.
package hostsharing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const serviceNameEnvVar = "SERVICE_NAME"

func serviceName(fn func() (string, error)) (string, error) {
	if name := os.Getenv(serviceNameEnvVar); name != "" {
		return name, nil
	}

	r, err := fn()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	name := strings.TrimSuffix(filepath.Base(r), ".fcgi")
	if name == "" {
		return "", fmt.Errorf("service name is empty after trimming suffix")
	}
	return name, nil
}

// ServiceName returns the name of the service by checking the SERVICE_NAME environment variable first,
// and falling back to the executable path if the variable is not set.
// It trims the ".fcgi" suffix if present.
// Returns the service name or an error if neither the environment variable nor the executable path can be determined.
func ServiceName() (string, error) {
	return serviceName(os.Executable)
}
