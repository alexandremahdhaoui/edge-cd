package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"libvirt.org/go/libvirt"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
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
	vmName := fmt.Sprintf("test-vm-%d", time.Now().UnixNano())
	imageURL := "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.qcow2"
	imagePath := filepath.Join(os.TempDir(), "ubuntu-24.04-server-cloudimg-amd64.qcow2")
	sshKeyPath := "~/.ssh/id_rsa" // Assuming default SSH key for now

	// Download image if not exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Logf("Downloading VM image from %s to %s...", imageURL, imagePath)
		cmd := exec.Command("wget", "-O", imagePath, imageURL)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to download VM image: %v\nOutput: %s", err, output)
		}
	}

	// Generate user data with SSH public key
	sshPublicKey, err := getSSHPublicKey(sshKeyPath)
	if err != nil {
		t.Fatalf("Failed to get SSH public key: %v", err)
	}

	userData := fmt.Sprintf(`
#cloud-config
hostname: %s
users:
  - name: ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - %s
`, vmName, sshPublicKey)

	cfg := NewVMConfig(vmName, imagePath, sshKeyPath)
	cfg.UserData = userData

	// --- Test VM Lifecycle ---
	conn, dom, err := CreateVM(cfg)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer func() {
		if err := DestroyVM(conn, dom); err != nil {
			t.Errorf("Failed to destroy VM: %v", err)
		}
	}()

	// Wait for VM to boot and get IP
	var ipAddress string
	for i := 0; i < 10; i++ { // Retry a few times
		ipAddress, err = GetVMIPAddress(conn, dom)
		if err == nil && ipAddress != "" {
			break
		}
		t.Logf("Waiting for VM IP address... attempt %d", i+1)
		time.Sleep(10 * time.Second)
	}

	if err != nil || ipAddress == "" {
		t.Fatalf("Failed to get VM IP address: %v", err)
	}
	t.Logf("VM %s has IP: %s", vmName, ipAddress)

	// Basic SSH connectivity test
	sshClient, err := ssh.NewClient(ipAddress, "ubuntu", cfg.SSHKeyPath, "22")
	if err != nil {
		t.Fatalf("Failed to create SSH client to VM: %v", err)
	}

	stdout, stderr, err := sshClient.Run("echo hello")
	if err != nil {
		t.Fatalf("Failed to run command on VM via SSH: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	if strings.TrimSpace(stdout) != "hello" {
		t.Fatalf("Unexpected SSH command output: got %q, want %q", strings.TrimSpace(stdout), "hello")
	}

	t.Log("VM lifecycle and basic SSH connectivity test passed.")
}

// getSSHPublicKey reads the public key from the given private key path.
func getSSHPublicKey(privateKeyPath string) (string, error) {
	// For now, assume id_rsa.pub exists next to id_rsa
	publicKeyPath := privateKeyPath + ".pub"
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		return "", fmt.Errorf("SSH public key not found at %s", publicKeyPath)
	}

	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH public key: %w", err)
	}

	return strings.TrimSpace(string(publicKeyBytes)), nil
}
