package vmm_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"libvirt.org/go/libvirt"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
)

func TestVMLifecycle(t *testing.T) {
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

	userData := fmt.Sprintf(`
#cloud-config
hostname: %s
users:
  - name: ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - %q
`, vmName, sshPublicKey)

	cfg := vmm.NewVMConfig(vmName, imageCachePath, sshKeyPath)
	cfg.UserData = userData

	t.Logf("[INFO] Creating VM with config %+v", cfg)
	// --- Test VM Lifecycle ---
	conn, dom, err := vmm.CreateVM(cfg)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer func() {
		if err := vmm.DestroyVM(conn, dom); err != nil {
			t.Errorf("Failed to destroy VM: %v", err)
		}
	}()

	t.Log("[INFO] Retrieving VM IP Adrr...")
	var ipAddress string
	ipAddress, err = vmm.GetVMIPAddress(conn, dom)

	if err != nil || ipAddress == "" {
		t.Fatalf("Failed to get VM IP address: %v", err)
	}
	t.Logf("VM %s has IP: %s", vmName, ipAddress)

	// Get and log serial console output
	consoleOutput, err := vmm.GetConsoleOutput(conn, dom)
	if err != nil {
		t.Logf("Failed to get console output: %v", err)
	} else {
		t.Logf("\n--- VM Console Output ---\n%s\n-------------------------", consoleOutput)
	}

	// Retry SSH connection and command execution
	var sshClient *ssh.Client
	var stdout, stderr string
	var sshErr error

	sshTimeout := time.After(30 * time.Second)
	sshTick := time.NewTicker(3 * time.Second)
	defer sshTick.Stop()

	for {
		select {
		case <-sshTimeout:
			t.Fatalf(
				"Timed out waiting for SSH connection to VM %s at %s: %v",
				vmName,
				ipAddress,
				sshErr,
			)

		case <-sshTick.C:
			sshClient, sshErr = ssh.NewClient(ipAddress, "ubuntu", cfg.SSHKeyPath, "22")
			if sshErr != nil {
				t.Logf("SSH connection failed: %v, retrying...", sshErr)
				continue
			}

			stdout, stderr, sshErr = sshClient.Run("echo hello")
			if sshErr == nil {
				if strings.TrimSpace(stdout) == "hello" {
					t.Log("VM lifecycle and basic SSH connectivity test passed.")
					return // Test passed
				}
				t.Logf(
					"Unexpected SSH command output: got %q, want %q, retrying...",
					strings.TrimSpace(stdout),
					"hello",
				)
			} else {
				t.Logf("Failed to run command on VM via SSH: %v\nStdout: %s\nStderr: %s, retrying...", sshErr, stdout, stderr)
			}
		}
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
