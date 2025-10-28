package e2e

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
	"sigs.k8s.io/yaml"
)

var (
	errInvalidTestEnvironment      = errors.New("invalid test environment: nil or empty ID")
	errTargetVMIPNotSet            = errors.New("target VM IP address not set")
	errGitServerVMIPNotSet         = errors.New("git server VM IP address not set")
	errEdgectlBinaryRequired       = errors.New("EdgectlBinaryPath is required")
	errCreateSSHClientForExecutor  = errors.New("failed to create SSH client")
	errEdgeCDRepoURLNotFound       = errors.New("edge-cd repository URL not found in test environment")
	errUserConfigRepoURLNotFound   = errors.New("user-config repository URL not found in test environment")
	errBootstrapCommand            = errors.New("bootstrap command failed")
	errBootstrapVerification       = errors.New("bootstrap verification failed")
	errVerificationFailed          = errors.New("verification failed")
	errCreateTempDirForBuild       = errors.New("failed to create temporary directory")
	errBuildEdgectl                = errors.New("failed to build edgectl binary")
	errRemoveTempDirAfterBuild     = errors.New("error removing temp dir")
	errFetchConfig                 = errors.New("failed to fetch config from target VM")
	errParseConfig                 = errors.New("failed to parse config YAML")
	errFileNotCreatedByService     = errors.New("file not created by edge-cd service within timeout")
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
func ExecuteBootstrapTest(
	ctx execcontext.Context,
	env *TestEnvironment,
	config ExecutorConfig,
) error {
	// Validate inputs
	if env == nil || env.ID == "" {
		return errInvalidTestEnvironment
	}
	if env.TargetVM.IP == "" {
		return errTargetVMIPNotSet
	}
	if env.GitServerVM.IP == "" {
		return errGitServerVMIPNotSet
	}
	if config.EdgectlBinaryPath == "" {
		return errEdgectlBinaryRequired
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
		return flaterrors.Join(err, errCreateSSHClientForExecutor)
	}

	// Get repository URLs from environment
	edgeCDRepoURL := env.GitSSHURLs["edge-cd"]
	userConfigRepoURL := env.GitSSHURLs["user-config"]

	if edgeCDRepoURL == "" {
		return errEdgeCDRepoURLNotFound
	}
	if userConfigRepoURL == "" {
		return errUserConfigRepoURLNotFound
	}

	// Define remote destination paths
	remoteEdgeCDRepoDestPath := "/home/ubuntu/edge-cd"
	remoteUserConfigRepoDestPath := "/home/ubuntu/edge-cd-config"

	injectEnv := "GIT_SSH_COMMAND=ssh -i /home/ubuntu/.ssh/id_ed25519 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

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

	// Create bootstrap log file
	bootstrapLogPath := filepath.Join(env.ArtifactPath, "bootstrap.log")
	bootstrapLogFile, err := os.Create(bootstrapLogPath)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("failed to create bootstrap log file at %s", bootstrapLogPath))
	}
	defer bootstrapLogFile.Close()

	// Store log path in environment and track for cleanup
	env.BootstrapLogPath = bootstrapLogPath
	env.ManagedResources = append(env.ManagedResources, bootstrapLogPath)

	// Show command output to both stderr and log file
	multiWriter := io.MultiWriter(os.Stderr, bootstrapLogFile)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter

	// Run bootstrap command
	if err := cmd.Run(); err != nil {
		return flaterrors.Join(err, errBootstrapCommand)
	}

	// Verify bootstrap results
	verifyErrors := verifyBootstrapResults(
		sshClient,
		remoteEdgeCDRepoDestPath,
		remoteUserConfigRepoDestPath,
		config.ServiceManager,
	)
	if len(verifyErrors) > 0 {
		return flaterrors.Join(fmt.Errorf("errors=%v", verifyErrors), errBootstrapVerification)
	}

	// Update environment status to passed
	env.Status = "passed"

	return nil
}

// EdgeCDConfig represents the structure of edge-cd config.yaml
type EdgeCDConfig struct {
	Files []string `yaml:"files"`
}

// waitForFile polls for a file to exist on the target VM, up to maxWait duration
func waitForFile(ctx execcontext.Context, sshClient *ssh.Client, filePath string, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	checkInterval := 2 * time.Second

	for time.Now().Before(deadline) {
		_, _, err := sshClient.Run(ctx, "[", "-f", filePath, "]")
		if err == nil {
			slog.Info("file created by edge-cd service", "file", filePath)
			return nil
		}

		time.Sleep(checkInterval)
	}

	return flaterrors.Join(
		fmt.Errorf("filePath=%s timeout=%s", filePath, maxWait),
		errFileNotCreatedByService,
	)
}

// verifyBootstrapResults checks that all expected files and services exist after bootstrap
func verifyBootstrapResults(
	sshClient *ssh.Client,
	edgeCDRepoPath, userConfigRepoPath, serviceManager string,
) []error {
	var errors []error

	verifications := []struct {
		name    string
		command []string
	}{
		{
			name:    "git package installed",
			command: []string{"dpkg", "-s", "git"},
		},
		{
			name:    "curl package installed",
			command: []string{"dpkg", "-s", "curl"},
		},
		{
			name:    "openssh-client package installed",
			command: []string{"dpkg", "-s", "openssh-client"},
		},
		{
			name:    "yq installed",
			command: []string{"which", "yq"},
		},
		{
			name:    "edge-cd repository cloned",
			command: []string{"[", "-d", fmt.Sprintf("%s/.git", edgeCDRepoPath), "]"},
		},
		{
			name:    "user-config repository cloned",
			command: []string{"[", "-d", fmt.Sprintf("%s/.git", userConfigRepoPath), "]"},
		},
		{
			name:    "config file placed",
			command: []string{"[", "-f", "/etc/edge-cd/config.yaml", "]"},
		},
	}

	// Service-specific verifications
	switch serviceManager {
	default:
		panic("")
	case "systemd":
		verifications = append(verifications, []struct {
			name    string
			command []string
		}{
			{
				name:    "systemd service file created",
				command: []string{"[", "-f", "/etc/systemd/system/edge-cd.service", "]"},
			},
			{
				name:    "systemd service enabled",
				command: []string{"systemctl", "is-enabled", "edge-cd.service"},
			},
			{
				name:    "systemd service active",
				command: []string{"systemctl", "is-active", "edge-cd.service"},
			},
		}...)
	case "procd":
		verifications = append(verifications, struct {
			name    string
			command []string
		}{
			name:    "procd init.d script created",
			command: []string{"[", "-f", "/etc/init.d/edge-cd", "]"},
		})
	}

	// Create an empty context for verification commands
	verifyCtx := execcontext.New(make(map[string]string), []string{})

	// Run all verifications
	for _, v := range verifications {
		_, _, err := sshClient.Run(verifyCtx, v.command...)
		if err != nil {
			errors = append(errors, flaterrors.Join(err, fmt.Errorf("verification=%s", v.name), errVerificationFailed))
		}
	}

	// Fetch and verify files specified in config.yaml are created by edge-cd service
	slog.Info("fetching config.yaml from target VM to verify edge-cd service file synchronization")
	configContent, stderr, err := sshClient.Run(verifyCtx, "cat", "/etc/edge-cd/config.yaml")
	if err != nil {
		errors = append(errors, flaterrors.Join(
			err,
			fmt.Errorf("stderr=%s", stderr),
			errFetchConfig,
		))
		return errors
	}

	// Parse config to extract files list
	var config EdgeCDConfig
	if err := yaml.Unmarshal([]byte(configContent), &config); err != nil {
		errors = append(errors, flaterrors.Join(err, errParseConfig))
		return errors
	}

	// Wait for each file to be created by the edge-cd service (up to 60 seconds each)
	if len(config.Files) > 0 {
		slog.Info("waiting for edge-cd service to create files", "count", len(config.Files))
		for _, filePath := range config.Files {
			if err := waitForFile(verifyCtx, sshClient, filePath, 60*time.Second); err != nil {
				errors = append(errors, err)
			}
		}
	} else {
		slog.Info("no files specified in config.yaml, skipping file verification")
	}

	return errors
}

// BuildEdgectlBinary builds the edgectl binary and returns its path.
// It creates a temporary directory for the binary.
func BuildEdgectlBinary(edgectlSourceDir string) (string, error) {
	// Create a temporary directory for the binary
	tmpDir, err := os.MkdirTemp("", "edgectl-build-")
	if err != nil {
		return "", flaterrors.Join(err, errCreateTempDirForBuild)
	}

	binaryPath := filepath.Join(tmpDir, "edgectl")
	cmd := exec.Command("go", "build", "-o", binaryPath, edgectlSourceDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if err := os.RemoveAll(tmpDir); err != nil {
			slog.Error("error removing temp dir", "err", err.Error(), "tempDir", tmpDir)
		}
		return "", flaterrors.Join(err, errBuildEdgectl)
	}

	return binaryPath, nil
}
