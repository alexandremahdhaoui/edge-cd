package git

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// RepoManager defines operations for Git repository management
type RepoManager interface {
	CloneRepo(url, branch, destPath string, sparseCheckoutPaths []string) error
	SyncRepo(repoPath, branch string, sparseCheckoutPaths []string) error
	GetCurrentCommit(repoPath string) (string, error)
	GetCommitDiff(repoPath, oldCommit, newCommit string) ([]string, error)
}

// gitRepoManager implements RepoManager
type gitRepoManager struct{}

// NewRepoManager creates a new RepoManager instance
func NewRepoManager() RepoManager {
	return &gitRepoManager{}
}

// CloneRepo clones a Git repository with sparse checkout
func (g *gitRepoManager) CloneRepo(url, branch, destPath string, sparseCheckoutPaths []string) error {
	// Handle file:// URLs - skip git operations
	if strings.HasPrefix(url, "file://") {
		slog.Info("Skipping git clone for file:// URL", "url", url)
		return nil
	}

	slog.Info("Cloning repository", "url", url, "branch", branch, "destPath", destPath)

	// git clone --filter=blob:none --no-checkout
	cmd := exec.Command("git", "clone", "--filter=blob:none", "--no-checkout", url, destPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w: %s", err, string(output))
	}

	// git sparse-checkout init
	cmd = exec.Command("git", "-C", destPath, "sparse-checkout", "init")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sparse-checkout init failed: %w: %s", err, string(output))
	}

	// git sparse-checkout set <paths>
	args := append([]string{"-C", destPath, "sparse-checkout", "set"}, sparseCheckoutPaths...)
	cmd = exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sparse-checkout set failed: %w: %s", err, string(output))
	}

	// git checkout <branch>
	cmd = exec.Command("git", "-C", destPath, "checkout", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w: %s", err, string(output))
	}

	// git fetch origin <branch>
	cmd = exec.Command("git", "-C", destPath, "fetch", "origin", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %w: %s", err, string(output))
	}

	// git pull
	cmd = exec.Command("git", "-C", destPath, "pull")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w: %s", err, string(output))
	}

	slog.Info("Repository cloned successfully", "destPath", destPath)
	return nil
}

// SyncRepo syncs an existing Git repository
func (g *gitRepoManager) SyncRepo(repoPath, branch string, sparseCheckoutPaths []string) error {
	// Check if this is a file:// URL by checking if it's a git repo
	if _, err := os.Stat(repoPath + "/.git"); err != nil {
		// Not a git repo, skip sync
		slog.Info("Skipping git sync for non-git directory", "repoPath", repoPath)
		return nil
	}

	slog.Info("Syncing repository", "repoPath", repoPath, "branch", branch)

	// git sparse-checkout set <paths>
	args := append([]string{"-C", repoPath, "sparse-checkout", "set"}, sparseCheckoutPaths...)
	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sparse-checkout set failed: %w: %s", err, string(output))
	}

	// git fetch origin <branch>
	cmd = exec.Command("git", "-C", repoPath, "fetch", "origin", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %w: %s", err, string(output))
	}

	// git reset --hard FETCH_HEAD
	cmd = exec.Command("git", "-C", repoPath, "reset", "--hard", "FETCH_HEAD")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset failed: %w: %s", err, string(output))
	}

	slog.Info("Repository synced successfully", "repoPath", repoPath)
	return nil
}

// GetCurrentCommit returns the current commit hash
func (g *gitRepoManager) GetCurrentCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	commit := strings.TrimSpace(string(output))
	slog.Info("Got current commit", "repoPath", repoPath, "commit", commit)
	return commit, nil
}

// GetCommitDiff returns the list of files changed between two commits
func (g *gitRepoManager) GetCommitDiff(repoPath, oldCommit, newCommit string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--name-only", oldCommit, newCommit)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	files := []string{}
	if len(strings.TrimSpace(string(output))) > 0 {
		files = strings.Split(strings.TrimSpace(string(output)), "\n")
	}

	slog.Info("Got commit diff", "repoPath", repoPath, "oldCommit", oldCommit[:7], "newCommit", newCommit[:7], "filesChanged", len(files))
	return files, nil
}
