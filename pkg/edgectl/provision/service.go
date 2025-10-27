package provision

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
	"gopkg.in/yaml.v3"
)

var (
	errLoadServiceManagerConfig     = errors.New("failed to load service manager config")
	errCopyServiceFile              = errors.New("failed to copy service file")
	errEnableService                = errors.New("failed to enable service")
	errStartService                 = errors.New("failed to start service")
	errReadServiceManagerConfigFile = errors.New("failed to read service manager config file")
	errParseServiceManagerConfig    = errors.New("failed to parse service manager config")
	errReadServiceTemplate          = errors.New("failed to read service template file")
	errParseServiceTemplate         = errors.New("failed to parse service template")
	errRenderServiceTemplate        = errors.New("failed to render service template")
	errPlaceServiceFile             = errors.New("failed to place service file")
)

// ServiceManagerConfig represents the structure of service manager config files
type ServiceManagerConfig struct {
	Commands      map[string][]string `yaml:"commands"`
	EdgeCDService struct {
		DestinationPath string `yaml:"destinationPath"`
	} `yaml:"edgeCDService"`
}

// ServiceTemplateData holds the data for rendering service file templates
type ServiceTemplateData struct {
	EdgeCDScriptPath string
	ConfigPath       string
	User             string
	Group            string
	EnvironmentVars  []EnvVar
	Args             []string
}

// EnvVar represents an environment variable key-value pair
type EnvVar struct {
	Key   string
	Value string
}

// SetupEdgeCDService sets up and enables the edge-cd service on the remote device.
// The context parameter should contain any required prepend commands (e.g., "sudo").
// localEdgeCDRepoPath is used to read config files and templates locally.
// remoteEdgeCDRepoPath is the path to the edge-cd repo on the remote target VM.
func SetupEdgeCDService(
	execCtx execcontext.Context,
	runner ssh.Runner,
	svcmgrName string,
	localEdgeCDRepoPath string,
	remoteEdgeCDRepoPath string,
	templateData ServiceTemplateData,
) error {
	var stdout, stderr string
	// Load the service manager configuration from LOCAL repo
	config, err := loadServiceManagerConfig(localEdgeCDRepoPath, svcmgrName)
	if err != nil {
		return err
	}

	// Render service file template
	slog.Info("rendering service file template", "serviceManager", svcmgrName)
	serviceContent, err := RenderServiceFile(localEdgeCDRepoPath, svcmgrName, templateData)
	if err != nil {
		return err
	}

	// Place rendered service file on remote device
	serviceDestPath := config.EdgeCDService.DestinationPath
	slog.Info("placing service file", "dest", serviceDestPath)
	if err := PlaceServiceFile(execCtx, runner, serviceContent, serviceDestPath); err != nil {
		return err
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

// RenderServiceFile renders a service file template with the provided data
func RenderServiceFile(
	localEdgeCDRepoPath string,
	svcmgrName string,
	data ServiceTemplateData,
) (string, error) {
	templatePath := filepath.Join(
		localEdgeCDRepoPath,
		"cmd/edge-cd/service-managers",
		svcmgrName,
		"service.gotpl",
	)

	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", flaterrors.Join(err, fmt.Errorf("templatePath=%s", templatePath), errReadServiceTemplate)
	}

	tmpl, err := template.New("service").Parse(string(templateContent))
	if err != nil {
		return "", flaterrors.Join(err, errParseServiceTemplate)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", flaterrors.Join(err, errRenderServiceTemplate)
	}

	return buf.String(), nil
}

// PlaceServiceFile transfers the rendered service file content to the remote device
func PlaceServiceFile(
	execCtx execcontext.Context,
	runner ssh.Runner,
	content, destPath string,
) error {
	// Create the directory first
	dirPath := filepath.Dir(destPath)
	stdout, stderr, err := runner.Run(execCtx, "mkdir", "-p", dirPath)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("dirPath=%s stdout=%s stderr=%s", dirPath, stdout, stderr), errPlaceServiceFile)
	}

	// Use base64 encoding to safely transfer content
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	shellCmd := fmt.Sprintf("echo %s | base64 -d > %s", encoded, destPath)
	stdout, stderr, err = runner.Run(execCtx, "sh", "-c", shellCmd)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", destPath, stdout, stderr), errPlaceServiceFile)
	}

	// Make service file executable (needed for procd init scripts)
	stdout, stderr, err = runner.Run(execCtx, "chmod", "755", destPath)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", destPath, stdout, stderr), errPlaceServiceFile)
	}

	return nil
}
