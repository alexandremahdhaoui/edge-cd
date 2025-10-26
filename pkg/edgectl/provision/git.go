package provision

import (
	"fmt"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// CloneOrPullRepo clones a git repository if it doesn't exist at the destination path,
// otherwise it performs a git pull to update it.
// Uses the default branch "main".
func CloneOrPullRepo(execCtx execcontext.Context, runner ssh.Runner, repoURL, destPath string) error {
	return CloneOrPullRepoWithBranch(execCtx, runner, repoURL, destPath, "main")
}

// CloneOrPullRepoWithBranch clones a git repository with a specific branch.
// If the repository already exists, it performs a git pull to update it.
func CloneOrPullRepoWithBranch(execCtx execcontext.Context, runner ssh.Runner, repoURL, destPath, branch string) error {
	return CloneOrPullRepoWithBranchAndEnv(
		execCtx,
		runner,
		repoURL,
		destPath,
		branch,
	)
}

// CloneOrPullRepoWithEnv clones a git repository with optional environment variables.
// The env parameter should be a complete environment variable assignment (e.g., "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no").
// Uses the default branch "main".
func CloneOrPullRepoWithEnv(execCtx execcontext.Context, runner ssh.Runner, repoURL, destPath, env string) error {
	// Fork the context with an additional environment variable if provided
	var ctx execcontext.Context = execCtx
	if env != "" {
		envKey, envValue := parseEnvVar(env)
		if envKey != "" {
			// Create a new context with the additional environment variable
			newEnvs := execCtx.Envs()
			newEnvs[envKey] = envValue
			ctx = execcontext.New(newEnvs, execCtx.PrependCmd())
		}
	}
	return CloneOrPullRepoWithBranchAndEnv(ctx, runner, repoURL, destPath, "main")
}

// CloneOrPullRepoWithBranchAndEnv clones a git repository with a specific branch and execution context.
// The context parameter should contain any required environment variables (e.g., GIT_SSH_COMMAND) and
// prepend commands (e.g., sudo).
//
// This function properly composes both the prepend command and environment variables into the remote execution
// via the Context and SSH Run interface.
func CloneOrPullRepoWithBranchAndEnv(
	execCtx execcontext.Context,
	runner ssh.Runner,
	repoURL, destPath, branch string,
) error {
	var baseCmd, op string
	// Check if repository already exists
	checkCmd := fmt.Sprintf("[ -d %s ]", destPath)
	_, _, err := runner.Run(execCtx, checkCmd)
	if err != nil { // Directory does not exist, so clone
		// Build clone command
		baseCmd = fmt.Sprintf("git clone -b %s %s %s", branch, repoURL, destPath)
		op = "clon"
	} else {
		// Directory exists, so pull
		baseCmd = fmt.Sprintf("git -C %s pull", destPath)
		op = "pull"
	}

	fmt.Printf("%sing repository %s (branch: %s) to %s...\n", op, repoURL, branch, destPath)

	// Execute with the provided context
	stdout, stderr, cloneErr := runner.Run(execCtx, baseCmd)
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
