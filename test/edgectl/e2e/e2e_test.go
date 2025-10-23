package e2e

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/cloudinit"
	"github.com/alexandremahdhaoui/edge-cd/pkg/gitserver"
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

	// Generate SSH key pair for the host (for connecting to Git server)
	hostSSHKeyPath := filepath.Join(tempDir, "id_rsa_host")
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", hostSSHKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate host SSH key pair: %v\nOutput: %s", err, output)
	}
	if err := os.Chmod(hostSSHKeyPath, 0o600); err != nil {
		t.Fatalf("Failed to set permissions on host SSH private key: %v", err)
	}

	// Generate SSH key pair for the VM (for git clone from VM to Git server)
	vmSSHKeyPath := filepath.Join(tempDir, "id_rsa_vm")
	cmd = exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", vmSSHKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate VM SSH key pair: %v\nOutput: %s", err, output)
	}
	if err := os.Chmod(vmSSHKeyPath, 0o600); err != nil {
		t.Fatalf("Failed to set permissions on VM SSH private key: %v", err)
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
	targetUser, err := cloudinit.NewUser("ubuntu", hostSSHKeyPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	sshKeys, err := cloudinit.NewRSAKeyFromPrivateKeyFile(vmSSHKeyPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	userData := cloudinit.UserData{
		Hostname:      "",
		Users:         []cloudinit.User{targetUser},
		SSHKeys:       sshKeys,
		SSHDeleteKeys: false,
	}

	// -- create new vm config
	cfg := vmm.NewVMConfig(
		vmName,
		imageCachePath,
		userData,
	)

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

	// Get and log serial console output
	consoleOutput, err := vmm.GetConsoleOutput(vmConn, dom)
	if err != nil {
		t.Logf("Failed to get console output: %v", err)
	} else {
		t.Logf("\n--- VM Console Output ---\n%s\n-------------------------", consoleOutput)
	}

	// Wait for SSH to be ready
	sshClient, err := ssh.NewClient(
		ipAddress,
		"ubuntu",
		hostSSHKeyPath,
		"22",
	) // Use hostSSHKeyPath for initial SSH connection to VM
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}
	if err := sshClient.AwaitServer(30 * time.Second); err != nil {
		t.Fatalf("SSH server in VM %s did not become ready in time: %v", vmName, err)
	}

	// -- Setup Git server
	repoName := "edge-cd"
	localIPAddr, err := getLocalIPAddr()
	if err != nil {
		t.Fatalf("Failed to get local ip addr: %v", err)
	}
	repoSrcPath, err := exec.Command("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		t.Fatalf("Error getting current repo path: %v", repoSrcPath)
	}
	gitServer := gitserver.Server{
		ServerAddr:  localIPAddr.String(),
		SSHPort:     0,
		HostKeyPath: hostSSHKeyPath,
		AuthorizedKeys: []string{
			userData.SSHKeys.RSAPublic,             // pub key of the vm
			userData.Users[0].SSHAuthorizedKeys[0], // pub key of the host
		},
		BaseDir: t.TempDir(),
		Repo: []gitserver.Repo{
			{
				Name: repoName,
				Source: gitserver.Source{
					Type:      gitserver.LocalSource,
					LocalPath: "",
				},
			},
		},
	}
	if err := gitServer.Run(); err != nil {
		t.Fatalf("Failed to setup Git server: %v", err)
	}
	t.Cleanup(func() {
		gitServer.Teardown()
	})
	gitRepoUrl := gitServer.GetRepoUrl(repoName)
	t.Logf("Git Repository URL: %s", gitRepoUrl)

	// Build the edgectl binary
	binaryPath := buildEdgectlHelper(t)
	configPath := "./test/edgectl/e2e/config"
	configSpec := "config.yaml"

	cmd = exec.Command(
		binaryPath,
		"bootstrap",
		"--target-addr",
		ipAddress,
		"--target-user",
		targetUser.Name,
		"--ssh-private-key",
		hostSSHKeyPath, // Use host SSH key for edgectl's connection to VM
		"--config-repo",
		gitRepoUrl, // Use the Git server for config repo
		"--config-path",
		configPath,
		"--config-spec",
		configSpec,
		"--edge-cd-repo",
		gitRepoUrl, // Use the Git server for edge-cd repo, which also serves package manager configs
		"--packages",
		"git,curl,openssh-client", // Ensure git and ssh-client are installed on VM
		"--service-manager",
		"systemd",
		"--package-manager",
		"apt",
	)

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run `edgectl bootstrap`: %v", err)
	}

	// Check packages
	_, _, err = sshClient.Run("dpkg -s git && dpkg -s curl && dpkg -s openssh-client")
	if err != nil {
		t.Errorf("git, curl, or openssh-client not installed: %v", err)
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

func getLocalIPAddr() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// Get the local address used for that connection
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
