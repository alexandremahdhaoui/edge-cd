package provision

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
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
		return "", fmt.Errorf("failed to parse config template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render config template: %w", err)
	}

	return buf.String(), nil
}

// PlaceConfigYAML takes the rendered config content and places it on the remote device.
func PlaceConfigYAML(execCtx execcontext.Context, runner ssh.Runner, content, destPath string) error {
	// Use printf to handle newlines and special characters correctly
	cmd := fmt.Sprintf("printf %%s '%s' > %s", content, destPath)
	stdout, stderr, err := runner.Run(execCtx, cmd)
	if err != nil {
		return fmt.Errorf(
			"failed to place config.yaml at %s: %w. Stdout: %s, Stderr: %s",
			destPath,
			err,
			stdout,
			stderr,
		)
	}
	return nil
}

// ReadLocalConfig reads a configuration file from the local filesystem.
func ReadLocalConfig(configPath, configSpec string) (string, error) {
	fullPath := filepath.Join(configPath, configSpec)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read local config file %s: %w", fullPath, err)
	}
	return string(content), nil
}


