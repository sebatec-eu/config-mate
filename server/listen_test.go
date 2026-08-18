package server

import "testing"

// TestListenAndServeAddr is a table-driven test of the listenAddr() pure
// function used by [ListenAndServe]'s HTTP branch.
//
// Each subtest explicitly sets ADDR and PORT with t.Setenv (never iterating
// over a map) so the precedence rule is unambiguous. t.Setenv also seeds
// both vars to "" where unset, so the case does not depend on the outer
// process environment.
//
// Moved verbatim from hostsharing/hostsharing_test.go — the precedence rule
// (ADDR > PORT > default) is part of ListenAndServe's contract.
func TestListenAndServeAddr(t *testing.T) {
	tests := []struct {
		name       string
		addrEnv    string
		portEnv    string
		defaultHTTPPort string
		expected   string
	}{
		{
			name:     "ADDR wins when set",
			addrEnv:  "127.0.0.1:8080",
			portEnv:  "9000",
			expected: "127.0.0.1:8080",
		},
		{
			name:     "PORT used when ADDR is empty",
			addrEnv:  "",
			portEnv:  "8080",
			expected: ":8080",
		},
		{
			name:     "default port used when neither set",
			addrEnv:  "",
			portEnv:  "",
			expected: ":9000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ADDR", tc.addrEnv)
			t.Setenv("PORT", tc.portEnv)

			if got := listenAddr(); got != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, got)
			}
		})
	}
}
