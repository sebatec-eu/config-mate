package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// serviceNameEnvVar is the environment variable consulted first when
// resolving the service name. Exported indirectly via ServiceName.
const serviceNameEnvVar = "SERVICE_NAME"

// serviceName resolves a service name. SERVICE_NAME env wins; otherwise it
// falls back to the executable's base name with any trailing ".fcgi" stripped.
// The executable lookup is injected for testability.
func serviceName(fn func() (string, error)) (string, error) {
	if name := getenv(serviceNameEnvVar); name != "" {
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

// ServiceName returns the name of the service by checking the SERVICE_NAME
// environment variable first, and falling back to the executable name if the
// variable is not set. The ".fcgi" suffix is trimmed from the executable name
// if present.
//
// It returns an error if neither the environment variable is set nor the
// executable path can be determined, or if the resulting name is empty.
func ServiceName() (string, error) {
	return serviceName(executablePath)
}
