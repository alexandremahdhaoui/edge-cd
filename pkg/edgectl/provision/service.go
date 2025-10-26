package provision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"gopkg.in/yaml.v3"
)

// ServiceManagerConfig represents the structure of service manager config files
type ServiceManagerConfig struct {
	Commands      map[string][]string `yaml:"commands"`
	EdgeCDService struct {
		DestinationPath string `yaml:"destinationPath"`
	} `yaml:"edgeCDService"`
}

// SetupEdgeCDService sets up and enables the edge-cd service on the remote device.
// The context parameter should contain any required prepend commands (e.g., "sudo").
func SetupEdgeCDService(
	execCtx execcontext.Context,
	runner ssh.Runner,
	serviceManagerName, edgeCDRepoPath string,
) error {
	// Load the service manager configuration
	config, err := loadServiceManagerConfig(edgeCDRepoPath, serviceManagerName)
	if err != nil {
		return err
	}

	// Determine source and destination paths for the service file
	serviceSourcePath := filepath.Join(
		edgeCDRepoPath,
		"cmd",
		"edge-cd",
		"service-managers",
		serviceManagerName,
		fmt.Sprintf("edge-cd.%s", serviceManagerName),
	)
	serviceDestPath := config.EdgeCDService.DestinationPath

	// Copy service file to destination using the context
	copyCmd := fmt.Sprintf("cp %s %s", serviceSourcePath, serviceDestPath)
	fmt.Printf("Copying service file from %s to %s...\n", serviceSourcePath, serviceDestPath)
	stdout, stderr, err := runner.Run(execCtx, copyCmd)
	if err != nil {
		return fmt.Errorf(
			"failed to copy service file: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	// Build and execute enable command
	enableCmd := buildCommand(config.Commands["enable"], "", "edge-cd")
	fmt.Printf("Enabling service %s...\n", serviceManagerName)
	stdout, stderr, err = runner.Run(execCtx, enableCmd)
	if err != nil {
		return fmt.Errorf(
			"failed to enable service: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	// Build and execute start command (fallback to restart if start doesn't exist)
	var startCmd string
	if len(config.Commands["start"]) > 0 {
		startCmd = buildCommand(config.Commands["start"], "", "edge-cd")
	} else if len(config.Commands["restart"]) > 0 {
		// Some service managers (like procd) use restart instead of start
		startCmd = buildCommand(config.Commands["restart"], "", "edge-cd")
	}

	if startCmd != "" {
		fmt.Printf("Starting service %s...\n", serviceManagerName)
		stdout, stderr, err = runner.Run(execCtx, startCmd)
		if err != nil {
			return fmt.Errorf(
				"failed to start service: %w. Stdout: %s, Stderr: %s",
				err,
				stdout,
				stderr,
			)
		}
	}

	return nil
}

// loadServiceManagerConfig loads the service manager configuration from the YAML file
func loadServiceManagerConfig(
	edgeCDRepoPath, serviceManagerName string,
) (*ServiceManagerConfig, error) {
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

// buildCommand constructs a command from a command pattern array.
// Replaces "__SERVICE_NAME__" with serviceName throughout the command parts.
// Returns a properly formatted command string.
func buildCommand(cmdPattern []string, _ string, serviceName string) string {
	if len(cmdPattern) == 0 {
		return ""
	}

	// Build the command, replacing placeholders
	var parts []string
	for _, part := range cmdPattern {
		if part == "%s" {
			// Skip the prepend placeholder - prepend is handled via execution.Context
			continue
		} else {
			// Replace __SERVICE_NAME__ placeholder within the string
			replacedPart := strings.ReplaceAll(part, "__SERVICE_NAME__", serviceName)
			parts = append(parts, replacedPart)
		}
	}

	return strings.Join(parts, " ")
}
