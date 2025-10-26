package provision

import (
	"fmt"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execution"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// CloneOrPullRepo clones a git repository if it doesn't exist at the destination path,
// otherwise it performs a git pull to update it.
// Uses the default branch "main".
func CloneOrPullRepo(runner ssh.Runner, repoURL, destPath string) error {
	return CloneOrPullRepoWithBranch(runner, repoURL, destPath, "main")
}

// CloneOrPullRepoWithBranch clones a git repository with a specific branch.
// If the repository already exists, it performs a git pull to update it.
func CloneOrPullRepoWithBranch(runner ssh.Runner, repoURL, destPath, branch string) error {
	return CloneOrPullRepoWithBranchAndEnv(runner, repoURL, destPath, branch, "", "")
}

// CloneOrPullRepoWithEnv clones a git repository with optional environment variables.
// The env parameter should be a complete environment variable assignment (e.g., "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no").
// The prependCmd is optional privilege escalation prefix (e.g., "sudo").
// Uses the default branch "main".
func CloneOrPullRepoWithEnv(runner ssh.Runner, repoURL, destPath, env string) error {
	return CloneOrPullRepoWithBranchAndEnv(runner, repoURL, destPath, "main", env, "")
}

// CloneOrPullRepoWithBranchAndEnv clones a git repository with a specific branch and optional environment variables.
// The env parameter should be a complete environment variable assignment (e.g., "GIT_SSH_COMMAND=ssh -o...").
// The prependCmd is optional privilege escalation prefix (e.g., "sudo").
//
// This function properly composes both the prepend command and environment variables into the remote execution
// via the CommandBuilder and SSH RunWithBuilder interface. No command string concatenation or environment stripping needed.
func CloneOrPullRepoWithBranchAndEnv(
	runner ssh.Runner,
	repoURL, destPath, branch, env, prependCmd string,
) error {
	var baseCmd, op string
	// Check if repository already exists
	checkCmd := fmt.Sprintf("[ -d %s ]", destPath)
	_, _, err := runner.Run(checkCmd)
	if err != nil { // Directory does not exist, so clone
		// Build clone command
		baseCmd = fmt.Sprintf("git clone -b %s %s %s", branch, repoURL, destPath)
		op = "clon"
	} else {
		// Directory exists, so pull
		baseCmd = fmt.Sprintf("git -C %s pull", destPath)
		op = "pull"
	}

	builder := execution.
		NewCommandBuilder(baseCmd).
		WithPrependCmd(prependCmd)

	// Add optional environment variable (e.g., "GIT_SSH_COMMAND=...")
	if env != "" {
		envKey, envValue := parseEnvVar(env)
		if envKey != "" {
			builder.WithEnvironment(envKey, envValue)
		}
	}

	fmt.Printf("%sing repository %s (branch: %s) to %s...\n", op, repoURL, branch, destPath)

	// Use BuilderRunner if available, otherwise fall back to regular Run
	stdout, stderr, cloneErr := runner.RunWithBuilder(builder)
	if cloneErr != nil {
		return fmt.Errorf(
			"failed to %se repository %s (branch: %s): %w. Stdout: %s, Stderr: %s",
			op,
			repoURL,
			branch,
			cloneErr,
			stdout,
			stderr,
		)
	}

	fmt.Printf("Successfully %sed repository %s (branch: %s).\n", op, repoURL, branch)

	return nil
}

// parseEnvVar parses an environment variable string in the format "KEY=value" and returns the key and value separately.
// If the format is invalid, it returns empty strings.
func parseEnvVar(envVar string) (key, value string) {
	parts := strings.SplitN(envVar, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
