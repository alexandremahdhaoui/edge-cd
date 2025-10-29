package provision

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/edge-cd/pkg/userconfig"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
	"sigs.k8s.io/yaml"
)

var (
	errParseConfigTemplate  = errors.New("failed to parse config template")
	errRenderConfigTemplate = errors.New("failed to render config template")
	errPlaceConfigYAML      = errors.New("failed to place config.yaml")
	errReadLocalConfig      = errors.New("failed to read local config file")
	errUnmarshalConfig      = errors.New("failed to unmarshal config")
	errMarshalConfig        = errors.New("failed to marshal config")
)

const configTemplate = `
# -- defines how EdgeCD clone itself
edgeCD:
  autoUpdate:
    enabled: true
  repo:
    url: "{{ .EdgeCDRepoURL }}"
    branch: "main" # Assuming default branch for now
    destinationPath: "{{ .EdgeCDRepoDestPath }}"

config:
  spec: "spec.yaml"
  path: "./devices/${HOSTNAME}"
  repo:
    url: "{{ .ConfigRepoURL }}"
    branch: "main" # Assuming default branch for now
    destPath: "/usr/local/src/deployment" # Assuming default path for now

pollingIntervalSecond: 60

extraEnvs:
  - HOME: /root
  - GIT_SSH_COMMAND: "ssh -o StrictHostKeyChecking=accept-new"

serviceManager:
  name: "{{ .ServiceManagerName }}"

packageManager:
  name: "{{ .PackageManagerName }}"
  autoUpgrade: false
  requiredPackages:
{{- range .RequiredPackages }}
    - {{ . }}
{{- end }}

# -- Sync directories (placeholder for now)
directories: []

# -- Sync single files (placeholder for now)
files: []
`

// ConfigTemplateData holds the data for rendering the config.yaml template.
type ConfigTemplateData struct {
	EdgeCDRepoURL      string
	EdgeCDRepoDestPath string
	ConfigRepoURL      string
	ServiceManagerName string
	PackageManagerName string
	RequiredPackages   []string
}

// RenderConfig renders the config.yaml template with the provided data.
func RenderConfig(data ConfigTemplateData) (string, error) {
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", flaterrors.Join(err, errParseConfigTemplate)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", flaterrors.Join(err, errRenderConfigTemplate)
	}

	return buf.String(), nil
}

// PlaceConfigYAML takes the rendered config content and places it on the remote device.
func PlaceConfigYAML(
	execCtx execcontext.Context,
	runner ssh.Runner,
	content, destPath string,
) error {
	// Create the directory first
	dirPath := filepath.Dir(destPath)
	stdout, stderr, err := runner.Run(execCtx, "mkdir", "-p", dirPath)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("dirPath=%s stdout=%s stderr=%s", dirPath, stdout, stderr), errPlaceConfigYAML)
	}

	// Use base64 encoding to safely transfer content (avoids all escaping issues)
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	shellCmd := fmt.Sprintf("echo %s | base64 -d > %s", encoded, destPath)
	stdout, stderr, err = runner.Run(execCtx, "sh", "-c", shellCmd)
	if err != nil {
		return flaterrors.Join(err, fmt.Errorf("destPath=%s stdout=%s stderr=%s", destPath, stdout, stderr), errPlaceConfigYAML)
	}
	return nil
}

// ReadLocalConfig reads a configuration file from the local filesystem.
func ReadLocalConfig(configPath, configSpec string) (string, error) {
	fullPath := filepath.Join(configPath, configSpec)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", flaterrors.Join(err, fmt.Errorf("fullPath=%s", fullPath), errReadLocalConfig)
	}
	return string(content), nil
}

// ReplaceRepoURLsInConfig replaces the repository URLs in a config YAML string.
// This is used when a static config file is provided but dynamic repo URLs need to be injected.
func ReplaceRepoURLsInConfig(configContent, edgeCDRepoURL, configRepoURL string) (string, error) {
	// Parse the YAML into the userconfig struct
	var config userconfig.Spec
	if err := yaml.Unmarshal([]byte(configContent), &config); err != nil {
		return "", flaterrors.Join(err, errUnmarshalConfig)
	}

	// Replace edgeCD repo URL if provided
	if edgeCDRepoURL != "" {
		config.EdgeCD.Repo.URL = edgeCDRepoURL
	}

	// Replace config repo URL if provided
	if configRepoURL != "" {
		config.Config.Repo.URL = configRepoURL
	}

	// Marshal back to YAML
	updatedContent, err := yaml.Marshal(config)
	if err != nil {
		return "", flaterrors.Join(err, errMarshalConfig)
	}

	return string(updatedContent), nil
}
