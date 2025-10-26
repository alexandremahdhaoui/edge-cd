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

		// Commands are now formatted with FormatCmd, so we expect quoted arguments
		expectedCloneCmd := execcontext.FormatCmd(ctx, "git", "clone", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		expectedUpdateCmd := execcontext.FormatCmd(ctx, "apt-get", "update")
		expectedInstallCmd := execcontext.FormatCmd(ctx, "apt-get", "install", "-y", "git", "curl")

		if err := provision.ProvisionPackages(ctx, mock, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			expectedCloneCmd,
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

		// Commands are now formatted with FormatCmd, so we expect quoted arguments
		expectedCloneCmd := execcontext.FormatCmd(ctx, "git", "clone", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		expectedUpdateCmd := execcontext.FormatCmd(ctx, "apt-get", "update")

		if err := provision.ProvisionPackages(ctx, mock, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			expectedCloneCmd,
			expectedUpdateCmd,
		}

		require.NoError(t, mock.AssertNumberOfCommandsRun(len(expectedCommands)))
		require.Equal(t, len(expectedCommands), len(mock.Commands))

		for i, cmd := range mock.Commands {
			assert.Equal(t, expectedCommands[i], cmd, "command at index %d mismatch", i)
		}
	})
}
