package provision

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
)

var (
	errParseConfigTemplate  = errors.New("failed to parse config template")
	errRenderConfigTemplate = errors.New("failed to render config template")
	errPlaceConfigYAML      = errors.New("failed to place config.yaml")
	errReadLocalConfig      = errors.New("failed to read local config file")
)

const configTemplate = `
# -- defines how EdgeCD clone itself
edgectl:
  autoUpdate:
    enabled: true
  repo:
    url: "{{ .EdgeCDRepoURL }}"
    branch: "main" # Assuming default branch for now
    destinationPath: "/usr/local/src/edge-cd" # Assuming default path for now

config:
  spec: "spec.yaml"
  path: "./devices/${HOSTNAME}"
  repo:
    url: "{{ .ConfigRepoURL }}"
    branch: "main" # Assuming default branch for now
    destinationPath: "/usr/local/src/deployment" # Assuming default path for now

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

	// Use sh -c to handle redirection properly
	shellCmd := fmt.Sprintf("printf '%%s' > %s", destPath)
	stdout, stderr, err = runner.Run(execCtx, "sh", "-c", shellCmd, "--", content)
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
