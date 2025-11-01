package provision_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvisionPackages(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "prov-pkg-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the package-managers directory structure
	pkgMgrDir := filepath.Join(tmpDir, "cmd", "edge-cd", "package-managers")
	if err := os.MkdirAll(pkgMgrDir, 0o755); err != nil {
		t.Fatalf("Failed to create package-managers dir: %v", err)
	}

	aptYaml := `
update: ["apt-get", "update"]
install: ["apt-get", "install", "-y"]
`
	if err := ioutil.WriteFile(filepath.Join(pkgMgrDir, "apt.yaml"), []byte(aptYaml), 0o644); err != nil {
		t.Fatalf("Failed to write apt.yaml: %v", err)
	}

	t.Run("should install multiple packages with apt", func(t *testing.T) {
		mock := ssh.NewMockRunner()
		packages := []string{"git", "curl"}
		localPkgMgrRepoPath := tmpDir
		remoteEdgeCDRepoURL := "https://github.com/alexandremahdhaoui/edge-cd.git"
		remoteEdgeCDRepoDestPath := "/usr/local/src/edge-cd"

		// Create context with no env vars or prepend commands
		ctx := execcontext.New(make(map[string]string), []string{})

		// Commands for sparse checkout clone sequence
		expectedTestCmd := execcontext.FormatCmd(ctx, "test", "-d", remoteEdgeCDRepoDestPath)
		expectedCloneCmd := execcontext.FormatCmd(ctx, "git", "clone", "--filter=blob:none", "--no-checkout", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		expectedSparseInitCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "init")
		expectedSparseSetCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "set", "cmd/edge-cd")
		expectedCheckoutCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "checkout", "main")
		expectedFetchCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "fetch", "origin", "main")
		expectedPullCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "pull")
		expectedUpdateCmd := execcontext.FormatCmd(ctx, "apt-get", "update")
		expectedInstallCmd := execcontext.FormatCmd(ctx, "apt-get", "install", "-y", "git", "curl")

		// Simulate directory doesn't exist (test -d fails)
		mock.SetResponse(expectedTestCmd, "", "", assert.AnError)

		if err := provision.ProvisionPackages(ctx, mock, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			expectedTestCmd,
			expectedCloneCmd,
			expectedSparseInitCmd,
			expectedSparseSetCmd,
			expectedCheckoutCmd,
			expectedFetchCmd,
			expectedPullCmd,
			expectedUpdateCmd,
			expectedInstallCmd,
		}

		require.NoError(t, mock.AssertNumberOfCommandsRun(len(expectedCommands)))
		require.Equal(t, len(expectedCommands), len(mock.Commands))

		for i, cmd := range mock.Commands {
			assert.Equal(t, expectedCommands[i], cmd, "command at index %d mismatch", i)
		}
	})

	t.Run("should do nothing if no packages are provided", func(t *testing.T) {
		mock := ssh.NewMockRunner()
		var packages []string
		localPkgMgrRepoPath := tmpDir
		remoteEdgeCDRepoURL := "https://github.com/alexandremahdhaoui/edge-cd.git"
		remoteEdgeCDRepoDestPath := "/usr/local/src/edge-cd"

		// Create context with no env vars or prepend commands
		ctx := execcontext.New(make(map[string]string), []string{})

		// Commands for sparse checkout clone sequence
		expectedTestCmd := execcontext.FormatCmd(ctx, "test", "-d", remoteEdgeCDRepoDestPath)
		expectedCloneCmd := execcontext.FormatCmd(ctx, "git", "clone", "--filter=blob:none", "--no-checkout", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		expectedSparseInitCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "init")
		expectedSparseSetCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "set", "cmd/edge-cd")
		expectedCheckoutCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "checkout", "main")
		expectedFetchCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "fetch", "origin", "main")
		expectedPullCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "pull")
		expectedUpdateCmd := execcontext.FormatCmd(ctx, "apt-get", "update")

		// Simulate directory doesn't exist (test -d fails)
		mock.SetResponse(expectedTestCmd, "", "", assert.AnError)

		if err := provision.ProvisionPackages(ctx, mock, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			expectedTestCmd,
			expectedCloneCmd,
			expectedSparseInitCmd,
			expectedSparseSetCmd,
			expectedCheckoutCmd,
			expectedFetchCmd,
			expectedPullCmd,
			expectedUpdateCmd,
		}

		require.NoError(t, mock.AssertNumberOfCommandsRun(len(expectedCommands)))
		require.Equal(t, len(expectedCommands), len(mock.Commands))

		for i, cmd := range mock.Commands {
			assert.Equal(t, expectedCommands[i], cmd, "command at index %d mismatch", i)
		}
	})

	t.Run("should be idempotent - clone on first run, sync on second run", func(t *testing.T) {
		mock := ssh.NewMockRunner()
		packages := []string{"git", "curl"}
		localPkgMgrRepoPath := tmpDir
		remoteEdgeCDRepoURL := "https://github.com/alexandremahdhaoui/edge-cd.git"
		remoteEdgeCDRepoDestPath := "/usr/local/src/edge-cd"

		// Create context with no env vars or prepend commands
		ctx := execcontext.New(make(map[string]string), []string{})

		// First call - simulate repository doesn't exist
		// The test -d command should fail (exit code 1)
		testDirCmd := execcontext.FormatCmd(ctx, "test", "-d", remoteEdgeCDRepoDestPath)
		mock.SetResponse(testDirCmd, "", "", assert.AnError) // test -d fails = dir doesn't exist

		// Run ProvisionPackages first time
		err := provision.ProvisionPackages(ctx, mock, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		require.NoError(t, err, "First call to ProvisionPackages should succeed")

		// Verify first call executed git clone with sparse checkout (not sync)
		expectedCloneCmd := execcontext.FormatCmd(ctx, "git", "clone", "--filter=blob:none", "--no-checkout", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		expectedSparseInitCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "init")
		firstCallCommands := make([]string, len(mock.Commands))
		copy(firstCallCommands, mock.Commands)

		assert.Contains(t, firstCallCommands, expectedCloneCmd, "First call should execute git clone with sparse checkout")
		assert.Contains(t, firstCallCommands, expectedSparseInitCmd, "First call should initialize sparse checkout")

		// Verify git reset was NOT called in first run (only used for sync)
		expectedResetCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "reset", "--hard", "FETCH_HEAD")
		assert.NotContains(t, firstCallCommands, expectedResetCmd, "First call should NOT execute git reset")

		// Second call - simulate repository exists
		// Create new mock for second call to have clean command history
		mock2 := ssh.NewMockRunner()

		// Now test -d should succeed (repository exists)
		mock2.SetResponse(testDirCmd, "", "", nil) // test -d succeeds = dir exists

		// Run ProvisionPackages second time
		err = provision.ProvisionPackages(ctx, mock2, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		require.NoError(t, err, "Second call to ProvisionPackages should succeed")

		// Verify second call executed sync commands (sparse-checkout set, fetch, reset)
		expectedSyncSparseSetCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "set", "cmd/edge-cd")
		expectedSyncFetchCmd := execcontext.FormatCmd(ctx, "git", "-C", remoteEdgeCDRepoDestPath, "fetch", "origin", "main")
		assert.Contains(t, mock2.Commands, expectedSyncSparseSetCmd, "Second call should set sparse checkout")
		assert.Contains(t, mock2.Commands, expectedSyncFetchCmd, "Second call should fetch from origin")
		assert.Contains(t, mock2.Commands, expectedResetCmd, "Second call should reset to FETCH_HEAD")

		// Verify second call did NOT execute clone
		assert.NotContains(t, mock2.Commands, expectedCloneCmd, "Second call should NOT execute git clone")
		assert.NotContains(t, mock2.Commands, expectedSparseInitCmd, "Second call should NOT initialize sparse checkout")

		// Both calls should succeed - proving idempotency
		t.Log("✓ ProvisionPackages is idempotent: first call clones with sparse checkout, second call syncs with fetch+reset")
	})
}
