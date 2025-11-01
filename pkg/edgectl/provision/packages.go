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
	errInstallYq                   = errors.New("failed to install yq")
	errCheckYqInstallation         = errors.New("failed to check yq installation")
	errDownloadYq                  = errors.New("failed to download yq")
	errMakeYqExecutable            = errors.New("failed to make yq executable")
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

	// Clone or sync the edge-cd repository on the remote device (idempotency check)
	// Uses sparse checkout to only fetch cmd/edge-cd directory
	// Check if repository already exists
	_, _, err = runner.Run(execCtx, "test", "-d", remoteEdgeCDRepoDestPath)
	if err != nil {
		// Directory does not exist, clone it with sparse checkout
		slog.Info("cloning edge-cd repository to remote with sparse checkout", "url", remoteEdgeCDRepoURL, "destPath", remoteEdgeCDRepoDestPath)

		// git clone --filter=blob:none --no-checkout
		stdout, stderr, cloneErr := runner.Run(execCtx, "git", "clone", "--filter=blob:none", "--no-checkout", remoteEdgeCDRepoURL, remoteEdgeCDRepoDestPath)
		if cloneErr != nil {
			return flaterrors.Join(cloneErr, fmt.Errorf("url=%s stdout=%s stderr=%s", remoteEdgeCDRepoURL, stdout, stderr), errCloneEdgeCDRepo)
		}

		// git sparse-checkout init
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "init")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		// git sparse-checkout set "cmd/edge-cd"
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "set", "cmd/edge-cd")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		// git checkout main
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "checkout", "main")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		// git fetch origin main
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "fetch", "origin", "main")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		// git pull (final sync after checkout)
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "pull")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		slog.Info("edge-cd repository cloned successfully with sparse checkout", "destPath", remoteEdgeCDRepoDestPath)
	} else {
		// Directory exists, sync it using fetch + reset (idempotent and robust)
		slog.Info("edge-cd repository already exists, syncing latest changes", "destPath", remoteEdgeCDRepoDestPath)

		// git sparse-checkout set "cmd/edge-cd"
		stdout, stderr, err := runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "sparse-checkout", "set", "cmd/edge-cd")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		// git fetch origin main
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "fetch", "origin", "main")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		// git reset --hard FETCH_HEAD (force update to match remote exactly)
		stdout, stderr, err = runner.Run(execCtx, "git", "-C", remoteEdgeCDRepoDestPath, "reset", "--hard", "FETCH_HEAD")
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", remoteEdgeCDRepoDestPath, stdout, stderr), errCloneEdgeCDRepo)
		}

		slog.Info("edge-cd repository synced successfully", "destPath", remoteEdgeCDRepoDestPath)
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

// InstallYq installs yq on the remote device if not already installed.
// This function is idempotent - it checks if yq is already installed before attempting installation.
func InstallYq(
	execCtx execcontext.Context,
	runner ssh.Runner,
) error {
	slog.Info("checking if yq is installed")

	// Check if yq is already installed (idempotency check)
	_, _, err := runner.Run(execCtx, "which", "yq")
	if err == nil {
		slog.Info("yq is already installed, skipping installation")
		return nil
	}

	slog.Info("yq not found, installing yq")

	// Download yq to /usr/local/bin
	stdout, stderr, err := runner.Run(
		execCtx,
		"wget",
		"-qO",
		"/usr/local/bin/yq",
		"https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64",
	)
	if err != nil {
		return flaterrors.Join(
			err,
			fmt.Errorf("stdout=%s stderr=%s", stdout, stderr),
			errDownloadYq,
		)
	}

	// Make yq executable
	stdout, stderr, err = runner.Run(execCtx, "chmod", "a+x", "/usr/local/bin/yq")
	if err != nil {
		return flaterrors.Join(
			err,
			fmt.Errorf("stdout=%s stderr=%s", stdout, stderr),
			errMakeYqExecutable,
		)
	}

	slog.Info("successfully installed yq")
	return nil
}
