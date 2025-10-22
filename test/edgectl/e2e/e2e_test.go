package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

func TestE2EFramework(t *testing.T) {
	preTestCleanup(t) // Call pre-test cleanup

	// Build the edgectl binary
	binaryPath := buildEdgectlHelper(t)

	// Get host's SSH public key and private key path
	privateKeyPath, sshPublicKey := getOrCreateSSHKeyPair(t)

	// Build the e2e target image (this is already done in TestDockerLifecycle, but good to ensure here)
	buildCmd := exec.Command(
		"docker",
		"build",
		"-t",
		e2eTargetImage,
		"--build-arg",
		"SSH_PUBLIC_KEY="+sshPublicKey,
		"./testdata",
	)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Docker image %s: %v\nOutput: %s", e2eTargetImage, err, output)
	}

	// Start the Docker container
	containerID, err := startContainerHelper(t)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Ensure cleanup happens
	t.Cleanup(func() {
		if containerID != "" {
			stopContainerHelper(t, containerID)
			cleanupContainerHelper(t, containerID)
		}
	})

	t.Logf("Running edgectl bootstrap --help inside container %s", containerID)

	// Copy the edgectl binary to the container
	scmd := exec.Command(
		"scp",
		"-i",
		privateKeyPath,
		"-o",
		"StrictHostKeyChecking=no",
		"-o",
		"UserKnownHostsFile=/dev/null",
		"-P",
		sshPort,
		binaryPath,
		"root@localhost:/usr/local/bin/edgectl",
	)
	if output, err := scmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to copy edgectl binary to container: %v\nOutput: %s", err, output)
	}

	// Run edgectl bootstrap --help inside the container
	sshCmd := exec.Command(
		"ssh",
		"-i",
		privateKeyPath,
		"-o",
		"StrictHostKeyChecking=no",
		"-o",
		"UserKnownHostsFile=/dev/null",
		"-p",
		sshPort,
		"root@localhost",
		"/usr/local/bin/edgectl bootstrap --help",
	)
	output, err := sshCmd.CombinedOutput()
	// Assert the command exits with code 0 (no error)
	if err != nil {
		t.Fatalf("edgectl bootstrap --help command failed: %v\nOutput: %s", err, output)
	}

	// Assert the help message is printed (check for a known string)
	expectedOutputPart := "Usage of /usr/local/bin/edgectl bootstrap:"
	if !strings.Contains(string(output), expectedOutputPart) {
		t.Errorf("Expected output to contain %q, but got:\n%s", expectedOutputPart, string(output))
	}
	t.Log("edgectl bootstrap --help command verified.")
}

func TestE2EProvisionPackages(t *testing.T) {
	preTestCleanup(t)
	// Build the edgectl binary (not strictly needed for this test, but good practice)
	_ = buildEdgectlHelper(t)

	// Get host's SSH public key and private key path
	privateKeyPath, sshPublicKey := getOrCreateSSHKeyPair(t)

	// Build the e2e target image with SSH public key
	buildCmd := exec.Command(
		"docker",
		"build",
		"-t",
		e2eTargetImage,
		"--build-arg",
		"SSH_PUBLIC_KEY="+sshPublicKey,
		"./testdata",
	)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Docker image %s: %v\nOutput: %s", e2eTargetImage, err, output)
	}

	// Start the Docker container
	containerID, err := startContainerHelper(t)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Ensure cleanup happens
	t.Cleanup(func() {
		if containerID != "" {
			stopContainerHelper(t, containerID)
			cleanupContainerHelper(t, containerID)
		}
	})

	// Create SSH client
	sshClient, err := ssh.NewClient("localhost", "root", privateKeyPath, "", sshPort)
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}

	packagesToInstall := []string{"git"}

	// Run package installation for the first time
	t.Logf("Installing packages: %v (first run)", packagesToInstall)
	if err := provision.InstallPackages(sshClient, packagesToInstall); err != nil {
		t.Fatalf("First package installation failed: %v", err)
	}

	// Verify package is installed
	checkCmd := fmt.Sprintf("dpkg -s %s &> /dev/null", packagesToInstall[0])
	_, _, err = sshClient.Run(checkCmd)
	if err != nil {
		t.Errorf("Package %s not found after first installation: %v", packagesToInstall[0], err)
	}

	// Run package installation again (test idempotency)
	t.Logf("Installing packages: %v (second run - idempotency test)", packagesToInstall)
	if err := provision.InstallPackages(sshClient, packagesToInstall); err != nil {
		t.Fatalf("Second package installation (idempotency test) failed: %v", err)
	}

	// Verify package is still installed after idempotent run
	_, _, err = sshClient.Run(checkCmd)
	if err != nil {
		t.Errorf("Package %s not found after second installation: %v", packagesToInstall[0], err)
	}

	t.Log("E2E package provisioning test passed.")
}

func TestE2ECloneEdgeCDRepo(t *testing.T) {
	preTestCleanup(t)

	// Build the edgectl binary (not strictly needed for this test, but good practice)
	_ = buildEdgectlHelper(t)

	// Get host's SSH public key and private key path
	privateKeyPath, sshPublicKey := getOrCreateSSHKeyPair(t)

	// Build the e2e target image with SSH public key
	buildCmd := exec.Command(
		"docker",
		"build",
		"-t",
		e2eTargetImage,
		"--build-arg",
		"SSH_PUBLIC_KEY="+sshPublicKey,
		"./testdata",
	)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Docker image %s: %v\nOutput: %s", e2eTargetImage, err, output)
	}

	// Start the Docker container
	containerID, err := startContainerHelper(t)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Ensure cleanup happens
	t.Cleanup(func() {
		if containerID != "" {
			stopContainerHelper(t, containerID)
			cleanupContainerHelper(t, containerID)
		}
	})

	// Create SSH client
	sshClient, err := ssh.NewClient("localhost", "root", privateKeyPath, "", sshPort)
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}

	// Install git package (required for cloning)
	packagesToInstall := []string{"git"}
	t.Logf("Installing packages: %v (for cloning)", packagesToInstall)
	if err := provision.InstallPackages(sshClient, packagesToInstall); err != nil {
		t.Fatalf("Failed to install git for cloning: %v", err)
	}

	edgeCDRepoURL := "https://github.com/alexandremahdhaoui/edge-cd.git"
	edgeCDDestPath := "/opt/edge-cd"

	// Run repository cloning for the first time
	t.Logf("Cloning edge-cd repository (first run)")
	if err := provision.CloneOrPullRepo(sshClient, edgeCDRepoURL, edgeCDDestPath); err != nil {
		t.Fatalf("First edge-cd repository cloning failed: %v", err)
	}

	// Verify repository exists
	checkCmd := fmt.Sprintf("[ -d %s/.git ]", edgeCDDestPath)
	_, _, err = sshClient.Run(checkCmd)
	if err != nil {
		t.Errorf("Edge-cd repository not found after first cloning: %v", err)
	}

	// Run repository cloning again (test idempotency)
	t.Logf("Cloning edge-cd repository (second run - idempotency test)")
	if err := provision.CloneOrPullRepo(sshClient, edgeCDRepoURL, edgeCDDestPath); err != nil {
		t.Fatalf("Second edge-cd repository cloning (idempotency test) failed: %v", err)
	}

	// Verify repository still exists after idempotent run
	_, _, err = sshClient.Run(checkCmd)
	if err != nil {
		t.Errorf("Edge-cd repository not found after second cloning: %v", err)
	}

	t.Log("E2E edge-cd repository cloning test passed.")
}

func TestE2ECloneUserConfigRepo(t *testing.T) {
	preTestCleanup(t)

	// Build the edgectl binary (not strictly needed for this test, but good practice)
	_ = buildEdgectlHelper(t)

	// Get host's SSH public key and private key path
	privateKeyPath, sshPublicKey := getOrCreateSSHKeyPair(t)

	// Build the e2e target image with SSH public key
	buildCmd := exec.Command(
		"docker",
		"build",
		"-t",
		e2eTargetImage,
		"--build-arg",
		"SSH_PUBLIC_KEY="+sshPublicKey,
		"./testdata",
	)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Docker image %s: %v\nOutput: %s", e2eTargetImage, err, output)
	}

	// Start the Docker container
	containerID, err := startContainerHelper(t)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Ensure cleanup happens
	t.Cleanup(func() {
		if containerID != "" {
			stopContainerHelper(t, containerID)
			cleanupContainerHelper(t, containerID)
		}
	})

	// Create SSH client
	sshClient, err := ssh.NewClient("localhost", "root", privateKeyPath, "", sshPort)
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}

	// Install git package (required for cloning)
	packagesToInstall := []string{"git"}
	t.Logf("Installing packages: %v (for cloning)", packagesToInstall)
	if err := provision.InstallPackages(sshClient, packagesToInstall); err != nil {
		t.Fatalf("Failed to install git for cloning: %v", err)
	}

	userConfigRepoURL := "https://github.com/alexandremahdhaoui/edge-cd.git"
	userConfigDestPath := "/opt/user-config"

	// Run repository cloning for the first time
	t.Logf("Cloning user config repository (first run)")
	if err := provision.CloneOrPullRepo(sshClient, userConfigRepoURL, userConfigDestPath); err != nil {
		t.Fatalf("First user config repository cloning failed: %v", err)
	}

	// Verify repository exists
	checkCmd := fmt.Sprintf("[ -d %s/.git ]", userConfigDestPath)
	_, _, err = sshClient.Run(checkCmd)
	if err != nil {
		t.Errorf("User config repository not found after first cloning: %v", err)
	}

	// Run repository cloning again (test idempotency)
	t.Logf("Cloning user config repository (second run - idempotency test)")
	if err := provision.CloneOrPullRepo(sshClient, userConfigRepoURL, userConfigDestPath); err != nil {
		t.Fatalf("Second user config repository cloning (idempotency test) failed: %v", err)
	}

	// Verify repository still exists after idempotent run
	_, _, err = sshClient.Run(checkCmd)
	if err != nil {
		t.Errorf("User config repository not found after second cloning: %v", err)
	}

	t.Log("E2E user config repository cloning test passed.")
}

func TestE2ESetupEdgeCDService(t *testing.T) {
	preTestCleanup(t)

	// Build the edgectl binary (not strictly needed for this test, but good practice)
	_ = buildEdgectlHelper(t)

	// Get host's SSH public key and private key path
	privateKeyPath, sshPublicKey := getOrCreateSSHKeyPair(t)

	// Build the e2e target image with SSH public key
	buildCmd := exec.Command(
		"docker",
		"build",
		"-t",
		e2eTargetImage,
		"--build-arg",
		"SSH_PUBLIC_KEY="+sshPublicKey,
		"./testdata",
	)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Docker image %s: %v\nOutput: %s", e2eTargetImage, err, output)
	}

	// Start the Docker container
	containerID, err := startContainerHelper(t)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Ensure cleanup happens
	t.Cleanup(func() {
		if containerID != "" {
			stopContainerHelper(t, containerID)
			cleanupContainerHelper(t, containerID)
		}
	})

	// Create SSH client
	sshClient, err := ssh.NewClient("localhost", "root", privateKeyPath, "", sshPort)
	if err != nil {
		t.Fatalf("Failed to create SSH client: %v", err)
	}

	// Install git package (required for cloning)
	packagesToInstall := []string{"git"}
	t.Logf("Installing packages: %v (for cloning)", packagesToInstall)
	if err := provision.InstallPackages(sshClient, packagesToInstall); err != nil {
		t.Fatalf("Failed to install git for cloning: %v", err)
	}

	edgeCDRepoURL := "https://github.com/alexandremahdhaoui/edge-cd.git"
	edgeCDRepoPath := "/opt/edge-cd"

	// Clone edge-cd repository
	t.Logf("Cloning edge-cd repository to %s", edgeCDRepoPath)
	if err := provision.CloneOrPullRepo(sshClient, edgeCDRepoURL, edgeCDRepoPath); err != nil {
		t.Fatalf("Failed to clone edge-cd repository: %v", err)
	}

	// Setup edge-cd service (systemd)
	serviceManagerName := "systemd"
	t.Logf("Setting up edge-cd service with %s", serviceManagerName)
	if err := provision.SetupEdgeCDService(sshClient, serviceManagerName, edgeCDRepoPath); err != nil {
		t.Fatalf("Failed to setup edge-cd service: %v", err)
	}

	// Verify service is enabled and active
	checkEnabledCmd := "systemctl is-enabled edge-cd.service"
	checkActiveCmd := "systemctl is-active edge-cd.service"

	_, _, err = sshClient.Run(checkEnabledCmd)
	if err != nil {
		t.Errorf("Edge-cd service is not enabled: %v", err)
	}

	_, _, err = sshClient.Run(checkActiveCmd)
	if err != nil {
		t.Errorf("Edge-cd service is not active: %v", err)
	}

	// Test idempotency
	t.Logf("Setting up edge-cd service again (idempotency test)")
	if err := provision.SetupEdgeCDService(sshClient, serviceManagerName, edgeCDRepoPath); err != nil {
		t.Fatalf("Idempotency test for edge-cd service failed: %v", err)
	}

	// Verify service is still enabled and active after idempotent run
	_, _, err = sshClient.Run(checkEnabledCmd)
	if err != nil {
		t.Errorf("Edge-cd service is not enabled after idempotent run: %v", err)
	}

	_, _, err = sshClient.Run(checkActiveCmd)
	if err != nil {
		t.Errorf("Edge-cd service is not active after idempotent run: %v", err)
	}

	t.Log("E2E edge-cd service setup test passed.")
}