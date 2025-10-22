package pkgmgr

import (
	"fmt"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// InstallPackages installs a list of packages on the remote device using opkg.
// It first checks if a package is already installed to ensure idempotency.
func InstallPackages(runner ssh.Runner, packages []string) error {
	// First, run opkg update once
	_, _, err := runner.Run("opkg update")
	if err != nil {
		return fmt.Errorf("failed to run opkg update: %w", err)
	}

	for _, pkg := range packages {
		// Check if package is already installed
		checkCmd := fmt.Sprintf("opkg list-installed | grep -q '^%s '", pkg)
		_, _, err := runner.Run(checkCmd)
		if err == nil {
			// Package is already installed, skip
			continue
		}

		// Package not installed, proceed with installation
		installCmd := fmt.Sprintf("opkg install %s", pkg)
		stdout, stderr, err := runner.Run(installCmd)
		if err != nil {
			return fmt.Errorf("failed to install package %s: %w\nStdout: %s\nStderr: %s", pkg, err, stdout, stderr)
		}
	}
	return nil
}
