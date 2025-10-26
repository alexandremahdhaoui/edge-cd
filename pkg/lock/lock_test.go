package lock_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/lock"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

const (
	e2eTargetImage = "edgectl-e2e-target"
	sshPort        = "2222" // Port on host to map to container's SSH port
)

// preTestCleanup stops and removes any existing containers that might interfere with the test.
func preTestCleanup(t *testing.T) {
	t.Helper()
	t.Log("Performing pre-test cleanup...")

	// Find containers using port 2222
	cmd := exec.Command("docker", "ps", "-q", "--filter", "publish=2222")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: Failed to list containers using port 2222: %v\nOutput: %s", err, output)
	}
	runningContainers := strings.Fields(strings.TrimSpace(string(output)))

	// Find containers by image name
	cmd = exec.Command("docker", "ps", "-aq", "--filter", "ancestor="+e2eTargetImage)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf(
			"Warning: Failed to list containers by image %s: %v\nOutput: %s",
			e2eTargetImage,
			err,
			output,
		)
	}
	allContainers := strings.Fields(strings.TrimSpace(string(output)))

	// Combine and deduplicate container IDs
	containersToClean := make(map[string]struct{})
	for _, id := range runningContainers {
		containersToClean[id] = struct{}{}
	}
	for _, id := range allContainers {
		containersToClean[id] = struct{}{}
	}

	if len(containersToClean) == 0 {
		t.Log("No existing containers found for cleanup.")
		return
	}

	t.Logf("Found %d containers to clean up: %v", len(containersToClean), containersToClean)

	for id := range containersToClean {
		t.Logf("Stopping container %s...", id)
		stopCmd := exec.Command("docker", "stop", id)
		if output, err := stopCmd.CombinedOutput(); err != nil {
			t.Logf("Warning: Failed to stop container %s: %v\nOutput: %s", id, err, output)
		}

		t.Logf("Removing container %s...", id)
		rmCmd := exec.Command("docker", "rm", id)
		if output, err := rmCmd.CombinedOutput(); err != nil {
			t.Logf("Warning: Failed to remove container %s: %v\nOutput: %s", id, err, output)
		}
	}
	t.Log("Pre-test cleanup complete.")
}

// getOrCreateSSHKeyPair checks for existing SSH keys or generates a temporary one.
// It returns the path to the private key and the public key content.
func getOrCreateSSHKeyPair(t *testing.T) (string, string) {
	t.Helper()

	// Try to find existing default keys
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	privateKeyPaths := []string{
		filepath.Join(homeDir, ".ssh", "id_rsa"),
		filepath.Join(homeDir, ".ssh", "id_ed25519"),
	}

	for _, path := range privateKeyPaths {
		if _, err := os.Stat(path); err == nil {
			// Private key exists, get its public key
			cmd := exec.Command("ssh-keygen", "-y", "-f", path)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf(
					"Warning: Found private key at %s but failed to get public key: %v\nOutput: %s",
					path,
					err,
					output,
				)
				continue
			}
			t.Logf("Using existing SSH key: %s", path)
			return path, strings.TrimSpace(string(output))
		}
	}

	// No existing key found, generate a temporary one
	tmpDir := t.TempDir() // Go 1.15+ provides t.TempDir() for automatic cleanup
	privateKeyPath := filepath.Join(tmpDir, "id_rsa_test")

	t.Logf("Generating temporary SSH key pair at %s", privateKeyPath)
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", privateKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate temporary SSH key pair: %v\nOutput: %s", err, output)
	}

	// Get the public key content
	cmd = exec.Command("ssh-keygen", "-y", "-f", privateKeyPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get public key from temporary private key: %v\nOutput: %s", err, output)
	}
	t.Log("Generated temporary SSH key pair successfully.")
	return privateKeyPath, strings.TrimSpace(string(output))
}

// startContainerHelper starts a Docker container and returns its ID.
func startContainerHelper(t *testing.T, sshPublicKey string) (string, error) {
	t.Helper()

	// Build the e2e target image first
	buildCmd := exec.Command(
		"docker",
		"build",
		"-t",
		e2eTargetImage,
		"--build-arg",
		"SSH_PUBLIC_KEY="+sshPublicKey,
		"../../test/edgectl/e2e/testdata",
	)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf(
			"failed to build Docker image %s: %w\nOutput: %s",
			e2eTargetImage,
			err,
			output,
		)
	}

	// Run the container, mapping SSH port 22 inside to sshPort on the host
	cmd := exec.Command("docker", "run", "-d", "-p", sshPort+":22", e2eTargetImage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to start container: %w\nOutput: %s", err, output)
	}
	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "", fmt.Errorf("container ID is empty")
	}

	// Wait for the SSH server to be ready
	t.Logf("Waiting for SSH server in container %s to be ready...", containerID)
	for i := 0; i < 30; i++ { // Try for 30 seconds
		cmd := exec.Command(
			"ssh",
			"-o",
			"StrictHostKeyChecking=no",
			"-o",
			"UserKnownHostsFile=/dev/null",
			"-p",
			sshPort,
			"root@localhost",
			"echo 'SSH ready'",
		)
		err := cmd.Run()
		if err == nil {
			t.Log("SSH server is ready.")
			return containerID, nil
		}
		time.Sleep(1 * time.Second)
	}
	return "", fmt.Errorf("SSH server in container %s did not become ready in time", containerID)
}

// stopContainerHelper stops a Docker container.
func stopContainerHelper(t *testing.T, containerID string) {
	t.Helper()
	cmd := exec.Command("docker", "stop", containerID)
	var output []byte
	var err error
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Failed to stop container %s: %v\nOutput: %s", containerID, err, output)
	}
}

// cleanupContainerHelper removes a Docker container.
func cleanupContainerHelper(t *testing.T, containerID string) {
	t.Helper()
	cmd := exec.Command("docker", "rm", containerID)
	var output []byte
	var err error
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Failed to remove container %s: %v\nOutput: %s", containerID, err, output)
	}
}

func TestLock(t *testing.T) {
	mockRunner := ssh.NewMockRunner()

	// Create execcontext for tests
	execCtx := execcontext.New(make(map[string]string), []string{})

	// Commands are now formatted with FormatCmd
	mkdirCmd := execcontext.FormatCmd(execCtx, "mkdir", "/tmp/edgectl.lock")
	rmdirCmd := execcontext.FormatCmd(execCtx, "rmdir", "/tmp/edgectl.lock")

	// Test Acquire success
	mockRunner.SetResponse(mkdirCmd, "", "", nil)
	err := lock.Acquire(execCtx, mockRunner)
	if err != nil {
		t.Errorf("Expected no error on Acquire, got %v", err)
	}
	if err := mockRunner.AssertCommandRun(mkdirCmd); err != nil {
		t.Error(err)
	}

	// Test Acquire contention
	mockRunner = ssh.NewMockRunner() // Reset mock
	mockRunner.SetResponse(
		mkdirCmd,
		"",
		"mkdir: cannot create directory '/tmp/edgectl.lock': File exists\n",
		errors.New("exit status 1"),
	)
	err = lock.Acquire(execCtx, mockRunner)
	if !errors.Is(err, lock.ErrLockHeld) {
		t.Errorf("Expected ErrLockHeld on Acquire contention, got %v", err)
	}

	// Test Release success
	mockRunner = ssh.NewMockRunner() // Reset mock
	mockRunner.SetResponse(rmdirCmd, "", "", nil)
	err = lock.Release(execCtx, mockRunner)
	if err != nil {
		t.Errorf("Expected no error on Release, got %v", err)
	}
	if err := mockRunner.AssertCommandRun(rmdirCmd); err != nil {
		t.Error(err)
	}

	// Test Release when lock doesn't exist
	mockRunner = ssh.NewMockRunner() // Reset mock
	mockRunner.SetResponse(
		rmdirCmd,
		"",
		"rmdir: failed to remove '/tmp/edgectl.lock': No such file or directory\n",
		errors.New("exit status 1"),
	)
	err = lock.Release(execCtx, mockRunner)
	if err != nil {
		t.Errorf("Expected no error on Release when lock doesn't exist, got %v", err)
	}
}

func TestE2ELock(t *testing.T) {
	preTestCleanup(t)

	privateKeyPath, sshPublicKey := getOrCreateSSHKeyPair(t)

	containerID, err := startContainerHelper(t, sshPublicKey)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	t.Cleanup(func() {
		stopContainerHelper(t, containerID)
		cleanupContainerHelper(t, containerID)
	})

	client, err := ssh.NewClient("localhost", "root", privateKeyPath, sshPort)
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}

	// Create execcontext for E2E tests
	execCtx := execcontext.New(make(map[string]string), []string{})

	// Test Acquire success
	err = lock.Acquire(execCtx, client)
	if err != nil {
		t.Fatalf("Expected no error on E2E Acquire, got %v", err)
	}

	// Verify lock file exists on remote
	stdout, stderr, err := client.Run(execCtx, "sh", "-c", "test -d /tmp/edgectl.lock && echo 'exists'")
	if err != nil {
		t.Fatalf("Failed to check lock file existence: %v\nStderr: %s", err, stderr)
	}
	if strings.TrimSpace(stdout) != "exists" {
		t.Errorf("Expected lock file to exist, but it doesn't. Stdout: %q", stdout)
	}

	// Test Acquire contention
	err = lock.Acquire(execCtx, client)
	if !errors.Is(err, lock.ErrLockHeld) {
		t.Errorf("Expected ErrLockHeld on E2E Acquire contention, got %v", err)
	}

	// Test Release success
	err = lock.Release(execCtx, client)
	if err != nil {
		t.Fatalf("Expected no error on E2E Release, got %v", err)
	}

	// Verify lock file is gone on remote
	stdout, stderr, err = client.Run(execCtx, "sh", "-c", "test -d /tmp/edgectl.lock && echo 'exists'")
	if err == nil {
		t.Fatalf("Expected error checking for non-existent lock file, but got none. Stdout: %q", stdout)
	}
	// `test -d` returns exit 1 if not found, and typically prints nothing to stderr.
	// So, we assert that stdout is empty and err is not nil.
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("Expected stdout to be empty after lock release, but got: %q", stdout)
	}
	if stderr != "" {
		t.Errorf("Expected stderr to be empty after lock release, but got: %q", stderr)
	}

	// Test Release when lock doesn't exist (should still succeed)
	err = lock.Release(execCtx, client)
	if err != nil {
		t.Fatalf("Expected no error on E2E Release when lock doesn't exist, got %v", err)
	}
}
