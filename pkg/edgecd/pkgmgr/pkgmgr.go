package pkgmgr

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PackageManager interface defines operations for package management
type PackageManager interface {
	Update() error
	Install(packages []string) error
	Upgrade(packages []string) error
}

// packageManager is the concrete implementation
type packageManager struct {
	name   string
	config *PackageManagerConfig
}

// PackageManagerConfig represents the YAML configuration structure
type PackageManagerConfig struct {
	Update  []string `yaml:"update"`
	Install []string `yaml:"install"`
	Upgrade []string `yaml:"upgrade"`
}

// NewPackageManager creates a new PackageManager by loading configuration
// from {edgeCDRepoPath}/cmd/edge-cd/package-managers/{name}.yaml
func NewPackageManager(name string, edgeCDRepoPath string) (PackageManager, error) {
	// Load config from cmd/edge-cd/package-managers/{name}.yaml
	configPath := filepath.Join(edgeCDRepoPath, "cmd", "edge-cd", "package-managers", fmt.Sprintf("%s.yaml", name))
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package manager config: %w", err)
	}

	var config PackageManagerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse package manager config: %w", err)
	}

	return &packageManager{
		name:   name,
		config: &config,
	}, nil
}

// Update runs the package manager update command
func (pm *packageManager) Update() error {
	slog.Info("Updating package manager cache", "packageManager", pm.name)

	if len(pm.config.Update) == 0 {
		return fmt.Errorf("update command not configured")
	}

	cmd := exec.Command(pm.config.Update[0], pm.config.Update[1:]...)
	if err := cmd.Run(); err != nil {
		slog.Error("Package manager update failed", "packageManager", pm.name, "error", err)
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// Install runs update and then installs the specified packages
func (pm *packageManager) Install(packages []string) error {
	if len(packages) == 0 {
		slog.Info("No packages to install")
		return nil
	}

	slog.Info("Installing packages", "packageManager", pm.name, "packages", packages)

	// Run update first
	if err := pm.Update(); err != nil {
		return err
	}

	if len(pm.config.Install) == 0 {
		return fmt.Errorf("install command not configured")
	}

	// Build command: install_cmd + packages
	args := append(pm.config.Install[1:], packages...)
	cmd := exec.Command(pm.config.Install[0], args...)

	if err := cmd.Run(); err != nil {
		slog.Error("Package installation failed", "packageManager", pm.name, "error", err)
		return fmt.Errorf("install failed: %w", err)
	}

	return nil
}

// Upgrade runs update and then upgrades the specified packages
func (pm *packageManager) Upgrade(packages []string) error {
	if len(packages) == 0 {
		slog.Info("No packages to upgrade")
		return nil
	}

	slog.Info("Upgrading packages", "packageManager", pm.name, "packages", packages)

	// Run update first
	if err := pm.Update(); err != nil {
		return err
	}

	if len(pm.config.Upgrade) == 0 {
		return fmt.Errorf("upgrade command not configured")
	}

	// Build command: upgrade_cmd + packages
	args := append(pm.config.Upgrade[1:], packages...)
	cmd := exec.Command(pm.config.Upgrade[0], args...)

	if err := cmd.Run(); err != nil {
		slog.Error("Package upgrade failed", "packageManager", pm.name, "error", err)
		return fmt.Errorf("upgrade failed: %w", err)
	}

	return nil
}
