package ssh_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
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

func TestE2ERealSSHClient(t *testing.T) {
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

	ctx := execcontext.New(make(map[string]string), []string{})
	stdout, stderr, err := client.Run(ctx, "echo hello from real client")
	if err != nil {
		t.Fatalf("Failed to run command via real SSH client: %v\nStderr: %s", err, stderr)
	}

	expectedStdout := "hello from real client\n"
	if stdout != expectedStdout {
		t.Errorf("Unexpected stdout. Got: %q, Expected: %q", stdout, expectedStdout)
	}
	if stderr != "" {
		t.Errorf("Unexpected stderr. Got: %q", stderr)
	}
}
