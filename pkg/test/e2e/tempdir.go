package e2e

import (
	"fmt"
	"os"
	"path/filepath"
)

// TempDirMarkerFile is the name of the marker file that identifies a managed temp directory
const TempDirMarkerFile = ".edge-cd-e2e-temp"

// CreateTempDirectory creates a new temporary directory with a marker file indicating it's managed by e2e tests
// If the directory already exists, it will be reused and the marker file will be created/verified
// Returns the absolute path to the created directory
func CreateTempDirectory(dirPath string) (string, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create temp directory %s: %w", dirPath, err)
	}

	// Create marker file to indicate this is a managed temp directory
	markerPath := filepath.Join(dirPath, TempDirMarkerFile)
	if err := os.WriteFile(markerPath, []byte(""), 0o644); err != nil {
		return "", fmt.Errorf("failed to create marker file in %s: %w", dirPath, err)
	}

	return dirPath, nil
}

// IsManagedTempDirectory checks if a directory is a managed temporary directory
// A directory is considered managed if:
// 1. It exists
// 2. It is a directory (not a file)
// 3. It contains the marker file (.edge-cd-e2e-temp)
func IsManagedTempDirectory(dirPath string) bool {
	// Check if path exists and is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}

	// Check if marker file exists
	markerPath := filepath.Join(dirPath, TempDirMarkerFile)
	if _, err := os.Stat(markerPath); err != nil {
		return false
	}

	return true
}
