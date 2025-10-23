package provision

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"sigs.k8s.io/yaml"
)

// PackageManager holds the commands for a specific package manager.
type PackageManager struct {
	UpdateCmd  []string
	InstallCmd []string
}

// PackageManagerCommands represents the structure of the package manager YAML files.
type PackageManagerCommands struct {
	Update  []string `yaml:"update"`
	Install []string `yaml:"install"`
}

// LoadPackageManager reads a package manager's configuration from a YAML file.
func LoadPackageManager(pkgMgr string, configsPath string) (*PackageManager, error) {
	yamlPath := filepath.Join(configsPath, pkgMgr+".yaml")
	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package manager config %s: %w", yamlPath, err)
	}

	var commands PackageManagerCommands
	err = yaml.Unmarshal(yamlFile, &commands)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal package manager config %s: %w", yamlPath, err)
	}

	return &PackageManager{
		UpdateCmd:  commands.Update,
		InstallCmd: commands.Install,
	}, nil
}

// ProvisionPackages installs a list of packages on the remote device.
func ProvisionPackages(runner ssh.Runner, packages []string, pkgMgr, configsPath string) error {
	pm, err := LoadPackageManager(pkgMgr, configsPath)
	if err != nil {
		return err
	}

	// Update package manager repos once
	if len(pm.UpdateCmd) > 0 {
		updateCmdStr := strings.Join(pm.UpdateCmd, " ")
		fmt.Printf("Updating package manager using %s...\n", pkgMgr)
		if stdout, stderr, err := runner.Run(updateCmdStr); err != nil {
			return fmt.Errorf("failed to update package manager: %w. Stdout: %s, Stderr: %s", err, stdout, stderr)
		}
	}

	// Install all packages in one command
	if len(packages) > 0 {
		installCmdSlice := append(pm.InstallCmd, packages...)
		installCmdStr := strings.Join(installCmdSlice, " ")
		fmt.Printf("Installing packages using %s...\n", pkgMgr)
		if stdout, stderr, err := runner.Run(installCmdStr); err != nil {
			return fmt.Errorf("failed to install packages: %w. Stdout: %s, Stderr: %s", err, stdout, stderr)
		}
	}

	fmt.Printf("Successfully provisioned packages.\n")
	return nil
}