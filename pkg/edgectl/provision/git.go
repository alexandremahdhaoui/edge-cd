package provision

import (
	"fmt"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

type GitRepo struct {
	URL    string
	Branch string
}

// CloneOrPullRepo clones a git repository with a specific branch and execution context.
// The context parameter should contain any required environment variables (e.g., GIT_SSH_COMMAND) and
// prepend commands (e.g., sudo).
//
// This function properly composes both the prepend command and environment variables into the remote execution
// via the Context and SSH Run interface.
func CloneOrPullRepo(
	execCtx execcontext.Context,
	runner ssh.Runner,
	destPath string,
	repo GitRepo,
) error {
	var op string
	// Check if repository already exists
	_, _, err := runner.Run(execCtx, "test", "-d", destPath)
	if err != nil { // Directory does not exist, so clone
		// Clone the repository
		op = "clon"
		fmt.Printf("%sing repository %s (branch: %s) to %s...\n", op, repo.URL, repo.Branch, destPath)
		stdout, stderr, cloneErr := runner.Run(execCtx, "git", "clone", "-b", repo.Branch, repo.URL, destPath)
		if cloneErr != nil {
			return fmt.Errorf(
				"failed to %se repository %s (branch: %s): %w. Stdout: %s, Stderr: %s",
				op,
				repo.URL,
				repo.Branch,
				cloneErr,
				stdout,
				stderr,
			)
		}
	} else {
		// Directory exists, so pull
		op = "pull"
		fmt.Printf("%sing repository %s (branch: %s) to %s...\n", op, repo.URL, repo.Branch, destPath)
		stdout, stderr, pullErr := runner.Run(execCtx, "git", "-C", destPath, "pull")
		if pullErr != nil {
			return fmt.Errorf(
				"failed to %se repository %s (branch: %s): %w. Stdout: %s, Stderr: %s",
				op,
				repo.URL,
				repo.Branch,
				pullErr,
				stdout,
				stderr,
			)
		}
	}

	fmt.Printf("Successfully %sed repository %s (branch: %s).\n", op, repo.URL, repo.Branch)

	return nil
}
