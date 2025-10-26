package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCreateTempDirectory tests that a temp directory is created with marker file
func TestCreateTempDirectory(t *testing.T) {
	// Create a temporary directory for testing
	baseDir := t.TempDir()
	testDir := filepath.Join(baseDir, "test-temp-dir")

	// Call CreateTempDirectory
	path, err := CreateTempDirectory(testDir)
	if err != nil {
		t.Fatalf("CreateTempDirectory failed: %v", err)
	}

	// Verify directory was created
	if path != testDir {
		t.Errorf("Expected path %s, got %s", testDir, path)
	}

	// Verify directory exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("Path exists but is not a directory")
	}

	// Verify marker file exists
	markerPath := filepath.Join(path, TempDirMarkerFile)
	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("Marker file does not exist: %v", err)
	}
}

// TestCreateTempDirectory_AlreadyExists tests that an existing directory is preserved
func TestCreateTempDirectory_AlreadyExists(t *testing.T) {
	baseDir := t.TempDir()
	testDir := filepath.Join(baseDir, "existing-dir")

	// Create directory first
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Call CreateTempDirectory on existing directory
	path, err := CreateTempDirectory(testDir)
	if err != nil {
		t.Fatalf("CreateTempDirectory failed on existing directory: %v", err)
	}

	if path != testDir {
		t.Errorf("Expected path %s, got %s", testDir, path)
	}

	// Verify marker file was created
	markerPath := filepath.Join(path, TempDirMarkerFile)
	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("Marker file does not exist: %v", err)
	}
}

// TestIsManagedTempDirectory_Valid tests detection of valid temp directory
func TestIsManagedTempDirectory_Valid(t *testing.T) {
	baseDir := t.TempDir()
	testDir := filepath.Join(baseDir, "valid-temp")

	// Create temp directory with marker
	_, err := CreateTempDirectory(testDir)
	if err != nil {
		t.Fatalf("CreateTempDirectory failed: %v", err)
	}

	// Verify it's detected as managed
	if !IsManagedTempDirectory(testDir) {
		t.Errorf("IsManagedTempDirectory returned false for valid temp directory")
	}
}

// TestIsManagedTempDirectory_Invalid_NoMarkerFile tests detection of directory without marker
func TestIsManagedTempDirectory_Invalid_NoMarkerFile(t *testing.T) {
	baseDir := t.TempDir()
	testDir := filepath.Join(baseDir, "no-marker")

	// Create directory without marker file
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Verify it's NOT detected as managed
	if IsManagedTempDirectory(testDir) {
		t.Errorf("IsManagedTempDirectory returned true for directory without marker")
	}
}

// TestIsManagedTempDirectory_Invalid_NotExist tests detection of non-existent directory
func TestIsManagedTempDirectory_Invalid_NotExist(t *testing.T) {
	nonExistent := "/tmp/this-definitely-does-not-exist-12345-67890"

	// Verify it's NOT detected as managed
	if IsManagedTempDirectory(nonExistent) {
		t.Errorf("IsManagedTempDirectory returned true for non-existent directory")
	}
}

// TestIsManagedTempDirectory_Invalid_IsFile tests detection when path is a file
func TestIsManagedTempDirectory_Invalid_IsFile(t *testing.T) {
	baseDir := t.TempDir()
	filePath := filepath.Join(baseDir, "not-a-directory.txt")

	// Create a file instead of directory
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify it's NOT detected as managed
	if IsManagedTempDirectory(filePath) {
		t.Errorf("IsManagedTempDirectory returned true for a file")
	}
}
