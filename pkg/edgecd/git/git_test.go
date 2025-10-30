package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) string {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	// Initialize git repo with explicit branch name
	cmd := exec.Command("git", "init", "-b", "master")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user name: %v", err)
	}

	// Create a test file and commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	return tmpDir
}

func TestNewRepoManager(t *testing.T) {
	mgr := NewRepoManager()
	if mgr == nil {
		t.Fatal("NewRepoManager returned nil")
	}

	_, ok := mgr.(*gitRepoManager)
	if !ok {
		t.Fatal("NewRepoManager did not return *gitRepoManager")
	}
}

func TestGetCurrentCommit(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := NewRepoManager()

	commit, err := mgr.GetCurrentCommit(repoPath)
	if err != nil {
		t.Fatalf("GetCurrentCommit failed: %v", err)
	}

	if commit == "" {
		t.Fatal("GetCurrentCommit returned empty commit hash")
	}

	// Verify commit hash is 40 characters (SHA-1)
	if len(commit) != 40 {
		t.Fatalf("GetCurrentCommit returned invalid commit hash length: %d (expected 40)", len(commit))
	}
}

func TestGetCurrentCommit_NonGitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "non-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	mgr := NewRepoManager()
	_, err = mgr.GetCurrentCommit(tmpDir)
	if err == nil {
		t.Fatal("GetCurrentCommit should fail for non-git repository")
	}
}

func TestGetCommitDiff(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := NewRepoManager()

	// Get first commit
	firstCommit, err := mgr.GetCurrentCommit(repoPath)
	if err != nil {
		t.Fatalf("GetCurrentCommit failed: %v", err)
	}

	// Create another file and commit
	testFile := filepath.Join(repoPath, "test2.txt")
	if err := os.WriteFile(testFile, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := exec.Command("git", "add", "test2.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Second commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Get second commit
	secondCommit, err := mgr.GetCurrentCommit(repoPath)
	if err != nil {
		t.Fatalf("GetCurrentCommit failed: %v", err)
	}

	// Get diff
	files, err := mgr.GetCommitDiff(repoPath, firstCommit, secondCommit)
	if err != nil {
		t.Fatalf("GetCommitDiff failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 changed file, got %d", len(files))
	}

	if files[0] != "test2.txt" {
		t.Fatalf("Expected changed file 'test2.txt', got '%s'", files[0])
	}
}

func TestGetCommitDiff_NoChanges(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := NewRepoManager()

	commit, err := mgr.GetCurrentCommit(repoPath)
	if err != nil {
		t.Fatalf("GetCurrentCommit failed: %v", err)
	}

	// Diff same commit
	files, err := mgr.GetCommitDiff(repoPath, commit, commit)
	if err != nil {
		t.Fatalf("GetCommitDiff failed: %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("Expected 0 changed files, got %d", len(files))
	}
}

func TestCloneRepo_FileURL(t *testing.T) {
	mgr := NewRepoManager()

	// file:// URLs should be skipped without error
	err := mgr.CloneRepo("file:///tmp/test", "main", "/tmp/dest", []string{})
	if err != nil {
		t.Fatalf("CloneRepo should succeed for file:// URL: %v", err)
	}
}

func TestCloneRepo_RealRepo(t *testing.T) {
	// This test requires a source repo to clone from
	sourceRepo := setupTestRepo(t)

	// Create destination directory
	destDir, err := os.MkdirTemp("", "git-clone-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(destDir)
	})

	cloneDest := filepath.Join(destDir, "cloned")

	mgr := NewRepoManager()

	// Clone the repo (using file:// URL for local clone)
	// Note: We'll test with a real git URL pattern but use the local filesystem
	// Use "." for sparse checkout to get all files
	err = mgr.CloneRepo(sourceRepo, "master", cloneDest, []string{"."})
	if err != nil {
		t.Fatalf("CloneRepo failed: %v", err)
	}

	// Verify cloned repo exists and has .git directory
	if _, err := os.Stat(filepath.Join(cloneDest, ".git")); err != nil {
		t.Fatalf("Cloned repo missing .git directory: %v", err)
	}

	// Verify test file exists
	if _, err := os.Stat(filepath.Join(cloneDest, "test.txt")); err != nil {
		t.Fatalf("Cloned repo missing test.txt: %v", err)
	}
}

func TestCloneRepo_InvalidURL(t *testing.T) {
	destDir, err := os.MkdirTemp("", "git-clone-invalid-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(destDir)
	})

	cloneDest := filepath.Join(destDir, "cloned")
	mgr := NewRepoManager()

	// Try to clone from non-existent URL
	err = mgr.CloneRepo("https://invalid-url-that-does-not-exist.com/repo.git", "main", cloneDest, []string{"."})
	if err == nil {
		t.Fatal("CloneRepo should fail for invalid URL")
	}

	// The error could be from any git operation during clone
	// We just need to ensure some git error occurred
	if err == nil {
		t.Fatal("Expected git operation to fail")
	}
}

func TestSyncRepo(t *testing.T) {
	// Create source repo
	sourceRepo := setupTestRepo(t)

	// Create a test-branch in source repo
	cmd := exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = sourceRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}

	// Add a file on the new branch
	testFile := filepath.Join(sourceRepo, "branch-file.txt")
	if err := os.WriteFile(testFile, []byte("branch content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd = exec.Command("git", "add", "branch-file.txt")
	cmd.Dir = sourceRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Branch commit")
	cmd.Dir = sourceRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Clone the source repo
	destDir, err := os.MkdirTemp("", "git-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(destDir)
	})

	cloneDest := filepath.Join(destDir, "cloned")

	mgr := NewRepoManager()

	// Clone the repo on master branch
	err = mgr.CloneRepo(sourceRepo, "master", cloneDest, []string{"."})
	if err != nil {
		t.Fatalf("CloneRepo failed: %v", err)
	}

	// Verify branch-file.txt doesn't exist yet (we're on master)
	if _, err := os.Stat(filepath.Join(cloneDest, "branch-file.txt")); err == nil {
		t.Fatal("branch-file.txt should not exist on master branch")
	}

	// Sync to test-branch
	err = mgr.SyncRepo(cloneDest, "test-branch", []string{"."})
	if err != nil {
		t.Fatalf("SyncRepo failed: %v", err)
	}

	// Verify we're now on the branch with the new file
	if _, err := os.Stat(filepath.Join(cloneDest, "branch-file.txt")); err != nil {
		t.Fatalf("SyncRepo did not sync to branch: %v", err)
	}
}

func TestSyncRepo_NonGitDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "non-git-sync-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	mgr := NewRepoManager()

	// SyncRepo should skip non-git directories gracefully
	err = mgr.SyncRepo(tmpDir, "main", []string{"*"})
	if err != nil {
		t.Fatalf("SyncRepo should skip non-git directory gracefully: %v", err)
	}
}

func TestMockRepoManager(t *testing.T) {
	mock := &MockRepoManager{
		GetCurrentCommitFunc: func(repoPath string) (string, error) {
			return "test-commit-123", nil
		},
		GetCommitDiffFunc: func(repoPath, oldCommit, newCommit string) ([]string, error) {
			return []string{"file1.txt", "file2.txt"}, nil
		},
	}

	// Test GetCurrentCommit
	commit, err := mock.GetCurrentCommit("/test/path")
	if err != nil {
		t.Fatalf("MockRepoManager.GetCurrentCommit failed: %v", err)
	}
	if commit != "test-commit-123" {
		t.Fatalf("Expected 'test-commit-123', got '%s'", commit)
	}

	// Test GetCommitDiff
	files, err := mock.GetCommitDiff("/test/path", "old", "new")
	if err != nil {
		t.Fatalf("MockRepoManager.GetCommitDiff failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}

	// Test default behavior (no func set)
	mock2 := &MockRepoManager{}
	commit, err = mock2.GetCurrentCommit("/test/path")
	if err != nil {
		t.Fatalf("MockRepoManager with no func should not fail: %v", err)
	}
	if commit != "mock-commit-hash" {
		t.Fatalf("Expected default 'mock-commit-hash', got '%s'", commit)
	}
}
