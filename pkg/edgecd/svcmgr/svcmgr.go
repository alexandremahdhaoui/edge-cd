package svcmgr

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServiceManager provides an interface for managing system services
type ServiceManager interface {
	Enable(serviceName string) error
	Restart(serviceName string) error
	Start(serviceName string) error
}

// serviceManager is the concrete implementation
type serviceManager struct {
	name   string
	config *ServiceManagerConfig
}

// ServiceManagerConfig represents the YAML configuration for a service manager
type ServiceManagerConfig struct {
	Commands struct {
		Enable  []string `yaml:"enable"`
		Restart []string `yaml:"restart"`
		Start   []string `yaml:"start,omitempty"`
	} `yaml:"commands"`
	EdgeCDService struct {
		DestinationPath string `yaml:"destinationPath"`
	} `yaml:"edgeCDService"`
}

// NewServiceManager creates a new ServiceManager by loading configuration
// from {edgeCDRepoPath}/cmd/edge-cd/service-managers/{name}/config.yaml
func NewServiceManager(name string, edgeCDRepoPath string) (ServiceManager, error) {
	// Load config from cmd/edge-cd/service-managers/{name}/config.yaml
	configPath := fmt.Sprintf("%s/cmd/edge-cd/service-managers/%s/config.yaml", edgeCDRepoPath, name)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service manager config: %w", err)
	}

	var config ServiceManagerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse service manager config: %w", err)
	}

	return &serviceManager{
		name:   name,
		config: &config,
	}, nil
}

// Enable enables a service to start on boot
func (sm *serviceManager) Enable(serviceName string) error {
	slog.Info("Enabling service", "service", serviceName)

	cmdArgs := sm.replaceServiceName(sm.config.Commands.Enable, serviceName)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	if err := cmd.Run(); err != nil {
		slog.Error("Service enable failed", "service", serviceName, "error", err)
		return err
	}

	return nil
}

// Restart restarts a running service
func (sm *serviceManager) Restart(serviceName string) error {
	slog.Info("Restarting service", "service", serviceName)

	cmdArgs := sm.replaceServiceName(sm.config.Commands.Restart, serviceName)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	if err := cmd.Run(); err != nil {
		slog.Error("Service restart failed", "service", serviceName, "error", err)
		return err
	}

	return nil
}

// Start starts a service
func (sm *serviceManager) Start(serviceName string) error {
	slog.Info("Starting service", "service", serviceName)

	// If Start command is not defined, skip
	if len(sm.config.Commands.Start) == 0 {
		slog.Info("Start command not defined for service manager, skipping", "serviceManager", sm.name)
		return nil
	}

	cmdArgs := sm.replaceServiceName(sm.config.Commands.Start, serviceName)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	if err := cmd.Run(); err != nil {
		slog.Error("Service start failed", "service", serviceName, "error", err)
		return err
	}

	return nil
}

// replaceServiceName replaces __SERVICE_NAME__ placeholder in command templates
func (sm *serviceManager) replaceServiceName(cmdTemplate []string, serviceName string) []string {
	result := make([]string, len(cmdTemplate))
	for i, arg := range cmdTemplate {
		result[i] = strings.ReplaceAll(arg, "__SERVICE_NAME__", serviceName)
	}
	return result
}
