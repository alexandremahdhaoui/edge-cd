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
//
// Uses robust fetch + reset approach for syncing existing repositories, which cannot fail due to:
// - Merge conflicts
// - Local modifications
// - Detached HEAD state
func CloneOrPullRepo(
	execCtx execcontext.Context,
	runner ssh.Runner,
	destPath string,
	repo GitRepo,
) error {
	// Check if repository already exists
	_, _, err := runner.Run(execCtx, "test", "-d", destPath)
	if err != nil {
		// Directory does not exist, clone it
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

		// After clone, fetch and pull to ensure up to date
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", destPath, "fetch", "origin", repo.Branch)
		if err != nil {
			return flaterrors.Join(
				err,
				fmt.Errorf("url=%s branch=%s stdout=%s stderr=%s", repo.URL, repo.Branch, stdout, stderr),
				errCloneRepo,
				errCloneOrPullRepo,
			)
		}

		stdout, stderr, err = runner.Run(execCtx, "git", "-C", destPath, "pull")
		if err != nil {
			return flaterrors.Join(
				err,
				fmt.Errorf("url=%s branch=%s stdout=%s stderr=%s", repo.URL, repo.Branch, stdout, stderr),
				errCloneRepo,
				errCloneOrPullRepo,
			)
		}

		slog.Info("repository cloned successfully", "url", repo.URL, "branch", repo.Branch, "destPath", destPath)
	} else {
		// Directory exists, sync it using fetch + reset (idempotent and robust)
		slog.Info("repository already exists, syncing latest changes", "url", repo.URL, "branch", repo.Branch, "destPath", destPath)

		// git fetch origin <branch>
		stdout, stderr, err := runner.Run(execCtx, "git", "-C", destPath, "fetch", "origin", repo.Branch)
		if err != nil {
			return flaterrors.Join(
				err,
				fmt.Errorf("url=%s branch=%s stdout=%s stderr=%s", repo.URL, repo.Branch, stdout, stderr),
				errPullRepo,
				errCloneOrPullRepo,
			)
		}

		// git reset --hard FETCH_HEAD (force update to match remote exactly)
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", destPath, "reset", "--hard", "FETCH_HEAD")
		if err != nil {
			return flaterrors.Join(
				err,
				fmt.Errorf("url=%s branch=%s stdout=%s stderr=%s", repo.URL, repo.Branch, stdout, stderr),
				errPullRepo,
				errCloneOrPullRepo,
			)
		}

		slog.Info("repository synced successfully", "url", repo.URL, "branch", repo.Branch, "destPath", destPath)
	}

	return nil
}
