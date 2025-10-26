package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// ExecutorConfig contains configuration for bootstrap test execution
type ExecutorConfig struct {
	// EdgectlBinaryPath is the path to the compiled edgectl binary
	EdgectlBinaryPath string

	// ConfigPath is the path to the config directory (relative to test root)
	ConfigPath string

	// ConfigSpec is the config specification file name
	ConfigSpec string

	// Packages is a comma-separated list of packages to install
	Packages string

	// ServiceManager is the service manager to use (systemd/procd)
	ServiceManager string

	// PackageManager is the package manager to use (apt/opkg)
	PackageManager string
}

// ExecuteBootstrapTest runs the bootstrap test on a pre-configured test environment.
// It does NOT create or destroy VMs - it only runs the bootstrap command and verifies results.
//
// This is the test-logic-only function that is called by both the test harness and CLI.
// Caller must have already called SetupTestEnvironment().
func ExecuteBootstrapTest(ctx context.Context, env *TestEnvironment, config ExecutorConfig) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate inputs
	if env == nil || env.ID == "" {
		return fmt.Errorf("invalid test environment: nil or empty ID")
	}
	if env.TargetVM.IP == "" {
		return fmt.Errorf("target VM IP address not set")
	}
	if env.GitServerVM.IP == "" {
		return fmt.Errorf("git server VM IP address not set")
	}
	if config.EdgectlBinaryPath == "" {
		return fmt.Errorf("EdgectlBinaryPath is required")
	}

	// Set defaults
	if config.ConfigPath == "" {
		config.ConfigPath = "./test/edgectl/e2e/config"
	}
	if config.ConfigSpec == "" {
		config.ConfigSpec = "config.yaml"
	}
	if config.Packages == "" {
		config.Packages = "git,curl,openssh-client"
	}
	if config.ServiceManager == "" {
		config.ServiceManager = "systemd"
	}
	if config.PackageManager == "" {
		config.PackageManager = "apt"
	}

	// Create SSH client to target VM
	sshClient, err := ssh.NewClient(
		env.TargetVM.IP,
		"ubuntu",
		env.SSHKeys.HostKeyPath,
		"22",
	)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	// Get repository URLs from environment
	edgeCDRepoURL := env.GitSSHURLs["edge-cd"]
	userConfigRepoURL := env.GitSSHURLs["user-config"]

	if edgeCDRepoURL == "" {
		return fmt.Errorf("edge-cd repository URL not found in test environment")
	}
	if userConfigRepoURL == "" {
		return fmt.Errorf("user-config repository URL not found in test environment")
	}

	// Define remote destination paths
	remoteEdgeCDRepoDestPath := "/home/ubuntu/edge-cd"
	remoteUserConfigRepoDestPath := "/home/ubuntu/edge-cd-config"

	injectEnv := "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

	// Build bootstrap command
	cmd := exec.Command(
		config.EdgectlBinaryPath,
		"bootstrap",
		"--target-addr", env.TargetVM.IP,
		"--target-user", "ubuntu",
		"--ssh-private-key", env.SSHKeys.HostKeyPath,
		"--config-repo", userConfigRepoURL,
		"--config-path", config.ConfigPath,
		"--config-spec", config.ConfigSpec,
		"--edge-cd-repo", edgeCDRepoURL,
		"--packages", config.Packages,
		"--service-manager", config.ServiceManager,
		"--package-manager", config.PackageManager,
		"--edge-cd-repo-dest", remoteEdgeCDRepoDestPath,
		"--user-config-repo-dest", remoteUserConfigRepoDestPath,
		"--inject-env", injectEnv,
	)

	// Set up environment for git operations
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf(
			"GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
			env.SSHKeys.HostKeyPath,
		),
	)

	// Show command output
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	// Run bootstrap command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bootstrap command failed: %w", err)
	}

	// Verify bootstrap results
	verifyErrors := verifyBootstrapResults(
		sshClient,
		remoteEdgeCDRepoDestPath,
		remoteUserConfigRepoDestPath,
		config.ServiceManager,
	)
	if len(verifyErrors) > 0 {
		return fmt.Errorf("bootstrap verification failed: %v", verifyErrors)
	}

	// Update environment status to passed
	env.Status = "passed"

	return nil
}

// verifyBootstrapResults checks that all expected files and services exist after bootstrap
func verifyBootstrapResults(
	sshClient *ssh.Client,
	edgeCDRepoPath, userConfigRepoPath, serviceManager string,
) []error {
	var errors []error

	verifications := []struct {
		name    string
		command string
	}{
		{
			name:    "git package installed",
			command: "dpkg -s git",
		},
		{
			name:    "curl package installed",
			command: "dpkg -s curl",
		},
		{
			name:    "openssh-client package installed",
			command: "dpkg -s openssh-client",
		},
		{
			name:    "edge-cd repository cloned",
			command: fmt.Sprintf("[ -d %s/.git ]", edgeCDRepoPath),
		},
		{
			name:    "user-config repository cloned",
			command: fmt.Sprintf("[ -d %s/.git ]", userConfigRepoPath),
		},
		{
			name:    "config file placed",
			command: "[ -f /etc/edge-cd/config.yaml ]",
		},
	}

	// Service-specific verifications
	if serviceManager == "systemd" {
		verifications = append(verifications, []struct {
			name    string
			command string
		}{
			{
				name:    "systemd service file created",
				command: "[ -f /etc/systemd/system/edge-cd.service ]",
			},
			{
				name:    "systemd service enabled",
				command: "systemctl is-enabled edge-cd.service",
			},
			{
				name:    "systemd service active",
				command: "systemctl is-active edge-cd.service",
			},
		}...)
	} else if serviceManager == "procd" {
		verifications = append(verifications, struct {
			name    string
			command string
		}{
			name:    "procd init.d script created",
			command: "[ -f /etc/init.d/edge-cd ]",
		})
	}

	// Create an empty context for verification commands
	verifyCtx := execcontext.New(make(map[string]string), []string{})

	// Run all verifications
	for _, v := range verifications {
		_, _, err := sshClient.Run(verifyCtx, v.command)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s verification failed: %w", v.name, err))
		}
	}

	return errors
}

// BuildEdgectlBinary builds the edgectl binary and returns its path.
// It creates a temporary directory for the binary.
func BuildEdgectlBinary(edgectlSourceDir string) (string, error) {
	// Create a temporary directory for the binary
	tmpDir, err := os.MkdirTemp("", "edgectl-build-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	binaryPath := filepath.Join(tmpDir, "edgectl")
	cmd := exec.Command("go", "build", "-o", binaryPath, edgectlSourceDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to build edgectl binary: %w", err)
	}

	return binaryPath, nil
}
