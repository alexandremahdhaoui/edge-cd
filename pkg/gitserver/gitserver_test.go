package gitserver_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/gitserver"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

func TestGitServerLifecycle(t *testing.T) {
	// Skip test if libvirt is not available or if running in CI without KVM
	if os.Getenv("CI") == "true" && os.Getenv("LIBVIRT_TEST") != "true" {
		t.Skip("Skipping libvirt Git server lifecycle test in CI without LIBVIRT_TEST=true")
	}

	// Create a temporary directory for the Git server's base directory and VM artifacts
	tempDir := t.TempDir()

	// Download image if not exists
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

	server := gitserver.NewServer(tempDir, imageCachePath, []gitserver.Repo{})

	// Generate a dummy host SSH key for the test client to connect to the Git server VM
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", server.HostKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate host SSH key pair for client: %v\nOutput: %s", err, output)
	}
	if err := os.Chmod(server.HostKeyPath, 0o600); err != nil {
		t.Fatalf("Failed to set permissions on client SSH private key: %v", err)
	}

	t.Log("Running Git server VM...")
	if err := server.Run(); err != nil {
		t.Fatalf("Failed to run Git server VM: %v", err)
	}
	t.Logf("Git server VM started successfully with IP: %s", server.GetVMIPAddress())

	// Verify the server is running by attempting to connect via SSH
	sshClient, err := ssh.NewClient(server.GetVMIPAddress(), "git", server.HostKeyPath, fmt.Sprintf("%d", server.SSHPort))
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
		t.Fatalf("Failed to run basic command on VM via SSH: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	t.Log("Basic SSH command executed successfully.")

	t.Log("Tearing down Git server VM...")
	if err := server.Teardown(); err != nil {
		t.Fatalf("Failed to teardown Git server VM: %v", err)
	}
	t.Log("Git server VM torn down successfully.")
}

func TestGitServerRepoInitAndPush(t *testing.T) {
	// Create a temporary directory for the Git server's base directory
	baseDir, err := ioutil.TempDir("", "gitserver-repo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// Generate a dummy host SSH key for the Git server's SSH client
	hostKeyPath := filepath.Join(baseDir, "id_rsa_host")
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", hostKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate host SSH key pair: %v\nOutput: %s", err, output)
	}
	if err := os.Chmod(hostKeyPath, 0o600); err != nil {
		t.Fatalf("Failed to set permissions on host SSH private key: %v", err)
	}

	// Generate a dummy authorized key for the Git server
	authorizedKeyPath := filepath.Join(baseDir, "id_rsa_authorized.pub")
	cmd = exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", filepath.Join(baseDir, "id_rsa_authorized"), "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate authorized SSH key pair: %v\nOutput: %s", err, output)
	}
	authorizedKeyBytes, err := ioutil.ReadFile(authorizedKeyPath)
	if err != nil {
		t.Fatalf("Failed to read authorized public key: %v", err)
	}
	authorizedKey := string(authorizedKeyBytes)

	// Create a dummy local repository to push
	localRepoPath := filepath.Join(baseDir, "local-repo")
	if err := os.MkdirAll(localRepoPath, 0o755); err != nil {
		t.Fatalf("Failed to create local repo dir: %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(localRepoPath, "testfile.txt"), []byte("Hello, Git!"), 0o644); err != nil {
		t.Fatalf("Failed to write testfile.txt: %v", err)
	}

	server := gitserver.Server{
		ServerAddr:     "localhost",
		SSHPort:        2223, // Use a different non-standard port
		HostKeyPath:    hostKeyPath,
		AuthorizedKeys: []string{authorizedKey},
		BaseDir:        baseDir,
		Repo: []gitserver.Repo{
			{
				Name: "test-repo.git",
				Source: gitserver.Source{
					Type:      gitserver.LocalSource,
					LocalPath: localRepoPath,
				},
			},
		},
	}

	t.Log("Running Git server with repo...")
	if err := server.Run(); err != nil {
		t.Fatalf("Failed to run Git server with repo: %v", err)
	}
	t.Log("Git server with repo started successfully.")

	defer func() {
		if err := server.Teardown(); err != nil {
			t.Errorf("Failed to teardown Git server: %v", err)
		}
	}()

	// Clone the repository from the Git server
	cloneDest := filepath.Join(baseDir, "cloned-repo")
	cloneCmd := exec.Command("git", "clone", server.GetRepoUrl("test-repo.git"), cloneDest)
	cloneCmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", hostKeyPath))
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone repository from Git server: %v\nOutput: %s", err, output)
	}
	t.Logf("Repository cloned to %s", cloneDest)

	// Verify content
	content, err := ioutil.ReadFile(filepath.Join(cloneDest, "testfile.txt"))
	if err != nil {
		t.Fatalf("Failed to read cloned file: %v", err)
	}
	if string(content) != "Hello, Git!" {
		t.Errorf("Expected 'Hello, Git!', got '%s'", string(content))
	}
	t.Log("Cloned repository content verified.")
}
