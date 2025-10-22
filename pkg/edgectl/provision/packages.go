package provision

import (
	"fmt"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// InstallPackages installs a list of packages on the remote device using apt-get.
// It first checks if each package is already installed to ensure idempotency.
func InstallPackages(runner ssh.Runner, packages []string) error {
	for _, pkg := range packages {
		// Check if package is already installed using dpkg
		checkCmd := fmt.Sprintf("dpkg -s %s &> /dev/null", pkg)
		_, _, err := runner.Run(checkCmd)
		if err == nil {
			// Package is already installed, skip
			fmt.Printf("Package %s already installed, skipping.\n", pkg)
			continue
		}

		// Install package using apt-get
		installCmd := fmt.Sprintf("apt-get update && apt-get install -y %s", pkg)
		fmt.Printf("Installing package %s...\n", pkg)
		stdout, stderr, err := runner.Run(installCmd)
		if err != nil {
			return fmt.Errorf("failed to install package %s: %w. Stdout: %s, Stderr: %s", pkg, err, stdout, stderr)
		}
		fmt.Printf("Successfully installed package %s.\n", pkg)
	}
	return nil
}