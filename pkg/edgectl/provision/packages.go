package provision

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
	"sigs.k8s.io/yaml"
)

var (
	errReadPackageManagerConfig    = errors.New("failed to read package manager config")
	errUnmarshalPackageManagerConfig = errors.New("failed to unmarshal package manager config")
	errCloneEdgeCDRepo             = errors.New("failed to clone edge-cd repository on remote")
	errUpdatePackageManager        = errors.New("failed to update package manager")
	errInstallPackages             = errors.New("failed to install packages")
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
		return nil, flaterrors.Join(err, fmt.Errorf("configPath=%s", configPath), errReadPackageManagerConfig)
	}

	var commands PackageManager
	err = yaml.Unmarshal(yamlFile, &commands)
	if err != nil {
		return nil, flaterrors.Join(err, fmt.Errorf("configPath=%s", configPath), errUnmarshalPackageManagerConfig)
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
	slog.Info("cloning edge-cd repository to remote", "url", remoteEdgeCDRepoURL, "destPath", remoteEdgeCDRepoDestPath)

	// Execute with the provided context
	stdout, stderr, cloneErr := runner.Run(execCtx, "git", "clone", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
	if cloneErr != nil {
		return flaterrors.Join(cloneErr, fmt.Errorf("url=%s stdout=%s stderr=%s", remoteEdgeCDRepoURL, stdout, stderr), errCloneEdgeCDRepo)
	}

	// Update package manager repos once
	if len(pm.Update) > 0 {
		slog.Info("updating package manager", "packageManager", pkgMgr)
		if stdout, stderr, updateErr := runner.Run(execCtx, pm.Update...); updateErr != nil {
			return flaterrors.Join(updateErr, fmt.Errorf("stdout=%s stderr=%s", stdout, stderr), errUpdatePackageManager)
		}
	}

	// Install all packages in one command
	if len(packages) > 0 {
		slog.Info("installing packages", "packageManager", pkgMgr, "packages", packages)
		if stdout, stderr, installErr := runner.Run(execCtx, append(pm.Install, packages...)...); installErr != nil {
			return flaterrors.Join(installErr, fmt.Errorf("stdout=%s stderr=%s", stdout, stderr), errInstallPackages)
		}
	}

	slog.Info("successfully provisioned packages")
	return nil
}
