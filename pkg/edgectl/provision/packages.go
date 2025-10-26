package provision

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"sigs.k8s.io/yaml"
)

// PackageManager holds the commands for a specific package manager.
type PackageManager struct {
	Update  []string `yaml:"update"`
	Install []string `yaml:"install"`
}

// LoadPackageManager reads a package manager's configuration from a local path.
func LoadPackageManager(pkgMgr string, rootConfigsPath string) (*PackageManager, error) {
	configPath := filepath.Join(
		rootConfigsPath,
		"cmd",
		"edge-cd",
		"package-managers",
		pkgMgr+".yaml",
	)
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package manager config from %s: %w", configPath, err)
	}

	var commands PackageManager
	err = yaml.Unmarshal(yamlFile, &commands)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal package manager config from %s: %w",
			configPath,
			err,
		)
	}

	return &PackageManager{
		Update:  commands.Update,
		Install: commands.Install,
	}, nil
}

// ProvisionPackages installs a list of packages on the remote device.
func ProvisionPackages(
	execCtx execcontext.Context,
	runner ssh.Runner,
	packages []string,
	pkgMgr string,
	localPkgMgrRepoPath string,
	remoteEdgeCDRepoURL string,
	remoteEdgeCDRepoDestPath string,
) error {
	// Load package manager configuration from the locally cloned repository
	pm, err := LoadPackageManager(pkgMgr, localPkgMgrRepoPath)
	if err != nil {
		return err
	}

	// Clone the edge-cd repository to its destination path on the remote device
	fmt.Printf(
		"Cloning edge-cd repository %s to %s on remote...\n",
		remoteEdgeCDRepoURL,
		remoteEdgeCDRepoDestPath,
	)

	// Execute with the provided context
	if stdout, stderr, cloneErr := runner.Run(execCtx, "git", "clone", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath); cloneErr != nil {
		return fmt.Errorf(
			"failed to clone edge-cd repository %s on remote: %w. Stdout: %s, Stderr: %s",
			remoteEdgeCDRepoURL,
			cloneErr,
			stdout,
			stderr,
		)
	}

	// Update package manager repos once
	if len(pm.Update) > 0 {
		fmt.Printf("Updating package manager using %s...\n", pkgMgr)
		if stdout, stderr, updateErr := runner.Run(execCtx, pm.Update...); updateErr != nil {
			return fmt.Errorf(
				"failed to update package manager: %w. Stdout: %s, Stderr: %s",
				updateErr,
				stdout,
				stderr,
			)
		}
	}

	// Install all packages in one command
	if len(packages) > 0 {
		fmt.Printf("Installing packages using %s...\n", pkgMgr)
		if stdout, stderr, installErr := runner.Run(execCtx, append(pm.Install, packages...)...); installErr != nil {
			return fmt.Errorf(
				"failed to install packages: %w. Stdout: %s, Stderr: %s",
				installErr,
				stdout,
				stderr,
			)
		}
	}

	fmt.Printf("Successfully provisioned packages.\n")
	return nil
}
