package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/userconfig"
)

func TestNewFileReconciler(t *testing.T) {
	fr := NewFileReconciler()
	if fr == nil {
		t.Fatal("NewFileReconciler() returned nil")
	}
}

func TestFilesEqual(t *testing.T) {
	tests := []struct {
		name     string
		content1 string
		content2 string
		want     bool
	}{
		{
			name:     "identical files",
			content1: "hello world",
			content2: "hello world",
			want:     true,
		},
		{
			name:     "different files",
			content1: "hello",
			content2: "world",
			want:     false,
		},
		{
			name:     "empty files",
			content1: "",
			content2: "",
			want:     true,
		},
		{
			name:     "one empty one not",
			content1: "hello",
			content2: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create two test files
			file1 := filepath.Join(tmpDir, "file1.txt")
			file2 := filepath.Join(tmpDir, "file2.txt")

			if err := os.WriteFile(file1, []byte(tt.content1), 0644); err != nil {
				t.Fatalf("Failed to create file1: %v", err)
			}

			if err := os.WriteFile(file2, []byte(tt.content2), 0644); err != nil {
				t.Fatalf("Failed to create file2: %v", err)
			}

			got := filesEqual(file1, file2)
			if got != tt.want {
				t.Errorf("filesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilesEqual_NonExistentFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	// Create only file1
	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	// file2 doesn't exist
	if filesEqual(file1, file2) {
		t.Error("filesEqual() should return false when one file doesn't exist")
	}

	// Neither file exists
	file3 := filepath.Join(tmpDir, "file3.txt")
	file4 := filepath.Join(tmpDir, "file4.txt")
	if filesEqual(file3, file4) {
		t.Error("filesEqual() should return false when both files don't exist")
	}
}

func TestParseFileMode(t *testing.T) {
	tests := []struct {
		name    string
		modeStr string
		want    os.FileMode
	}{
		{
			name:    "empty string defaults to 0644",
			modeStr: "",
			want:    0644,
		},
		{
			name:    "valid mode 755",
			modeStr: "755",
			want:    0755,
		},
		{
			name:    "valid mode 644",
			modeStr: "644",
			want:    0644,
		},
		{
			name:    "valid mode 600",
			modeStr: "600",
			want:    0600,
		},
		{
			name:    "valid mode 777",
			modeStr: "777",
			want:    0777,
		},
		{
			name:    "invalid mode defaults to 0644",
			modeStr: "invalid",
			want:    0644,
		},
		{
			name:    "decimal number defaults to 0644",
			modeStr: "999",
			want:    0644,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFileMode(tt.modeStr)
			if got != tt.want {
				t.Errorf("parseFileMode(%q) = %o, want %o", tt.modeStr, got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	content := "test content"

	// Create source file
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify destination file exists and has correct content
	gotContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(gotContent) != content {
		t.Errorf("Destination file content = %q, want %q", gotContent, content)
	}
}

func TestReconcileContent(t *testing.T) {
	tmpDir := t.TempDir()
	fr := NewFileReconciler().(*fileReconciler)

	tests := []struct {
		name              string
		file              userconfig.FileSpec
		existingContent   string
		shouldExist       bool
		wantRestart       []string
		wantReboot        bool
		expectedContent   string
		expectedFileCount int // Number of times file should be written (0 = no drift, 1 = drift)
	}{
		{
			name: "new file with content",
			file: userconfig.FileSpec{
				Type:     "content",
				DestPath: filepath.Join(tmpDir, "new-content.txt"),
				Content:  "hello from content",
				FileMod:  "644",
			},
			shouldExist:       false,
			expectedContent:   "hello from content",
			expectedFileCount: 1,
		},
		{
			name: "file exists with same content - no drift",
			file: userconfig.FileSpec{
				Type:     "content",
				DestPath: filepath.Join(tmpDir, "same-content.txt"),
				Content:  "same content",
				FileMod:  "644",
			},
			shouldExist:       true,
			existingContent:   "same content",
			expectedContent:   "same content",
			expectedFileCount: 0,
		},
		{
			name: "file exists with different content - drift detected",
			file: userconfig.FileSpec{
				Type:     "content",
				DestPath: filepath.Join(tmpDir, "diff-content.txt"),
				Content:  "new content",
				FileMod:  "644",
			},
			shouldExist:       true,
			existingContent:   "old content",
			expectedContent:   "new content",
			expectedFileCount: 1,
		},
		{
			name: "content with service restart",
			file: userconfig.FileSpec{
				Type:     "content",
				DestPath: filepath.Join(tmpDir, "service-restart.txt"),
				Content:  "trigger restart",
				FileMod:  "644",
				SyncBehavior: &userconfig.SyncBehavior{
					RestartServices: []string{"nginx", "apache"},
				},
			},
			shouldExist:       false,
			expectedContent:   "trigger restart",
			wantRestart:       []string{"nginx", "apache"},
			expectedFileCount: 1,
		},
		{
			name: "content with reboot flag",
			file: userconfig.FileSpec{
				Type:     "content",
				DestPath: filepath.Join(tmpDir, "reboot.txt"),
				Content:  "trigger reboot",
				FileMod:  "755",
				SyncBehavior: &userconfig.SyncBehavior{
					Reboot: true,
				},
			},
			shouldExist:       false,
			expectedContent:   "trigger reboot",
			wantReboot:        true,
			expectedFileCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup existing file if needed
			if tt.shouldExist {
				if err := os.WriteFile(tt.file.DestPath, []byte(tt.existingContent), 0644); err != nil {
					t.Fatalf("Failed to create existing file: %v", err)
				}
			}

			result := &ReconcileResult{
				ServicesToRestart: []string{},
			}

			err := fr.reconcileContent(tt.file, result)
			if err != nil {
				t.Fatalf("reconcileContent() error = %v", err)
			}

			// Verify file content
			gotContent, err := os.ReadFile(tt.file.DestPath)
			if err != nil {
				t.Fatalf("Failed to read destination file: %v", err)
			}

			if string(gotContent) != tt.expectedContent {
				t.Errorf("File content = %q, want %q", gotContent, tt.expectedContent)
			}

			// Verify services to restart
			if len(result.ServicesToRestart) != len(tt.wantRestart) {
				t.Errorf("ServicesToRestart length = %d, want %d", len(result.ServicesToRestart), len(tt.wantRestart))
			}

			// Verify reboot flag
			if result.RequiresReboot != tt.wantReboot {
				t.Errorf("RequiresReboot = %v, want %v", result.RequiresReboot, tt.wantReboot)
			}
		})
	}
}

func TestReconcileFile(t *testing.T) {
	tmpDir := t.TempDir()
	configRepoPath := filepath.Join(tmpDir, "config-repo")
	configPath := "devices/router1"
	fr := NewFileReconciler().(*fileReconciler)

	// Create config repo directory structure
	srcDir := filepath.Join(configRepoPath, configPath, "files")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	tests := []struct {
		name            string
		file            userconfig.FileSpec
		srcContent      string
		destContent     string
		destExists      bool
		wantRestart     []string
		wantReboot      bool
		wantDestContent string
	}{
		{
			name: "new file - drift detected",
			file: userconfig.FileSpec{
				Type:     "file",
				SrcPath:  "files/test1.txt",
				DestPath: filepath.Join(tmpDir, "dest", "test1.txt"),
				FileMod:  "644",
			},
			srcContent:      "source content",
			destExists:      false,
			wantDestContent: "source content",
		},
		{
			name: "identical files - no drift",
			file: userconfig.FileSpec{
				Type:     "file",
				SrcPath:  "files/test2.txt",
				DestPath: filepath.Join(tmpDir, "dest", "test2.txt"),
				FileMod:  "644",
			},
			srcContent:      "same content",
			destContent:     "same content",
			destExists:      true,
			wantDestContent: "same content",
		},
		{
			name: "different files - drift detected",
			file: userconfig.FileSpec{
				Type:     "file",
				SrcPath:  "files/test3.txt",
				DestPath: filepath.Join(tmpDir, "dest", "test3.txt"),
				FileMod:  "755",
			},
			srcContent:      "new content",
			destContent:     "old content",
			destExists:      true,
			wantDestContent: "new content",
		},
		{
			name: "file with service restart",
			file: userconfig.FileSpec{
				Type:     "file",
				SrcPath:  "files/test4.txt",
				DestPath: filepath.Join(tmpDir, "dest", "test4.txt"),
				FileMod:  "644",
				SyncBehavior: &userconfig.SyncBehavior{
					RestartServices: []string{"test-service"},
				},
			},
			srcContent:      "trigger restart",
			destExists:      false,
			wantRestart:     []string{"test-service"},
			wantDestContent: "trigger restart",
		},
		{
			name: "file with reboot",
			file: userconfig.FileSpec{
				Type:     "file",
				SrcPath:  "files/test5.txt",
				DestPath: filepath.Join(tmpDir, "dest", "test5.txt"),
				FileMod:  "644",
				SyncBehavior: &userconfig.SyncBehavior{
					Reboot: true,
				},
			},
			srcContent:      "trigger reboot",
			destExists:      false,
			wantReboot:      true,
			wantDestContent: "trigger reboot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create source file
			srcPath := filepath.Join(configRepoPath, configPath, tt.file.SrcPath)
			if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
				t.Fatalf("Failed to create source directory: %v", err)
			}
			if err := os.WriteFile(srcPath, []byte(tt.srcContent), 0644); err != nil {
				t.Fatalf("Failed to create source file: %v", err)
			}

			// Create destination file if needed
			if tt.destExists {
				if err := os.MkdirAll(filepath.Dir(tt.file.DestPath), 0755); err != nil {
					t.Fatalf("Failed to create dest directory: %v", err)
				}
				if err := os.WriteFile(tt.file.DestPath, []byte(tt.destContent), 0644); err != nil {
					t.Fatalf("Failed to create dest file: %v", err)
				}
			}

			result := &ReconcileResult{
				ServicesToRestart: []string{},
			}

			err := fr.reconcileFile(configRepoPath, configPath, tt.file, result)
			if err != nil {
				t.Fatalf("reconcileFile() error = %v", err)
			}

			// Verify destination file content
			gotContent, err := os.ReadFile(tt.file.DestPath)
			if err != nil {
				t.Fatalf("Failed to read destination file: %v", err)
			}

			if string(gotContent) != tt.wantDestContent {
				t.Errorf("Destination content = %q, want %q", gotContent, tt.wantDestContent)
			}

			// Verify services to restart
			if len(result.ServicesToRestart) != len(tt.wantRestart) {
				t.Errorf("ServicesToRestart length = %d, want %d", len(result.ServicesToRestart), len(tt.wantRestart))
			}

			// Verify reboot flag
			if result.RequiresReboot != tt.wantReboot {
				t.Errorf("RequiresReboot = %v, want %v", result.RequiresReboot, tt.wantReboot)
			}
		})
	}
}

func TestReconcileDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configRepoPath := filepath.Join(tmpDir, "config-repo")
	configPath := "devices/router1"
	fr := NewFileReconciler().(*fileReconciler)

	// Create config repo directory structure with multiple files
	srcDir := filepath.Join(configRepoPath, configPath, "config-dir")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create source files
	files := map[string]string{
		"file1.txt":        "content 1",
		"file2.txt":        "content 2",
		"subdir/file3.txt": "content 3",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(srcDir, relPath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create source file %s: %v", relPath, err)
		}
	}

	destDir := filepath.Join(tmpDir, "dest-dir")

	file := userconfig.FileSpec{
		Type:     "directory",
		SrcPath:  "config-dir",
		DestPath: destDir,
		FileMod:  "644",
		SyncBehavior: &userconfig.SyncBehavior{
			RestartServices: []string{"test-service"},
		},
	}

	result := &ReconcileResult{
		ServicesToRestart: []string{},
	}

	err := fr.reconcileDirectory(configRepoPath, configPath, file, result)
	if err != nil {
		t.Fatalf("reconcileDirectory() error = %v", err)
	}

	// Verify all files were copied
	for relPath, expectedContent := range files {
		destPath := filepath.Join(destDir, relPath)
		gotContent, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("Failed to read destination file %s: %v", relPath, err)
			continue
		}

		if string(gotContent) != expectedContent {
			t.Errorf("File %s content = %q, want %q", relPath, gotContent, expectedContent)
		}
	}

	// Verify services to restart (should be added for each file)
	if len(result.ServicesToRestart) == 0 {
		t.Error("Expected services to restart to be populated")
	}
}

func TestReconcileFiles(t *testing.T) {
	tmpDir := t.TempDir()
	configRepoPath := filepath.Join(tmpDir, "config-repo")
	configPath := "devices/router1"
	fr := NewFileReconciler()

	// Create config repo directory structure
	srcDir := filepath.Join(configRepoPath, configPath, "files")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create a source file
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	files := []userconfig.FileSpec{
		{
			Type:     "file",
			SrcPath:  "files/test.txt",
			DestPath: filepath.Join(tmpDir, "dest", "test.txt"),
			FileMod:  "644",
			SyncBehavior: &userconfig.SyncBehavior{
				RestartServices: []string{"service1"},
			},
		},
		{
			Type:     "content",
			DestPath: filepath.Join(tmpDir, "dest", "content.txt"),
			Content:  "inline content",
			FileMod:  "644",
			SyncBehavior: &userconfig.SyncBehavior{
				RestartServices: []string{"service2"},
				Reboot:          true,
			},
		},
	}

	result, err := fr.ReconcileFiles(configRepoPath, configPath, files)
	if err != nil {
		t.Fatalf("ReconcileFiles() error = %v", err)
	}

	// Verify result
	if !result.RequiresReboot {
		t.Error("Expected RequiresReboot to be true")
	}

	if len(result.ServicesToRestart) != 2 {
		t.Errorf("ServicesToRestart length = %d, want 2", len(result.ServicesToRestart))
	}

	// Verify files exist
	destFile := filepath.Join(tmpDir, "dest", "test.txt")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	contentFile := filepath.Join(tmpDir, "dest", "content.txt")
	if _, err := os.Stat(contentFile); os.IsNotExist(err) {
		t.Error("Content file was not created")
	}
}

func TestReconcileFiles_UnknownType(t *testing.T) {
	tmpDir := t.TempDir()
	fr := NewFileReconciler()

	files := []userconfig.FileSpec{
		{
			Type:     "unknown-type",
			DestPath: filepath.Join(tmpDir, "test.txt"),
		},
	}

	_, err := fr.ReconcileFiles("", "", files)
	if err == nil {
		t.Error("Expected error for unknown file type, got nil")
	}

	if err.Error() != "unknown file type: unknown-type" {
		t.Errorf("Error message = %q, want %q", err.Error(), "unknown file type: unknown-type")
	}
}

func TestReconcileFile_PermissionsSet(t *testing.T) {
	tmpDir := t.TempDir()
	configRepoPath := filepath.Join(tmpDir, "config-repo")
	configPath := "devices/router1"
	fr := NewFileReconciler().(*fileReconciler)

	// Create config repo directory structure
	srcDir := filepath.Join(configRepoPath, configPath, "files")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	destFile := filepath.Join(tmpDir, "dest", "test.txt")

	file := userconfig.FileSpec{
		Type:     "file",
		SrcPath:  "files/test.txt",
		DestPath: destFile,
		FileMod:  "755",
	}

	result := &ReconcileResult{
		ServicesToRestart: []string{},
	}

	err := fr.reconcileFile(configRepoPath, configPath, file, result)
	if err != nil {
		t.Fatalf("reconcileFile() error = %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	gotMode := info.Mode().Perm()
	wantMode := os.FileMode(0755)

	if gotMode != wantMode {
		t.Errorf("File permissions = %o, want %o", gotMode, wantMode)
	}
}
