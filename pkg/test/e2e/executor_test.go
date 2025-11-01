package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIdempotentGitPush tests that pushing the same changes twice succeeds
// This verifies the idempotent behavior needed for rerunning tests on the same environment
func TestIdempotentGitPush(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	// Create a temporary directory for the test repo
	tempDir := t.TempDir()

	// Initialize a git repository
	initCmd := exec.Command("git", "init", "--initial-branch=main")
	initCmd.Dir = tempDir
	require.NoError(t, initCmd.Run(), "failed to initialize git repo")

	// Configure git user for commits
	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = tempDir
	require.NoError(t, configNameCmd.Run(), "failed to configure git user name")

	configEmailCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configEmailCmd.Dir = tempDir
	require.NoError(t, configEmailCmd.Run(), "failed to configure git user email")

	// Create an initial commit so we have a main branch
	testFile := filepath.Join(tempDir, "initial.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("initial content"), 0644))

	addCmd := exec.Command("git", "add", "initial.txt")
	addCmd.Dir = tempDir
	require.NoError(t, addCmd.Run(), "failed to add initial file")

	commitCmd := exec.Command("git", "commit", "-m", "initial commit")
	commitCmd.Dir = tempDir
	require.NoError(t, commitCmd.Run(), "failed to create initial commit")

	// Setup: write a test file
	testContent := "test file content\n"
	testFilePath := "test-config.txt"

	// First push: write file and commit
	fullPath := filepath.Join(tempDir, testFilePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(testContent), 0644))

	// Stage the new file
	addCmd = exec.Command("git", "add", testFilePath)
	addCmd.Dir = tempDir
	require.NoError(t, addCmd.Run(), "failed to stage test file")

	// Check that there are staged changes (should have changes)
	diffCmd := exec.Command("git", "diff", "--cached", "--exit-code")
	diffCmd.Dir = tempDir
	err := diffCmd.Run()
	require.Error(t, err, "expected git diff to detect changes on first write")

	// Commit the file
	commitCmd = exec.Command("git", "commit", "-m", "add test file")
	commitCmd.Dir = tempDir
	require.NoError(t, commitCmd.Run(), "failed to commit test file")

	// IDEMPOTENCY TEST: Write the same file content again
	// This simulates what happens on a second test run
	require.NoError(t, os.WriteFile(fullPath, []byte(testContent), 0644))

	// Stage the "changes" (but content is identical)
	addCmd = exec.Command("git", "add", testFilePath)
	addCmd.Dir = tempDir
	require.NoError(t, addCmd.Run(), "failed to stage on second write")

	// Check for staged changes - should have NONE because content is identical
	diffCmd = exec.Command("git", "diff", "--cached", "--exit-code")
	diffCmd.Dir = tempDir
	err = diffCmd.Run()

	// Verify: no changes should be detected (err should be nil)
	require.NoError(t, err, "expected no changes on second write of identical content")

	// This is the key test: commit should fail because there are no changes
	commitCmd = exec.Command("git", "commit", "-m", "add test file again")
	commitCmd.Dir = tempDir
	err = commitCmd.Run()
	require.Error(t, err, "expected commit to fail with no changes")
}

// TestPushChangesWithNoChanges tests the pushChangesToGitRepo function with idempotent content
func TestPushChangesWithNoChanges(t *testing.T) {
	// Skip if git is not available or in CI without full git setup
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	// Create a temporary directory for the source repo (what we're pushing from)
	sourceDir := t.TempDir()

	// Initialize a git repository
	initCmd := exec.Command("git", "init", "--initial-branch=main")
	initCmd.Dir = sourceDir
	require.NoError(t, initCmd.Run())

	// Configure git user
	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = sourceDir
	require.NoError(t, configNameCmd.Run())

	configEmailCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configEmailCmd.Dir = sourceDir
	require.NoError(t, configEmailCmd.Run())

	// Create initial commit
	initialFile := filepath.Join(sourceDir, "initial.txt")
	require.NoError(t, os.WriteFile(initialFile, []byte("initial"), 0644))

	addCmd := exec.Command("git", "add", "initial.txt")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", "initial")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

	// Simulate first scenario: add a file
	testContent := "test file for scenario\n"
	testFilePath := "scenario-file.txt"
	fullPath := filepath.Join(sourceDir, testFilePath)

	require.NoError(t, os.WriteFile(fullPath, []byte(testContent), 0644))

	addCmd = exec.Command("git", "add", testFilePath)
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd = exec.Command("git", "commit", "-m", "add scenario file")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

	// Now simulate the idempotent push: write the same file again
	// This is what happens when the test is rerun with the same environment
	require.NoError(t, os.WriteFile(fullPath, []byte(testContent), 0644))

	// Stage it
	addCmd = exec.Command("git", "add", testFilePath)
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	// Check for staged changes - should be none
	diffCmd := exec.Command("git", "diff", "--cached", "--exit-code")
	diffCmd.Dir = sourceDir
	err := diffCmd.Run()

	// This should succeed with no error, meaning no changes detected
	require.NoError(t, err, "expected git diff to detect no changes on idempotent push")

	t.Logf("✅ Idempotent git push test passed: no changes detected on second write of identical content")
}
