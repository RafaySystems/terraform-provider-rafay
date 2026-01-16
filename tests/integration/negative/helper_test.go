//go:build !planonly
// +build !planonly

package rafay_test

import (
	"os"
	"testing"
)

// Minimal env so provider Configure() is happy in plan-only runs.
func setDummyEnv(t *testing.T) {
	_ = os.Setenv("RCTL_API_KEY", "dummy")
	_ = os.Setenv("RCTL_PROJECT", "default")
	_ = os.Setenv("RCTL_REST_ENDPOINT", "console.example.dev")
}
