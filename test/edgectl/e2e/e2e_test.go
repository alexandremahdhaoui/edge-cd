package e2e

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	te2e "github.com/alexandremahdhaoui/edge-cd/pkg/test/e2e"
	"libvirt.org/go/libvirt"
)

// keepArtifacts flag to preserve test artifacts for manual inspection
var keepArtifacts = flag.Bool("keep-artifacts", false, "Keep test artifacts for manual inspection and debugging")

// TestE2EBootstrapCommand runs a complete end-to-end bootstrap test
func TestE2EBootstrapCommand(t *testing.T) {
	// Skip test if libvirt is not available or if running in CI without KVM
	if os.Getenv("CI") == "true" && os.Getenv("LIBVIRT_TEST") != "true" {
		t.Skip("Skipping libvirt VM lifecycle test in CI without LIBVIRT_TEST=true")
	}

	// Ensure libvirt connection is possible (basic check)
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		t.Skipf("Skipping libvirt VM lifecycle test: failed to connect to libvirt: %v", err)
	}
	conn.Close()

	ctx := execcontext.New(make(map[string]string), []string{})
	tempDir := t.TempDir()

	// Setup test environment
	setupConfig := te2e.SetupConfig{
		ArtifactDir:    filepath.Join(tempDir, "artifacts"),
		ImageCacheDir:  filepath.Join(os.TempDir(), "edgectl"),
		EdgeCDRepoPath: getEdgeCDRepoPath(t),
		DownloadImages: true,
	}

	testEnv, err := te2e.SetupTestEnvironment(ctx, setupConfig)
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}
	t.Logf("Created test environment: %s", testEnv.ID)

	// Cleanup: Destroy VMs and artifacts (unless keep-artifacts flag is set)
	defer func() {
		if !*keepArtifacts {
			if err := te2e.TeardownTestEnvironment(ctx, testEnv); err != nil {
				t.Logf("Warning: Failed to teardown test environment: %v", err)
			}
		}
	}()

	// Build edgectl binary
	binaryPath, err := te2e.BuildEdgectlBinary("../../../cmd/edgectl")
	if err != nil {
		t.Fatalf("Failed to build edgectl binary: %v", err)
	}

	// Execute bootstrap test
	executorConfig := te2e.ExecutorConfig{
		EdgectlBinaryPath: binaryPath,
		ConfigPath:        "./test/edgectl/e2e/config",
		ConfigSpec:        "config.yaml",
		Packages:          "git,curl,openssh-client",
		ServiceManager:    "systemd",
		PackageManager:    "apt",
	}

	if err := te2e.ExecuteBootstrapTest(ctx, testEnv, executorConfig); err != nil {
		testEnv.Status = "failed"
		t.Fatalf("Bootstrap test failed: %v", err)
	}

	// Test passed
	testEnv.Status = "passed"

	// Save artifacts if flag is set
	if *keepArtifacts {
		artifactStoreDir := filepath.Join(os.ExpandEnv("$HOME"), ".edge-cd", "e2e")
		if err := os.MkdirAll(artifactStoreDir, 0755); err != nil {
			t.Logf("Warning: Failed to create artifact store directory: %v", err)
		} else {
			artifactStoreFile := filepath.Join(artifactStoreDir, "artifacts.json")
			store := te2e.NewJSONArtifactStore(artifactStoreFile)
			if err := store.Save(ctx, testEnv); err != nil {
				t.Logf("Warning: Failed to save test environment to artifact store: %v", err)
			}
		}

		// Print summary for keep-artifacts mode
		t.Logf("\n=== Test Environment Summary ===")
		t.Logf("ID: %s", testEnv.ID)
		t.Logf("Status: %s", testEnv.Status)
		t.Logf("Artifact Path: %s", testEnv.ArtifactPath)
		t.Logf("\n=== VMs ===")
		t.Logf("Target VM: %s", testEnv.TargetVM.Name)
		t.Logf("  IP Address: %s", testEnv.TargetVM.IP)
		t.Logf("  SSH Key: %s", testEnv.SSHKeys.HostKeyPath)
		t.Logf("  SSH Command: ssh -i %s ubuntu@%s", testEnv.SSHKeys.HostKeyPath, testEnv.TargetVM.IP)
		t.Logf("\nGit Server VM: %s", testEnv.GitServerVM.Name)
		t.Logf("  IP Address: %s", testEnv.GitServerVM.IP)
		t.Logf("  SSH Key: %s", testEnv.SSHKeys.HostKeyPath)
		t.Logf("  SSH Command: ssh -i %s git@%s", testEnv.SSHKeys.HostKeyPath, testEnv.GitServerVM.IP)

		t.Logf("\n=== Git Repositories ===")
		for repoName, repoURL := range testEnv.GitSSHURLs {
			t.Logf("%s: %s", repoName, repoURL)
		}

		t.Logf("\n=== Artifacts Preserved ===")
		t.Logf("Test artifacts have been preserved in: %s", testEnv.ArtifactPath)
		t.Logf("SSH Key for host access: %s/", testEnv.SSHKeys.HostKeyPath)
		t.Logf("\nTo access the test environment:")
		t.Logf("  # Connect to target VM")
		t.Logf("  ssh -i %s ubuntu@%s", testEnv.SSHKeys.HostKeyPath, testEnv.TargetVM.IP)
		t.Logf("\n  # Connect to git server")
		t.Logf("  ssh -i %s git@%s", testEnv.SSHKeys.HostKeyPath, testEnv.GitServerVM.IP)
		t.Logf("\nTo cleanup artifacts later:")
		t.Logf("  # Destroy VMs and remove artifacts")
		t.Logf("  virsh destroy %s && virsh undefine --remove-all-storage %s", testEnv.TargetVM.Name, testEnv.TargetVM.Name)
		t.Logf("  virsh destroy %s && virsh undefine --remove-all-storage %s", testEnv.GitServerVM.Name, testEnv.GitServerVM.Name)
	}
}

// getEdgeCDRepoPath returns the path to the edge-cd repository
func getEdgeCDRepoPath(t *testing.T) string {
	t.Helper()
	b, err := exec.Command("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		t.Fatalf("Error getting current repo path: %v", err)
	}
	return strings.TrimSpace(string(b))
}

