package files

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alexandremahdhaoui/edge-cd/pkg/userconfig"
)

// FileReconciler reconciles file specifications to ensure files on the system
// match those defined in the configuration repository.
type FileReconciler interface {
	ReconcileFiles(configRepoPath, configPath string, files []userconfig.FileSpec) (*ReconcileResult, error)
}

// fileReconciler is the implementation of FileReconciler.
type fileReconciler struct{}

// ReconcileResult contains the results of file reconciliation.
type ReconcileResult struct {
	ServicesToRestart []string
	RequiresReboot    bool
}

// NewFileReconciler creates a new FileReconciler instance.
func NewFileReconciler() FileReconciler {
	return &fileReconciler{}
}

// ReconcileFiles reconciles all file specifications.
func (fr *fileReconciler) ReconcileFiles(configRepoPath, configPath string, files []userconfig.FileSpec) (*ReconcileResult, error) {
	result := &ReconcileResult{
		ServicesToRestart: []string{},
	}

	for _, file := range files {
		switch file.Type {
		case "file":
			if err := fr.reconcileFile(configRepoPath, configPath, file, result); err != nil {
				return nil, err
			}
		case "directory":
			if err := fr.reconcileDirectory(configRepoPath, configPath, file, result); err != nil {
				return nil, err
			}
		case "content":
			if err := fr.reconcileContent(file, result); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown file type: %s", file.Type)
		}
	}

	return result, nil
}

// reconcileFile reconciles a single file from the config repository.
func (fr *fileReconciler) reconcileFile(configRepoPath, configPath string, file userconfig.FileSpec, result *ReconcileResult) error {
	srcPath := filepath.Join(configRepoPath, configPath, file.SrcPath)
	destPath := file.DestPath

	// Check if files are identical (drift detection)
	if filesEqual(srcPath, destPath) {
		return nil // No drift
	}

	// Drift detected - copy file
	slog.Info("Drift detected: updating file", "destPath", destPath)

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy file
	if err := copyFile(srcPath, destPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Set permissions
	fileMode := parseFileMode(file.FileMod)
	if err := os.Chmod(destPath, fileMode); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Track services to restart
	if file.SyncBehavior != nil {
		result.ServicesToRestart = append(result.ServicesToRestart, file.SyncBehavior.RestartServices...)
		if file.SyncBehavior.Reboot {
			result.RequiresReboot = true
		}
	}

	return nil
}

// reconcileDirectory reconciles all files from a directory in the config repository.
func (fr *fileReconciler) reconcileDirectory(configRepoPath, configPath string, file userconfig.FileSpec, result *ReconcileResult) error {
	srcDirPath := filepath.Join(configRepoPath, configPath, file.SrcPath)
	destDirPath := file.DestPath

	// Ensure destination directory exists
	if err := os.MkdirAll(destDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Walk the source directory and copy all files
	return filepath.Walk(srcDirPath, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if srcPath == srcDirPath {
			return nil
		}

		// Compute relative path
		relPath, err := filepath.Rel(srcDirPath, srcPath)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}

		destPath := filepath.Join(destDirPath, relPath)

		if info.IsDir() {
			// Create subdirectory
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			return nil
		}

		// Check if files are identical (drift detection)
		if filesEqual(srcPath, destPath) {
			return nil // No drift
		}

		// Drift detected - copy file
		slog.Info("Drift detected: updating file", "destPath", destPath)

		// Copy file
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}

		// Set permissions
		fileMode := parseFileMode(file.FileMod)
		if err := os.Chmod(destPath, fileMode); err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}

		// Track services to restart
		if file.SyncBehavior != nil {
			result.ServicesToRestart = append(result.ServicesToRestart, file.SyncBehavior.RestartServices...)
			if file.SyncBehavior.Reboot {
				result.RequiresReboot = true
			}
		}

		return nil
	})
}

// reconcileContent reconciles inline content to a file.
func (fr *fileReconciler) reconcileContent(file userconfig.FileSpec, result *ReconcileResult) error {
	destPath := file.DestPath

	// Check if destination file exists and matches content
	existingContent, err := os.ReadFile(destPath)
	if err == nil && bytes.Equal(existingContent, []byte(file.Content)) {
		return nil // No drift
	}

	// Drift detected - write content
	slog.Info("Drift detected: updating file", "destPath", destPath)

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write content
	fileMode := parseFileMode(file.FileMod)
	if err := os.WriteFile(destPath, []byte(file.Content), fileMode); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Track services to restart
	if file.SyncBehavior != nil {
		result.ServicesToRestart = append(result.ServicesToRestart, file.SyncBehavior.RestartServices...)
		if file.SyncBehavior.Reboot {
			result.RequiresReboot = true
		}
	}

	return nil
}

// filesEqual compares two files byte-by-byte (equivalent to cmp command).
func filesEqual(path1, path2 string) bool {
	// Read both files
	data1, err1 := os.ReadFile(path1)
	data2, err2 := os.ReadFile(path2)

	// If either read fails, they're not equal
	if err1 != nil || err2 != nil {
		return false
	}

	// Byte-by-byte comparison (equivalent to cmp)
	return bytes.Equal(data1, data2)
}

// parseFileMode parses an octal file mode string (e.g., "755" → 0755).
// Defaults to 0644 for invalid input.
func parseFileMode(modeStr string) os.FileMode {
	if modeStr == "" {
		return 0644
	}

	// Parse octal string (e.g., "755" → 0755)
	mode, err := strconv.ParseUint(modeStr, 8, 32)
	if err != nil {
		slog.Error("Invalid file mode, using default 0644", "mode", modeStr)
		return 0644
	}

	return os.FileMode(mode)
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644) // chmod happens after
}
