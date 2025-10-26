package provision_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
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

		// Set expected responses for the mock runner
		mock.SetResponse(fmt.Sprintf("git clone %s %s", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath), "", "", nil)
		mock.SetResponse("apt-get update", "", "", nil)
		mock.SetResponse("apt-get install -y git curl", "", "", nil)

		if err := provision.ProvisionPackages(mock, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			fmt.Sprintf("git clone %s %s", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath),
			"apt-get update",
			"apt-get install -y git curl",
		}

		if err := mock.AssertNumberOfCommandsRun(len(expectedCommands)); err != nil {
			t.Fatal(err)
		}

		for i, cmd := range mock.Commands {
			if cmd != expectedCommands[i] {
				t.Errorf("expected command '%s' at index %d, got '%s'", expectedCommands[i], i, cmd)
			}
		}
	})

	t.Run("should do nothing if no packages are provided", func(t *testing.T) {
		mock := ssh.NewMockRunner()
		var packages []string
		localPkgMgrRepoPath := tmpDir
		remoteEdgeCDRepoURL := "https://github.com/alexandremahdhaoui/edge-cd.git"
		remoteEdgeCDRepoDestPath := "/usr/local/src/edge-cd"

		// Set expected responses for the mock runner
		mock.SetResponse(fmt.Sprintf("git clone %s %s", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath), "", "", nil)
		mock.SetResponse("apt-get update", "", "", nil)

		if err := provision.ProvisionPackages(mock, packages, "apt", localPkgMgrRepoPath, remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			fmt.Sprintf("git clone %s %s", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath),
			"apt-get update",
		}

		if err := mock.AssertNumberOfCommandsRun(len(expectedCommands)); err != nil {
			t.Fatal(err)
		}

		for i, cmd := range mock.Commands {
			if cmd != expectedCommands[i] {
				t.Errorf("expected command '%s' at index %d, got '%s'", expectedCommands[i], i, cmd)
			}
		}
	})
}
