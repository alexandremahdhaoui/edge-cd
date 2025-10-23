package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
	"libvirt.org/go/libvirt"
)

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

	// --- Configuration ---

	// Create a temporary directory for test artifacts
	tempDir := t.TempDir()
	cacheDir := filepath.Join(os.TempDir(), "edgectl")
	fmt.Println(cacheDir)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("failed to create cache directory for vm image %q", cacheDir)
	}

	vmName := fmt.Sprintf("test-vm-%d", time.Now().UnixNano())
	imageName := "ubuntu-24.04-server-cloudimg-amd64.img"
	imageURL := "https://cloud-images.ubuntu.com/releases/noble/release/" + imageName
	imageCachePath := filepath.Join(cacheDir, imageName)

	// Generate SSH key pair in the temporary directory
	sshKeyPath := filepath.Join(tempDir, "id_rsa")
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", sshKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate SSH key pair: %v\nOutput: %s", err, output)
	}

	// Set restrictive permissions on the private key file
	if err := os.Chmod(sshKeyPath, 0o600); err != nil {
		t.Fatalf("Failed to set permissions on SSH private key: %v", err)
	}

	// Download image if not exists
	if _, err := os.Stat(imageCachePath); os.IsNotExist(err) {
		t.Logf("Downloading VM image from %s to %s...", imageURL, imageCachePath)
		cmd := exec.Command(
			"wget",
			"--progress=dot",
			"-e", "dotbytes=3M",
			"-O", imageCachePath,
			imageURL,
		)

		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to download VM image: %v", err)
		}
	}

	// Generate user data with SSH public key
	sshPublicKey, err := getSSHPublicKey(sshKeyPath)
	if err != nil {
		t.Fatalf("Failed to get SSH public key: %v", err)
	}
	t.Logf("SSH Public Key: %s", sshPublicKey)

	targetUser := "ubuntu"
	userData := fmt.Sprintf(`
#cloud-config
hostname: %s
users:
  - name: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - %q
`, vmName, targetUser, sshPublicKey)

	cfg := vmm.NewVMConfig(vmName, imageCachePath, sshKeyPath)
	cfg.UserData = userData

	t.Logf("[INFO] Creating VM with config %+v", cfg)
	// --- Test VM Lifecycle ---
	vmConn, dom, err := vmm.CreateVM(cfg)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer func() {
		if err := vmm.DestroyVM(vmConn, dom); err != nil {
			t.Errorf("Failed to destroy VM: %v", err)
		}
	}()

	t.Log("[INFO] Retrieving VM IP Adrr...")
	var ipAddress string
	ipAddress, err = vmm.GetVMIPAddress(vmConn, dom)

	if err != nil || ipAddress == "" {
		t.Fatalf("Failed to get VM IP address: %v", err)
	}
	t.Logf("VM %s has IP: %s", vmName, ipAddress)

	// Wait for SSH to be ready
	sshClient, err := ssh.NewClient(ipAddress, "ubuntu", sshKeyPath, "22")
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}
	if err := sshClient.AwaitServer(30 * time.Second); err != nil {
		t.Fatalf("SSH server in VM %s did not become ready in time: %v", vmName, err)
	}

	// Build the edgectl binary
	binaryPath := buildEdgectlHelper(t)
	edgCDRepo := "https://github.com/alexandremahdhaoui/edge-cd.git"
	configPath := "./test/edgectl/e2e/config"
	configSpec := "config.yaml"

	cmd = exec.Command(
		binaryPath,
		"bootstrap",
		"--target", ipAddress,
		"--user", targetUser,
		"--config-repo", edgCDRepo,
		"--config-path", configPath,
		"--config-spec", configSpec,
		"--edge-cd-repo", edgCDRepo,
		"--packages", "git,curl",
		"--service-manager", "systemd",
		"--package-manager", "apt",
	)

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run `edgectl bootstrap`: %v", err)
	}

	// Check packages
	_, _, err = sshClient.Run("dpkg -s git && dpkg -s curl")
	if err != nil {
		t.Errorf("git and curl not installed: %v", err)
	}

	// Check repos
	_, _, err = sshClient.Run("[ -d /opt/edge-cd/.git ]")
	if err != nil {
		t.Errorf("edge-cd repo not found: %v", err)
	}

	_, _, err = sshClient.Run("[ -d /opt/user-config/.git ]")
	if err != nil {
		t.Errorf("user-config repo not found: %v", err)
	}

	// Check config file
	_, _, err = sshClient.Run("[ -f /etc/edge-cd/config.yaml ]")
	if err != nil {
		t.Errorf("config.yaml not found: %v", err)
	}

	// Check service file
	_, _, err = sshClient.Run("[ -f /etc/systemd/system/edge-cd.service ]")
	if err != nil {
		t.Errorf("edge-cd.service not found: %v", err)
	}

	// Check service status
	_, _, err = sshClient.Run("systemctl is-enabled edge-cd.service")
	if err != nil {
		t.Errorf("edge-cd.service not enabled: %v", err)
	}

	_, _, err = sshClient.Run("systemctl is-active edge-cd.service")
	if err != nil {
		t.Errorf("edge-cd.service not active: %v", err)
	}
}

// getSSHPublicKey reads the public key from the given private key path.
func getSSHPublicKey(privateKeyPath string) (string, error) {
	// For now, assume id_rsa.pub exists next to id_rsa
	publicKeyPath := privateKeyPath + ".pub"
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		return "", fmt.Errorf("SSH public key not found at %s", publicKeyPath)
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH public key: %w", err)
	}

	return strings.TrimSpace(string(publicKeyBytes)), nil
}

// buildEdgectlHelper builds the edgectl binary and returns its path.
// It creates a temporary directory for the binary and cleans it up after the test.
func buildEdgectlHelper(t *testing.T) string {
	t.Helper()

	// Create a temporary directory for the built binary
	tmpDir, err := os.MkdirTemp("", "edgectl-build-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	binaryPath := filepath.Join(tmpDir, "edgectl")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../../cmd/edgectl")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build edgectl binary: %v", err)
	}

	return binaryPath
}
