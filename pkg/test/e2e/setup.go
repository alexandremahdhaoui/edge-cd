package e2e

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/cloudinit"
	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/gitserver"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
)

// SetupConfig contains configuration for test environment setup
type SetupConfig struct {
	// ArtifactDir is the base directory for test artifacts
	ArtifactDir string

	// ImageCacheDir is where VM images are cached
	ImageCacheDir string

	// EdgeCDRepoPath is the path to the local edge-cd repository
	EdgeCDRepoPath string

	// DownloadImages controls whether to download missing VM images
	DownloadImages bool
}

// SetupTestEnvironment creates a complete test environment with VMs, git server, and SSH keys.
// It is the single source of truth for test setup and is used by both the test harness and CLI.
//
// Returns fully populated TestEnvironment ready for ExecuteBootstrapTest().
// Caller is responsible for calling TeardownTestEnvironment() when done.
func SetupTestEnvironment(
	execCtx execcontext.Context,
	config SetupConfig,
) (*TestEnvironment, error) {
	// Validate config
	if config.ArtifactDir == "" {
		return nil, fmt.Errorf("ArtifactDir is required")
	}
	if config.ImageCacheDir == "" {
		return nil, fmt.Errorf("ImageCacheDir is required")
	}
	if config.EdgeCDRepoPath == "" {
		return nil, fmt.Errorf("EdgeCDRepoPath is required")
	}

	// Create artifact directory
	if err := os.MkdirAll(config.ArtifactDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create artifact directory: %w", err)
	}

	// Create in-memory test environment
	manager := NewManager(config.ArtifactDir)
	testEnv, err := manager.CreateEnvironment(execCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create test environment: %w", err)
	}

	// Create the root temp directory with marker file: /tmp/e2e-<test-id>
	// The marker file ensures we only delete managed temp directories
	tempDirRoot := filepath.Join(os.TempDir(), testEnv.ID)
	if _, err := CreateTempDirectory(tempDirRoot); err != nil {
		return nil, fmt.Errorf("failed to create managed temp directory root: %w", err)
	}
	testEnv.TempDirRoot = tempDirRoot

	// Create component-specific subdirectories
	vmmTempDir := filepath.Join(tempDirRoot, "vmm")
	gitServerTempDir := filepath.Join(tempDirRoot, "gitserver")
	artifactsTempDir := filepath.Join(tempDirRoot, "artifacts")

	for _, dir := range []string{vmmTempDir, gitServerTempDir, artifactsTempDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create temp subdirectory %s: %w", dir, err)
		}
	}

	// Create artifact subdirectory for this specific test (using the new structure)
	artifactDir := filepath.Join(config.ArtifactDir, "artifacts", testEnv.ID)
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create artifact subdirectory: %w", err)
	}
	testEnv.ArtifactPath = artifactDir

	// Download VM image if needed
	imageName := "ubuntu-24.04-server-cloudimg-amd64.img"
	imageURL := "https://cloud-images.ubuntu.com/releases/noble/release/" + imageName
	imageCachePath := filepath.Join(config.ImageCacheDir, imageName)

	if _, err := os.Stat(imageCachePath); os.IsNotExist(err) {
		if config.DownloadImages {
			if err := downloadVMImage(imageURL, imageCachePath); err != nil {
				return nil, fmt.Errorf("failed to download VM image: %w", err)
			}
		} else {
			return nil, fmt.Errorf("VM image not found and DownloadImages is false: %s", imageCachePath)
		}
	}

	// Generate SSH key pair for host access to target VM
	hostKeyPath := filepath.Join(artifactDir, "id_rsa_host")

	if err := generateSSHKeyPair(hostKeyPath); err != nil {
		return nil, fmt.Errorf("failed to generate host SSH key: %w", err)
	}
	testEnv.SSHKeys.HostKeyPath = hostKeyPath
	testEnv.SSHKeys.HostKeyPubPath = hostKeyPath + ".pub"

	// Create target VM (pass VMM temp directory)
	targetVM, err := setupTargetVM(execCtx, testEnv, imageCachePath, vmmTempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to setup target VM: %w", err)
	}
	testEnv.TargetVM = *targetVM
	// Track created files from target VM
	testEnv.ManagedResources = append(testEnv.ManagedResources, targetVM.CreatedFiles...)

	// Create git server VM (pass git server temp directory)
	gitServerVM, err := setupGitServer(
		execCtx,
		testEnv,
		imageCachePath,
		config.EdgeCDRepoPath,
		gitServerTempDir,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to setup git server: %w", err)
	}
	testEnv.GitServerVM = *gitServerVM.VMMetadata
	testEnv.GitSSHURLs = gitServerVM.GitSSHURLs
	// Track created files from git server VM
	testEnv.ManagedResources = append(
		testEnv.ManagedResources,
		gitServerVM.VMMetadata.CreatedFiles...)

	// Backward compat: Track temp root directory (TempDirRoot is the primary tracker now)
	// testEnv.TempDirs is deprecated but kept for backward compatibility
	testEnv.TempDirs = []string{tempDirRoot}

	// Update status
	testEnv.Status = "created"
	if err := manager.UpdateEnvironment(execCtx, testEnv); err != nil {
		return nil, fmt.Errorf("failed to update test environment: %w", err)
	}

	return testEnv, nil
}

// setupTargetVM creates and configures the target VM for testing
func setupTargetVM(
	execCtx execcontext.Context,
	env *TestEnvironment,
	imageCachePath string,
	vmmTempDir string,
) (*vmm.VMMetadata, error) {
	// Read SSH public keys
	hostPubKey, err := os.ReadFile(env.SSHKeys.HostKeyPubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read host public key: %w", err)
	}

	// Create ubuntu user with host's public key in authorized_keys
	ubuntuUser := cloudinit.NewUserWithAuthorizedKeys("ubuntu", []string{string(hostPubKey)})

	// Setup cloud-init user data
	userData := cloudinit.UserData{
		Hostname: fmt.Sprintf("test-target-%s", env.ID),
		Users:    []cloudinit.User{ubuntuUser},
		RunCommands: []string{
			"KEY_PATH='/home/ubuntu/.ssh/id_ed25519'",
			"USER_HOME='/home/ubuntu'",
			"mkdir -p ${USER_HOME}/.ssh",
			"chmod 700 ${USER_HOME}/.ssh",
			"/usr/bin/ssh-keygen -t ed25519 -N \"\" -f ${KEY_PATH} -q",
			"chown ubuntu:ubuntu -R ${USER_HOME}",
			"chmod 600 ${KEY_PATH}",
			"systemctl restart sshd",
		},
	}

	// Create VM config
	vmConfig := vmm.NewVMConfig(
		fmt.Sprintf("test-target-%s", env.ID),
		imageCachePath,
		userData,
	)
	// Set temp directory for VM artifacts
	vmConfig.TempDir = vmmTempDir

	// Create VMM with base directory option and provision VM
	vmManager, err := vmm.NewVMM(vmm.WithBaseDir(vmmTempDir))
	if err != nil {
		return nil, fmt.Errorf("failed to create VMM: %w", err)
	}
	defer vmManager.Close()

	metadata, err := vmManager.CreateVM(vmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create target VM: %w", err)
	}

	if metadata.IP == "" {
		return nil, fmt.Errorf("target VM created but no IP address available")
	}

	// Wait for SSH to become available
	sshClient, err := ssh.NewClient(
		metadata.IP,
		"ubuntu",
		env.SSHKeys.HostKeyPath,
		"22",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := sshClient.AwaitServer(60 * time.Second); err != nil {
		return nil, fmt.Errorf("target VM SSH server did not become ready: %w", err)
	}

	return metadata, nil
}

// FetchTargetVMPublicKey fetches the public SSH key from the target VM that it will actually use
// This is created by cloud-init and is the key the target VM will use for outbound connections
func FetchTargetVMPublicKey(
	execCtx execcontext.Context,
	metadata *vmm.VMMetadata,
	hostKeyPath string,
) (string, error) {
	// Create SSH client to target VM using host key
	sshClient, err := ssh.NewClient(
		metadata.IP,
		"ubuntu",
		hostKeyPath,
		"22",
	)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH client: %w", err)
	}

	// Fetch the default public key that cloud-init created
	publicKey, stderr, err := sshClient.Run(execCtx, "cat", "${HOME}/.ssh/id_ed25519.pub")
	if err != nil {
		return "", fmt.Errorf("failed to fetch target VM public key: %w\nstderr: %s", err, stderr)
	}

	slog.Info("successfully fetched public key", "publicKey", publicKey, "fromIp", metadata.IP)

	// Trim whitespace to ensure proper formatting in authorized_keys
	return publicKey, nil
}

// setupGitServer creates and configures the git server VM
// Returns the git server status
func setupGitServer(
	execCtx execcontext.Context,
	env *TestEnvironment,
	imageCachePath, edgeCDRepoPath string,
	gitServerTempDir string,
) (*gitserver.Status, error) {
	// Use provided temp directory for git server
	repos := []gitserver.Repo{
		{
			Name: "edge-cd",
			Source: gitserver.Source{
				Type:      gitserver.LocalSource,
				LocalPath: edgeCDRepoPath,
			},
		},
		{
			Name: "user-config",
			Source: gitserver.Source{
				Type:      gitserver.LocalSource,
				LocalPath: edgeCDRepoPath,
			},
		},
	}

	server := gitserver.NewServer(gitServerTempDir, imageCachePath, repos)

	// Configure authorized keys
	// Get public key from host
	hostPubKey, err := os.ReadFile(env.SSHKeys.HostKeyPubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read host public key: %w", err)
	}

	// Fetch target VM's actual public key (created by cloud-init)
	targetPubKey, err := FetchTargetVMPublicKey(execCtx, &env.TargetVM, env.SSHKeys.HostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target VM public key: %w", err)
	}

	server.AuthorizedKeys = []string{
		targetPubKey,                          // Target VM uses this to clone from git server
		strings.TrimSpace(string(hostPubKey)), // Host uses this to SSH to git server
	}

	// Run git server (this creates the VM and sets up repositories)
	if err := server.Run(execCtx); err != nil {
		return nil, fmt.Errorf("failed to run git server: %w", err)
	}

	status := server.Status()
	if status == nil {
		return nil, fmt.Errorf("git server status is nil after successful Run()")
	}

	return status, nil
}

// generateSSHKeyPair generates an RSA SSH key pair
func generateSSHKeyPair(keyPath string) error {
	cmd := exec.Command(
		"ssh-keygen",
		"-t", "rsa",
		"-b", "2048",
		"-f", keyPath,
		"-N", "",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssh-keygen failed: %w\nOutput: %s", err, output)
	}

	// Ensure proper permissions on private key
	if err := os.Chmod(keyPath, 0o600); err != nil {
		return fmt.Errorf("failed to set SSH key permissions: %w", err)
	}

	return nil
}

// downloadVMImage downloads a VM image using wget
func downloadVMImage(imageURL, destPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create image cache directory: %w", err)
	}

	cmd := exec.Command(
		"wget",
		"--progress=dot",
		"-e", "dotbytes=3M",
		"-O", destPath,
		imageURL,
	)

	// Show progress on stderr
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Clean up partial file
		os.Remove(destPath)
		return fmt.Errorf("failed to download VM image: %w", err)
	}

	return nil
}
