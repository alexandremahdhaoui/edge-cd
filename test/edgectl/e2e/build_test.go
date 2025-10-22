package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildEdgectlHelper builds the edgectl binary and returns its path.
// It creates a temporary directory for the binary and cleans it up after the test.
func buildEdgectlHelper(t *testing.T) string {
	t.Helper()

	// Create a temporary directory for the built binary
	tmpDir, err := os.MkdirTemp("", "edgectl-build-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	binaryPath := filepath.Join(tmpDir, "edgectl")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../../cmd/edgectl")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build edgectl binary: %v", err)
	}

	return binaryPath
}

func TestBuildHelper(t *testing.T) {
	binaryPath := buildEdgectlHelper(t)

	// Assert that the binary file exists
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("Failed to stat binary at %s: %v", binaryPath, err)
	}
	if info.IsDir() {
		t.Fatalf("Expected %s to be a file, but it's a directory", binaryPath)
	}
	if info.Mode().Perm()&0o111 == 0 { // Check for execute permissions
		t.Fatalf("Expected binary at %s to be executable, but it is not", binaryPath)
	}
}

