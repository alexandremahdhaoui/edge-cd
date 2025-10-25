package gitserver_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/gitserver"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// downloadVMImage downloads the Ubuntu cloud image if it doesn't exist
func downloadVMImage(t *testing.T) string {
	cacheDir := filepath.Join(os.TempDir(), "edgectl")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("failed to create cache directory for vm image %q", cacheDir)
	}

	imageName := "ubuntu-24.04-server-cloudimg-amd64.img"
	imageURL := "https://cloud-images.ubuntu.com/releases/noble/release/" + imageName
	imageCachePath := filepath.Join(cacheDir, imageName)

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

	return imageCachePath
}

// generateClientSSHKey generates an SSH key pair for the test client
func generateClientSSHKey(t *testing.T, keyPath string) {
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", keyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate host SSH key pair for client: %v\nOutput: %s", err, output)
	}
	if err := os.Chmod(keyPath, 0o600); err != nil {
		t.Fatalf("Failed to set permissions on client SSH private key: %v", err)
	}
}

// createLocalGitRepo creates a local Git repository with test files
func createLocalGitRepo(t *testing.T, repoPath string, files map[string]string) {
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("Failed to create local repo directory: %v", err)
	}

	// Create test files
	for fileName, content := range files {
		filePath := filepath.Join(repoPath, fileName)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", fileName, err)
		}
	}

	// Initialize the local repo as a Git repository
	initCmd := exec.Command("git", "init")
	initCmd.Dir = repoPath
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize local repo: %v\nOutput: %s", err, output)
	}

	// Configure git user for the test
	configCmds := [][]string{
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}
	for _, args := range configCmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to configure git: %v\nOutput: %s", err, output)
		}
	}

	// Add and commit all files
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = repoPath
	if output, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to git add: %v\nOutput: %s", err, output)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = repoPath
	if output, err := commitCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to git commit: %v\nOutput: %s", err, output)
	}
}

func TestGitServerLifecycle(t *testing.T) {
	// Skip test if libvirt is not available or if running in CI without KVM
	if os.Getenv("CI") == "true" && os.Getenv("LIBVIRT_TEST") != "true" {
		t.Skip("Skipping libvirt Git server lifecycle test in CI without LIBVIRT_TEST=true")
	}

	// Create a temporary directory for the Git server's base directory and VM artifacts
	tempDir := t.TempDir()

	// Download image if not exists
	imageCachePath := downloadVMImage(t)

	server := gitserver.NewServer(tempDir, imageCachePath, []gitserver.Repo{})

	// Generate a dummy host SSH key for the test client to connect to the Git server VM
	clientKeyPath := filepath.Join(tempDir, "id_rsa_client")
	generateClientSSHKey(t, clientKeyPath)

	// Read the public key and add it to the server's authorized keys
	publicKeyBytes, err := os.ReadFile(clientKeyPath + ".pub")
	if err != nil {
		t.Fatalf("Failed to read client public key: %v", err)
	}
	server.AuthorizedKeys = append(server.AuthorizedKeys, string(publicKeyBytes))

	t.Log("Running Git server VM...")
	if err := server.Run(context.Background()); err != nil {
		t.Fatalf("Failed to run Git server VM: %v", err)
	}
	t.Logf("Git server VM started successfully with IP: %s", server.GetVMIPAddress())

	// Verify the server is running by attempting to connect via SSH
	sshClient, err := ssh.NewClient(
		server.GetVMIPAddress(),
		"git",
		clientKeyPath,
		fmt.Sprintf("%d", server.SSHPort),
	)
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}
	if err := sshClient.AwaitServer(60 * time.Second); err != nil { // Increased timeout for VM startup
		t.Fatalf("Git server VM did not become ready in time: %v", err)
	}
	t.Log("SSH connection to Git server VM successful.")

	// Run a simple command to verify SSH connectivity
	stdout, stderr, err := sshClient.Run("echo hello from client")
	if err != nil || strings.TrimSpace(stdout) != "hello from client" {
		t.Fatalf(
			"Failed to run basic command on VM via SSH: %v\nStdout: %s\nStderr: %s",
			err,
			stdout,
			stderr,
		)
	}
	t.Log("Basic SSH command executed successfully.")

	t.Log("Tearing down Git server VM...")
	if err := server.Teardown(); err != nil {
		t.Fatalf("Failed to teardown Git server VM: %v", err)
	}
	t.Log("Git server VM torn down successfully.")
}

func TestGitServerWithRepo(t *testing.T) {
	// Skip test if libvirt is not available or if running in CI without KVM
	if os.Getenv("CI") == "true" && os.Getenv("LIBVIRT_TEST") != "true" {
		t.Skip("Skipping libvirt Git server repo test in CI without LIBVIRT_TEST=true")
	}

	// Create a temporary directory for the Git server's base directory and VM artifacts
	tempDir := t.TempDir()

	// Download image if not exists
	imageCachePath := downloadVMImage(t)

	// Create a local test repository with a test file
	localRepoPath := filepath.Join(tempDir, "test-repo")
	createLocalGitRepo(t, localRepoPath, map[string]string{
		"hello.txt": "Hello from test repo!",
	})

	// Create Git server with the local repo
	repoName := "test-repo.git"
	repos := []gitserver.Repo{
		{
			Name: repoName,
			Source: gitserver.Source{
				Type:      gitserver.LocalSource,
				LocalPath: localRepoPath,
			},
		},
	}

	server := gitserver.NewServer(tempDir, imageCachePath, repos)

	// Generate a dummy host SSH key for the test client to connect to the Git server VM
	clientKeyPath := filepath.Join(tempDir, "id_rsa_client")
	generateClientSSHKey(t, clientKeyPath)

	// Read the public key and add it to the server's authorized keys
	publicKeyBytes, err := os.ReadFile(clientKeyPath + ".pub")
	if err != nil {
		t.Fatalf("Failed to read client public key: %v", err)
	}
	server.AuthorizedKeys = append(server.AuthorizedKeys, string(publicKeyBytes))

	t.Log("Running Git server VM with repo...")
	if err := server.Run(context.Background()); err != nil {
		t.Fatalf("Failed to run Git server VM: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Teardown(); err != nil {
			t.Logf("Failed to teardown Git server VM: %v", err)
		}
	})

	t.Logf("Git server VM started successfully with IP: %s", server.GetVMIPAddress())

	// Clone the repository from the Git server
	cloneDir := filepath.Join(tempDir, "cloned-repo")
	repoURL := server.GetRepoUrl(repoName)
	t.Logf("Cloning from: %s", repoURL)

	cloneCmd := exec.Command("git", "clone", repoURL, cloneDir)
	cloneCmd.Env = append(
		os.Environ(),
		fmt.Sprintf(
			"GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -p %d",
			clientKeyPath,
			server.SSHPort,
		),
	)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone repository from Git server: %v\nOutput: %s", err, output)
	}
	t.Logf("Repository cloned successfully to %s", cloneDir)

	// Verify the cloned file exists and has the correct content
	clonedFilePath := filepath.Join(cloneDir, "hello.txt")
	content, err := os.ReadFile(clonedFilePath)
	if err != nil {
		t.Fatalf("Failed to read cloned file: %v", err)
	}

	expectedContent := "Hello from test repo!"
	if string(content) != expectedContent {
		t.Errorf(
			"Cloned file content mismatch: expected %q, got %q",
			expectedContent,
			string(content),
		)
	}

	t.Log("Repository cloning and file verification successful!")
}
