package provision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execution"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"gopkg.in/yaml.v3"
)

// ServiceManagerConfig represents the structure of service manager config files
type ServiceManagerConfig struct {
	Commands map[string][]string `yaml:"commands"`
	EdgeCDService struct {
		DestinationPath string `yaml:"destinationPath"`
	} `yaml:"edgeCDService"`
}

// SetupEdgeCDService sets up and enables the edge-cd service on the remote device.
// prependCmd is prepended to commands that require elevated privileges (e.g., "sudo")
func SetupEdgeCDService(runner ssh.Runner, serviceManagerName, edgeCDRepoPath, prependCmd string) error {
	// Determine source and destination paths for the service file
	serviceSourcePath := filepath.Join(
		edgeCDRepoPath,
		"cmd",
		"edge-cd",
		"service-managers",
		serviceManagerName,
		"service",
	)
	var serviceDestPath string
	var enableCmd, startCmd string
	var stdout, stderr string // Declare once at the beginning
	var err error             // Declare once at the beginning

	switch serviceManagerName {
	case "systemd":
		serviceDestPath = "/etc/systemd/system/edge-cd.service"

		// Load systemd config to get command patterns
		config, err := loadServiceManagerConfig(edgeCDRepoPath, serviceManagerName)
		if err != nil {
			return fmt.Errorf("failed to load service manager config: %w", err)
		}

		// Check if the service is masked by a symlink to /dev/null and remove it
		checkMaskedSymlinkCmd := fmt.Sprintf(
			"test -L %s && readlink %s | grep -q '/dev/null'",
			serviceDestPath,
			serviceDestPath,
		)
		_, _, err = runner.Run(checkMaskedSymlinkCmd)
		if err == nil { // If the symlink to /dev/null exists
			// Use CommandBuilder to compose rm command with optional prepend
			baseRmCmd := fmt.Sprintf("rm %s", serviceDestPath)
			rmBuilder := execution.NewCommandBuilder(baseRmCmd)
			if prependCmd != "" {
				rmBuilder.WithPrependCmd(prependCmd)
			}
			rmMaskedSymlinkCmd := rmBuilder.ComposeCommand()
			fmt.Printf("Removing masked symlink %s...\n", serviceDestPath)
			stdout, stderr, err = runner.Run(rmMaskedSymlinkCmd)
			if err != nil {
				return fmt.Errorf(
					"failed to remove masked symlink: %w. Stdout: %s, Stderr: %s",
					err,
					stdout,
					stderr,
				)
			}
		}

		// Build daemon-reload command from config
		daemonReloadCmd := buildCommand(config.Commands["daemonReload"], prependCmd, "")
		fmt.Printf("Reloading systemd daemon...\n")
		stdout, stderr, err = runner.Run(daemonReloadCmd)
		if err != nil {
			return fmt.Errorf(
				"failed to reload systemd daemon: %w. Stdout: %s, Stderr: %s",
				err,
				stdout,
				stderr,
			)
		}

		// Build unmask command from config
		unmaskCmd := buildCommand(config.Commands["unmask"], prependCmd, "edge-cd.service")
		fmt.Printf("Unmasking service %s...\n", serviceManagerName)
		stdout, stderr, err = runner.Run(unmaskCmd)
		if err != nil {
			return fmt.Errorf(
				"failed to unmask service: %w. Stdout: %s, Stderr: %s",
				err,
				stdout,
				stderr,
			)
		}

		// Build enable and start commands from config
		enableCmd = buildCommand(config.Commands["enable"], prependCmd, "edge-cd.service")
		startCmd = buildCommand(config.Commands["start"], prependCmd, "edge-cd.service")

	case "procd":
		serviceDestPath = "/etc/init.d/edge-cd"
		enableCmd = "/etc/init.d/edge-cd enable"
		startCmd = "/etc/init.d/edge-cd start"
	default:
		return fmt.Errorf("unsupported service manager: %s", serviceManagerName)
	}

	// Copy service file to destination
	// Use CommandBuilder to compose cp command with optional prepend
	baseCpCmd := fmt.Sprintf("cp %s %s", serviceSourcePath, serviceDestPath)
	cpBuilder := execution.NewCommandBuilder(baseCpCmd)
	if prependCmd != "" {
		cpBuilder.WithPrependCmd(prependCmd)
	}
	copyCmd := cpBuilder.ComposeCommand()
	fmt.Printf("Copying service file from %s to %s...\n", serviceSourcePath, serviceDestPath)
	stdout, stderr, err = runner.Run(copyCmd)
	if err != nil {
		return fmt.Errorf(
			"failed to copy service file: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	fmt.Printf("Enabling service %s...\n", serviceManagerName)
	stdout, stderr, err = runner.Run(enableCmd)
	if err != nil {
		return fmt.Errorf(
			"failed to enable service: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	fmt.Printf("Starting service %s...\n", serviceManagerName)
	stdout, stderr, err = runner.Run(startCmd)
	if err != nil {
		return fmt.Errorf(
			"failed to start service: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	return nil
}

// loadServiceManagerConfig loads the service manager configuration from the YAML file
func loadServiceManagerConfig(edgeCDRepoPath, serviceManagerName string) (*ServiceManagerConfig, error) {
	configPath := filepath.Join(
		edgeCDRepoPath,
		"cmd",
		"edge-cd",
		"service-managers",
		serviceManagerName,
		"config.yaml",
	)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service manager config file: %w", err)
	}

	var config ServiceManagerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse service manager config: %w", err)
	}

	return &config, nil
}

// buildCommand constructs a shell command from a command pattern array using CommandBuilder.
// Replaces "%s" with prependCmd and "__SERVICE_NAME__" with serviceName.
// Returns a properly formatted command string with prepend handled via CommandBuilder.
func buildCommand(cmdPattern []string, prependCmd, serviceName string) string {
	if len(cmdPattern) == 0 {
		return ""
	}

	// First pass: build the base command without the %s placeholder
	var parts []string
	for _, part := range cmdPattern {
		if part == "%s" {
			// Skip the placeholder - we'll handle prepend via CommandBuilder
			continue
		} else if part == "__SERVICE_NAME__" {
			parts = append(parts, serviceName)
		} else {
			parts = append(parts, part)
		}
	}

	baseCmd := strings.Join(parts, " ")

	// Use CommandBuilder to compose with prependCmd
	builder := execution.NewCommandBuilder(baseCmd)
	if prependCmd != "" {
		builder.WithPrependCmd(prependCmd)
	}

	// For buildCommand, we return just the composed string (not an exec.Cmd)
	// Extract it from the builder's composition
	return builder.ComposeCommand()
}
