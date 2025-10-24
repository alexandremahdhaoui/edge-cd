package vmm_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/cloudinit"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
)

func TestVMMStructLifecycle(t *testing.T) {
	// Skip test if libvirt is not available or if running in CI without KVM
	if os.Getenv("CI") == "true" && os.Getenv("LIBVIRT_TEST") != "true" {
		t.Skip("Skipping libvirt VM lifecycle test in CI without LIBVIRT_TEST=true")
	}

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

	// -- Generate user data with VM's SSH public key
	publicKeyBytes, err := os.ReadFile(sshKeyPath + ".pub")
	if err != nil {
		t.Fatal(err.Error())
	}
	targetUser := cloudinit.NewUserWithAuthorizedKeys("ubuntu", []string{string(publicKeyBytes)})

	userData := cloudinit.UserData{
		Hostname:      vmName,
		PackageUpdate: true,
		Packages:      []string{"qemu-guest-agent"},
		Users:         []cloudinit.User{targetUser},
		WriteFiles: []cloudinit.WriteFile{
			{
				Path:        "/etc/systemd/system/mnt-virtiofs.mount",
				Permissions: "0644",
				Content: `[Unit]
Description=VirtioFS Mount
After=network-online.target

[Service]
Restart=always

[Mount]
What=virtiofs_share
Where=/mnt/virtiofs
Type=virtiofs
Options=defaults,nofail

[Install]
WantedBy=multi-user.target`,
			},
		},
		RunCommands: []string{
			"mkdir -p /mnt/virtiofs",
		},
	}

	// Define virtiofs share for the VM
	virtiofsSharePath := filepath.Join(tempDir, "virtiofs_share")
	if err := os.MkdirAll(virtiofsSharePath, 0o755); err != nil {
		t.Fatalf("Failed to create virtiofs share directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(virtiofsSharePath, "host_file.txt"), []byte("Hello from host!"), 0o644); err != nil {
		t.Fatalf("Failed to write host_file.txt: %v", err)
	}

	// -- create new vm config
	cfg := vmm.NewVMConfig(
		vmName,
		imageCachePath,
		userData,
	)
	cfg.VirtioFS = []vmm.VirtioFSConfig{
		{
			Tag:        "virtiofs_share", // Changed to match cloud-init mount tag
			MountPoint: virtiofsSharePath,
		},
	}

	vmmInstance, err := vmm.NewVMM()
	if err != nil {
		t.Fatalf("Failed to create VMM instance: %v", err)
	}
	defer vmmInstance.Close()

	t.Logf("[INFO] Creating VM with config %+v", cfg)
	// --- Test VM Lifecycle ---
	_, err = vmmInstance.CreateVM(cfg)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	defer func() {
		if err := vmmInstance.DestroyVM(vmName); err != nil {
			t.Errorf("Failed to destroy VM: %v", err)
		}
	}()

	t.Log("[INFO] Retrieving VM IP Adrr...")
	var ipAddress string
	ipAddress, err = vmmInstance.GetVMIPAddress(vmName)

	if err != nil || ipAddress == "" {
		t.Fatalf("Failed to get VM IP address: %v", err)
	}
	t.Logf("VM %s has IP: %s", vmName, ipAddress)

	// Get and log serial console output
	consoleOutput, err := vmmInstance.GetConsoleOutput(vmName)
	if err != nil {
		t.Logf("Failed to get console output: %v", err)
	} else {
		t.Logf("\n--- VM Console Output ---\n%s\n-------------------------", consoleOutput)
	}

	// Retry SSH connection and command execution
	var sshClient *ssh.Client
	var stdout, stderr string
	var sshErr error

	sshTimeout := time.After(60 * time.Second) // Increased timeout for VM startup
	sshTick := time.NewTicker(5 * time.Second)
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
			sshClient, sshErr = ssh.NewClient(ipAddress, "ubuntu", sshKeyPath, "22")
			if sshErr != nil {
				t.Logf("SSH connection failed: %v, retrying...", sshErr)
				continue
			}

			// Verify basic SSH connectivity
			stdout, stderr, sshErr = sshClient.Run("echo hello")
			if sshErr != nil || strings.TrimSpace(stdout) != "hello" {
				t.Logf(
					"Failed to run basic command on VM via SSH: %v\nStdout: %s\nStderr: %s, retrying...",
					sshErr,
					stdout,
					stderr,
				)
				continue
			}
			t.Log("VM lifecycle and basic SSH connectivity test passed.")

			stdout, stderr, sshErr = sshClient.Run("sudo systemctl enable --now mnt-virtiofs.mount")
			if sshErr != nil {
				t.Logf(
					"Error running 'sudo systemctl enable --now mnt-virtiofs.mount' on VM: %v\nStdout: %s\nStderr: %s",
					sshErr,
					stdout,
					stderr,
				)
			} else {
				t.Logf("VM 'systemctl status mnt-virtiofs.mount' output:\n%s", stdout)
			}

			// Verify virtiofs mount
			stdout, stderr, sshErr = sshClient.Run("ls /mnt/virtiofs/host_file.txt")
			if sshErr != nil || !strings.Contains(stdout, "host_file.txt") {
				t.Errorf(
					"VirtioFS mount not working or host_file.txt not found: %v\nStdout: %s\nStderr: %s",
					sshErr,
					stdout,
					stderr,
				)
			} else {
				t.Log("VirtioFS mount verified.")
			}

			return // Test passed
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

