package provision

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
)

var (
	errCloneRepo       = errors.New("failed to clone repository")
	errPullRepo        = errors.New("failed to pull repository")
	errCloneOrPullRepo = errors.New("failed to clone or pull repository")
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
				slog.Info("cloning repository", "url", repo.URL, "branch", repo.Branch, "destPath", destPath)
		stdout, stderr, cloneErr := runner.Run(execCtx, "git", "clone", "-b", repo.Branch, repo.URL, destPath)
		if cloneErr != nil {
			return flaterrors.Join(
				cloneErr,
				fmt.Errorf("url=%s branch=%s stdout=%s stderr=%s", repo.URL, repo.Branch, stdout, stderr),
				errCloneRepo,
				errCloneOrPullRepo,
			)
		}
	} else {
		// Directory exists, so pull
		op = "pull"
				slog.Info("pulling repository", "url", repo.URL, "branch", repo.Branch, "destPath", destPath)
		stdout, stderr, pullErr := runner.Run(execCtx, "git", "-C", destPath, "pull")
		if pullErr != nil {
			return flaterrors.Join(
				pullErr,
				fmt.Errorf("url=%s branch=%s stdout=%s stderr=%s", repo.URL, repo.Branch, stdout, stderr),
				errPullRepo,
				errCloneOrPullRepo,
			)
		}
	}

	slog.Info("successfully cloned/pulled repository", "url", repo.URL, "branch", repo.Branch, "operation", op)

	return nil
}
