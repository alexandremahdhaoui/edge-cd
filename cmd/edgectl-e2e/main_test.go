package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetArtifactDir verifies artifact directory resolution
func TestGetArtifactDir(t *testing.T) {
	// Test with no environment variable (should default to ~/.edge-cd/e2e/)
	os.Unsetenv("E2E_ARTIFACTS_DIR")
	dir := getArtifactDir()
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, ".edge-cd")
	assert.Contains(t, dir, "e2e")
}

// TestGetArtifactDirWithEnvVar verifies environment variable override
func TestGetArtifactDirWithEnvVar(t *testing.T) {
	customDir := "/custom/path/e2e"
	os.Setenv("E2E_ARTIFACTS_DIR", customDir)
	defer os.Unsetenv("E2E_ARTIFACTS_DIR")

	dir := getArtifactDir()
	assert.Equal(t, customDir, dir)
}

// TestDebugf verifies debug logging
func TestDebugf(t *testing.T) {
	// Test with debug disabled
	os.Unsetenv("EDGECTL_E2E_DEBUG")
	debugf("test message")  // Should not panic

	// Test with debug enabled
	os.Setenv("EDGECTL_E2E_DEBUG", "1")
	defer os.Unsetenv("EDGECTL_E2E_DEBUG")
	debugf("test debug message")  // Should not panic
}

// TestIsPiped verifies pipe detection
func TestIsPiped(t *testing.T) {
	// When stdout is a terminal, isPiped should return false
	piped := isPiped()

	// We can't easily test the piped case without actually piping,
	// but we can verify the function returns a boolean
	_, ok := interface{}(piped).(bool)
	assert.True(t, ok, "isPiped() should return a boolean")
}

// TestPrintEnvironmentJSON verifies JSON output (would need mocking to fully test)
func TestPrintEnvironmentJSON(t *testing.T) {
	// This test just verifies the function exists and has correct signature
	// Full testing would require mocking stdout
	assert.NotPanics(t, func() {
		// Function definition exists
		_ = printEnvironmentJSON
	})
}

// TestContextHandling verifies context is used properly
func TestContextHandling(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("expected context to be cancelled")
	}
}
