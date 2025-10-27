package provision

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
	"gopkg.in/yaml.v3"
)

var (
	errLoadServiceManagerConfig   = errors.New("failed to load service manager config")
	errCopyServiceFile            = errors.New("failed to copy service file")
	errEnableService              = errors.New("failed to enable service")
	errStartService               = errors.New("failed to start service")
	errReadServiceManagerConfigFile = errors.New("failed to read service manager config file")
	errParseServiceManagerConfig    = errors.New("failed to parse service manager config")
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
	svcmgrName, edgeCDRepoPath string,
) error {
	var stdout, stderr string
	// Load the service manager configuration
	config, err := loadServiceManagerConfig(edgeCDRepoPath, svcmgrName)
	if err != nil {
		return err
	}

	// Determine source and destination paths for the service file
	serviceSourcePath := filepath.Join(
		edgeCDRepoPath,
		"cmd",
		"edge-cd",
		"service-managers",
		svcmgrName,
		fmt.Sprintf("edge-cd.%s", svcmgrName),
	)
	serviceDestPath := config.EdgeCDService.DestinationPath

	// Copy service file to destination using the context
	slog.Info("copying service file", "source", serviceSourcePath, "dest", serviceDestPath)
	stdout, stderr, err = runner.Run(execCtx, "cp", serviceSourcePath, serviceDestPath)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("serviceSourcePath=%s serviceDestPath=%s stdout=%s stderr=%s", serviceSourcePath, serviceDestPath, stdout, stderr), errCopyServiceFile)
	}

	// Build and execute enable command
	enableCmdRaw := substituteServiceName(config.Commands["enable"], "edge-cd")
	slog.Info("enabling service", "serviceManager", svcmgrName)
	stdout, stderr, err = runner.Run(execCtx, enableCmdRaw...)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("stdout=%s stderr=%s", stdout, stderr), errEnableService)
	}

	// Build and execute start command (fallback to restart if start doesn't exist)
	var startCmdRaw []string
	if len(config.Commands["start"]) > 0 {
		startCmdRaw = substituteServiceName(config.Commands["start"], "edge-cd")
	} else if len(config.Commands["restart"]) > 0 {
		// Some service managers (like procd) use restart instead of start
		startCmdRaw = substituteServiceName(config.Commands["restart"], "edge-cd")
	}

	if len(startCmdRaw) > 0 {
		slog.Info("starting service", "serviceManager", svcmgrName)
		stdout, stderr, err = runner.Run(execCtx, startCmdRaw...)
		if err != nil {
			return flaterrors.Join(err, fmt.Errorf("stdout=%s stderr=%s", stdout, stderr), errStartService)
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
		return nil, flaterrors.Join(err, errReadServiceManagerConfigFile)
	}

	var config ServiceManagerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, flaterrors.Join(err, errParseServiceManagerConfig)
	}

	return &config, nil
}

// substituteServiceName replaces "__SERVICE_NAME__" placeholder in command arguments
func substituteServiceName(cmdArgs []string, serviceName string) []string {
	result := make([]string, len(cmdArgs))
	for i, arg := range cmdArgs {
		result[i] = strings.ReplaceAll(arg, "__SERVICE_NAME__", serviceName)
	}
	return result
}
