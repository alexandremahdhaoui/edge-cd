package e2e

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/edge-cd/pkg/userconfig"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
	"sigs.k8s.io/yaml"
)

var (
	errInvalidTestEnvironment     = errors.New("invalid test environment: nil or empty ID")
	errTargetVMIPNotSet           = errors.New("target VM IP address not set")
	errGitServerVMIPNotSet        = errors.New("git server VM IP address not set")
	errEdgectlBinaryRequired      = errors.New("EdgectlBinaryPath is required")
	errCreateSSHClientForExecutor = errors.New("failed to create SSH client")
	errEdgeCDRepoURLNotFound      = errors.New(
		"edge-cd repository URL not found in test environment",
	)
	errUserConfigRepoURLNotFound = errors.New(
		"user-config repository URL not found in test environment",
	)
	errBootstrapCommand        = errors.New("bootstrap command failed")
	errBootstrapVerification   = errors.New("bootstrap verification failed")
	errVerificationFailed      = errors.New("verification failed")
	errCreateTempDirForBuild   = errors.New("failed to create temporary directory")
	errBuildEdgectl            = errors.New("failed to build edgectl binary")
	errRemoveTempDirAfterBuild = errors.New("error removing temp dir")
	errFetchConfig             = errors.New("failed to fetch config from target VM")
	errParseConfig             = errors.New("failed to parse config YAML")
	errFileNotCreatedByService = errors.New("file not created by edge-cd service within timeout")
	errReconciliationTestFailed = errors.New("reconciliation test scenario failed")
)

// ReconciliationTestScenario defines a test scenario for reconciliation testing
type ReconciliationTestScenario struct {
	// Name is the test scenario name for logging
	Name string

	// FileChanges maps file paths (relative to repo root) to new content
	FileChanges map[string]string

	// ExpectedTargetFiles maps target VM file paths to expected content
	ExpectedTargetFiles map[string]string

	// CommitMessage is the git commit message
	CommitMessage string
}

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
		return flaterrors.Join(
			err,
			fmt.Errorf("failed to create bootstrap log file at %s", bootstrapLogPath),
		)
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

	// Reconciliation Tests: Verify edge-cd can detect and reconcile configuration changes
	slog.Info("Running reconciliation test scenarios")

	// Scenario 1: Modify existing file content
	scenario1 := ReconciliationTestScenario{
		Name: "modify existing file content",
		FileChanges: map[string]string{
			"test/edgectl/e2e/config/files/quagga_bgpd.conf": "# Modified BGP Configuration\n# Test reconciliation\nhostname test-router-modified\n",
		},
		ExpectedTargetFiles: map[string]string{
			"/etc/quagga/bgpd.conf": "# Modified BGP Configuration\n# Test reconciliation\nhostname test-router-modified\n",
		},
		CommitMessage: "test: modify bgpd.conf for reconciliation verification",
	}

	// Scenario 2: Add new file to config
	// This requires updating config.yaml to include the new file
	updatedConfigYAML := `pollingIntervalSecond: 5

config:
  spec: config.yaml
  path: ./test/edgectl/e2e/config
  repo:
    destPath: /home/ubuntu/edge-cd-config
    url: ssh://git@192.168.122.87:22/srv/git/user-config.git

edgeCD:
  repo:
    branch: main
    destinationPath: /usr/local/src/edge-cd
    url: ssh://git@192.168.122.87:22/srv/git/edge-cd.git

extraEnvs:
  - GIT_SSH_COMMAND: "ssh -i /home/ubuntu/.ssh/id_ed25519 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

log:
  format: console

serviceManager:
  name: systemd

packageManager:
  name: apt
  requiredPackages:
    - git
    - curl
  autoUpgrade: false

files:
  - type: file
    srcPath: files/quagga_bgpd.conf
    destPath: /etc/quagga/bgpd.conf
  - type: file
    srcPath: files/quagga_zebra.conf
    destPath: /etc/quagga/zebra.conf
  - type: file
    srcPath: files/new-config-file.txt
    destPath: /etc/test/new-config-file.txt
`

	scenario2 := ReconciliationTestScenario{
		Name: "add new file to config",
		FileChanges: map[string]string{
			"test/edgectl/e2e/config/files/new-config-file.txt": "new file content for reconciliation test\n",
			"test/edgectl/e2e/config/config.yaml":               updatedConfigYAML,
		},
		ExpectedTargetFiles: map[string]string{
			"/etc/test/new-config-file.txt": "new file content for reconciliation test\n",
		},
		CommitMessage: "test: add new configuration file",
	}

	// Scenario 3: Update multiple files simultaneously
	scenario3 := ReconciliationTestScenario{
		Name: "update multiple files",
		FileChanges: map[string]string{
			"test/edgectl/e2e/config/files/quagga_bgpd.conf":  "# BGP Config - Updated Again\n",
			"test/edgectl/e2e/config/files/quagga_zebra.conf": "# Zebra Config - Updated\n",
		},
		ExpectedTargetFiles: map[string]string{
			"/etc/quagga/bgpd.conf":  "# BGP Config - Updated Again\n",
			"/etc/quagga/zebra.conf": "# Zebra Config - Updated\n",
		},
		CommitMessage: "test: update multiple config files",
	}

	// Execute all scenarios sequentially
	for _, scenario := range []ReconciliationTestScenario{scenario1, scenario2, scenario3} {
		if err := executeReconciliationTest(ctx, env, sshClient, scenario); err != nil {
			return flaterrors.Join(
				err,
				fmt.Errorf("scenario=%s", scenario.Name),
				errReconciliationTestFailed,
			)
		}
	}

	slog.Info("All reconciliation test scenarios passed")

	// Update environment status to passed
	env.Status = "passed"

	return nil
}

// waitForFiles polls for a file to exist on the target VM, up to maxWait duration
func waitForFiles(
	ctx execcontext.Context,
	sshClient *ssh.Client,
	files []string,
	maxWait time.Duration,
) error {
	deadline := time.Now().Add(maxWait)
	checkInterval := 2 * time.Second

	fSet := make(map[string]struct{})

	for time.Now().Before(deadline) {
		for _, f := range files {
			_, _, err := sshClient.Run(ctx, "[", "-f", f, "]")
			if err != nil {
				continue
			}
			fSet[f] = struct{}{}
		}
		if len(fSet) == len(files) {
			slog.Info(
				"all exepcted files created by edge-cd service",
				"files",
				fmt.Sprintf("%+v", files),
			)
			return nil
		}
		time.Sleep(checkInterval)
	}

	return flaterrors.Join(
		fmt.Errorf("files=%s timeout=%s", files, maxWait),
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
			errors = append(
				errors,
				flaterrors.Join(err, fmt.Errorf("verification=%s", v.name), errVerificationFailed),
			)
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

	// Parse spec to extract files list
	var spec userconfig.Spec
	if err := yaml.Unmarshal([]byte(configContent), &spec); err != nil {
		errors = append(errors, flaterrors.Join(err, errParseConfig))
		return errors
	}
	expectedFiles := make([]string, 0)
	for _, f := range spec.Files {
		expectedFiles = append(expectedFiles, f.DestPath)
	}

	// Wait for each file to be created by the edge-cd service (up to 60 seconds each)
	if len(spec.Files) > 0 {
		slog.Info("waiting for edge-cd service to create files", "count", len(spec.Files))
		if err := waitForFiles(verifyCtx, sshClient, expectedFiles, 60*time.Second); err != nil {
			errors = append(errors, err)
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

// getEdgeCDServiceLogs retrieves the edge-cd service logs
// edge-cd writes logs to both journald and /var/log/edge-cd.log
func getEdgeCDServiceLogs(ctx execcontext.Context, sshClient *ssh.Client) (string, error) {
	// Try reading from the log file first (more reliable for recent logs)
	stdout, _, err := sshClient.Run(ctx, "cat", "/var/log/edge-cd.log")
	if err == nil && len(strings.TrimSpace(stdout)) > 0 {
		return stdout, nil
	}

	// Fall back to journalctl if log file doesn't exist
	stdout, stderr, err := sshClient.Run(ctx, "journalctl", "-u", "edge-cd", "--no-pager")
	if err != nil {
		return "", fmt.Errorf("failed to get edge-cd service logs: %w (stderr: %s)", err, stderr)
	}
	return stdout, nil
}

// waitForReconciliationLoop waits for edge-cd to complete a reconciliation loop
// It checks that the edge-cd service is active and reconciliation has occurred
func waitForReconciliationLoop(ctx execcontext.Context, sshClient *ssh.Client, timeoutSeconds int) error {
	// Simple approach: wait for the service to be running and stable
	// edge-cd polls every 5 seconds, so wait at least one full cycle plus buffer
	waitTime := 10 * time.Second
	if timeoutSeconds > 0 && waitTime > time.Duration(timeoutSeconds)*time.Second {
		waitTime = time.Duration(timeoutSeconds) * time.Second
	}

	slog.Debug("Waiting for edge-cd reconciliation loop", "waitTime", waitTime)
	time.Sleep(waitTime)

	// Verify the service is still active
	_, _, err := sshClient.Run(ctx, "systemctl", "is-active", "edge-cd")
	if err != nil {
		return fmt.Errorf("edge-cd service is not active after waiting: %w", err)
	}

	slog.Debug("Reconciliation loop wait completed")
	return nil
}

// verifyFileContent verifies that a file on the target VM has the expected content
func verifyFileContent(
	ctx execcontext.Context,
	sshClient *ssh.Client,
	filePath string,
	expectedContent string,
) error {
	// Read file via SSH
	stdout, stderr, err := sshClient.Run(ctx, "cat", filePath)
	if err != nil {
		// Check if it's a file not found error
		if strings.Contains(stderr, "No such file or directory") {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("failed to read file %s: %w (stderr: %s)", filePath, err, stderr)
	}

	// Compare content
	if stdout != expectedContent {
		// Create helpful diff message
		expectedSnippet := expectedContent
		actualSnippet := stdout
		if len(expectedContent) > 100 {
			expectedSnippet = expectedContent[:100] + "..."
		}
		if len(stdout) > 100 {
			actualSnippet = stdout[:100] + "..."
		}

		return fmt.Errorf(
			"file content mismatch for %s:\n  expected: %q\n  got: %q",
			filePath,
			expectedSnippet,
			actualSnippet,
		)
	}

	slog.Debug("File content verified", "path", filePath)
	return nil
}

// pushChangesToGitRepo pushes configuration changes to the git server
func pushChangesToGitRepo(
	ctx execcontext.Context,
	gitRepoURL string,
	sshKeyPath string,
	changes map[string]string,
	commitMessage string,
) error {
	// Create temp directory for local clone
	tempDir, err := os.MkdirTemp("", "e2e-git-push-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up git SSH command for authentication
	gitSSHCommand := fmt.Sprintf(
		"ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
		sshKeyPath,
	)
	gitEnv := append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=%s", gitSSHCommand))

	// Clone the repository
	cloneCmd := exec.Command("git", "clone", gitRepoURL, tempDir)
	cloneCmd.Env = gitEnv
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", gitRepoURL, err)
	}

	// Apply changes
	for filePath, content := range changes {
		fullPath := filepath.Join(tempDir, filePath)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
		}

		// Write file content
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		slog.Debug("Applied change to file", "path", filePath)
	}

	// Stage changes
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tempDir
	addCmd.Env = gitEnv
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Commit changes
	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Dir = tempDir
	commitCmd.Env = gitEnv
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push changes
	pushCmd := exec.Command("git", "push", "origin", "main")
	pushCmd.Dir = tempDir
	pushCmd.Env = gitEnv
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	slog.Debug("Successfully pushed changes to git repository", "url", gitRepoURL)
	return nil
}

// executeReconciliationTest orchestrates a complete reconciliation test scenario
// This combines waiting for reconciliation, pushing changes, and verifying results
func executeReconciliationTest(
	ctx execcontext.Context,
	env *TestEnvironment,
	sshClient *ssh.Client,
	scenario ReconciliationTestScenario,
) error {
	slog.Info("starting reconciliation test scenario", "name", scenario.Name)

	// Step 1: Wait for initial reconciliation loop
	slog.Debug("Waiting for initial reconciliation loop")
	if err := waitForReconciliationLoop(ctx, sshClient, 30); err != nil {
		return fmt.Errorf("initial reconciliation failed for scenario %q: %w", scenario.Name, err)
	}
	slog.Info("initial reconciliation complete")

	// Step 2: Push changes to git repo
	slog.Debug("Pushing changes to git repository")
	if err := pushChangesToGitRepo(
		ctx,
		env.GitSSHURLs["user-config"],
		env.SSHKeys.HostKeyPath,
		scenario.FileChanges,
		scenario.CommitMessage,
	); err != nil {
		return fmt.Errorf("failed to push changes for scenario %q: %w", scenario.Name, err)
	}
	slog.Info("pushed changes to git repo")

	// Step 3: Wait for edge-cd to reconcile changes (longer timeout)
	slog.Debug("Waiting for reconciliation after changes")
	if err := waitForReconciliationLoop(ctx, sshClient, 60); err != nil {
		return fmt.Errorf("reconciliation after changes failed for scenario %q: %w", scenario.Name, err)
	}
	slog.Info("reconciliation after changes complete")

	// Step 4: Verify each file on target VM
	slog.Debug("Verifying expected files on target VM")
	for targetPath, expectedContent := range scenario.ExpectedTargetFiles {
		if err := verifyFileContent(ctx, sshClient, targetPath, expectedContent); err != nil {
			return fmt.Errorf("file verification failed for scenario %q: %w", scenario.Name, err)
		}
		slog.Info("verified file", "path", targetPath)
	}

	slog.Info("reconciliation test scenario passed", "name", scenario.Name)
	return nil
}
