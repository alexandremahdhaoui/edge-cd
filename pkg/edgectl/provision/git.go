package provision

import (
	"fmt"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// CloneOrPullRepo clones a git repository if it doesn't exist at the destination path,
// otherwise it performs a git pull to update it.
func CloneOrPullRepo(runner ssh.Runner, repoURL, destPath string) error {
	// Check if repository already exists
	checkCmd := fmt.Sprintf("[ -d %s ]", destPath)
	_, _, err := runner.Run(checkCmd)

	if err != nil { // Directory does not exist, so clone
		cloneCmd := fmt.Sprintf("git clone %s %s", repoURL, destPath)
		fmt.Printf("Cloning repository %s to %s...\n", repoURL, destPath)
		stdout, stderr, cloneErr := runner.Run(cloneCmd)
		if cloneErr != nil {
			return fmt.Errorf("failed to clone repository %s: %w. Stdout: %s, Stderr: %s", repoURL, cloneErr, stdout, stderr)
		}
		fmt.Printf("Successfully cloned repository %s.\n", repoURL)
	} else { // Directory exists, so pull
		pullCmd := fmt.Sprintf("git -C %s pull", destPath)
		fmt.Printf("Pulling latest changes for repository %s in %s...\n", repoURL, destPath)
		stdout, stderr, pullErr := runner.Run(pullCmd)
		if pullErr != nil {
			return fmt.Errorf("failed to pull repository %s: %w. Stdout: %s, Stderr: %s", repoURL, pullErr, stdout, stderr)
		}
		fmt.Printf("Successfully pulled latest changes for repository %s.\n", repoURL)
	}

	return nil
}
