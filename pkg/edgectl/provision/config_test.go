package provision

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

func TestRenderConfigYAML(t *testing.T) {
	data := ConfigTemplateData{
		EdgeCDRepoURL:      "https://github.com/example/edge-cd.git",
		ConfigRepoURL:      "https://github.com/example/config.git",
		ServiceManagerName: "systemd",
		PackageManagerName: "apt",
		RequiredPackages:   []string{"git", "curl"},
	}

	expectedYAML := `
# -- defines how EdgeCD clone itself
edgectl:
  autoUpdate:
    enabled: true
  repo:
    url: "https://github.com/example/edge-cd.git"
    branch: "main" # Assuming default branch for now
    destinationPath: "/usr/local/src/edge-cd" # Assuming default path for now

config:
  path: "./devices/${HOSTNAME}"
  filename: "config.yaml"
  repo:
    url: "https://github.com/example/config.git"
    branch: "main" # Assuming default branch for now
    destinationPath: "/usr/local/src/deployment" # Assuming default path for now

pollingIntervalSecond: 60

extraEnvs:
  - HOME: /root
  - GIT_SSH_COMMAND: "ssh -o StrictHostKeyChecking=accept-new"

serviceManager:
  name: "systemd"

packageManager:
  name: "apt"
  autoUpgrade: false
  requiredPackages:
    - git
    - curl

# -- Sync directories (placeholder for now)
directories: []

# -- Sync single files (placeholder for now)
files: []
`

	renderedYAML, err := RenderConfig(data)
	if err != nil {
		t.Fatalf("RenderConfig failed: %v", err)
	}

	// Normalize whitespace for comparison
	expectedYAML = strings.TrimSpace(expectedYAML)
	renderedYAML = strings.TrimSpace(renderedYAML)

	if renderedYAML != expectedYAML {
		t.Errorf("Rendered YAML does not match expected.\nExpected:\n%s\nGot:\n%s", expectedYAML, renderedYAML)
	}
}

func TestPlaceConfigYAMLRemote(t *testing.T) {
	mockContent := "test: \n  key: value\n"
	mockDestPath := "/tmp/test-config.yaml"

	expectedCmd := fmt.Sprintf("printf %%s '%s' > %s", mockContent, mockDestPath)

	mockRunner := ssh.NewMockRunner()

	err := PlaceConfigYAML(mockRunner, mockContent, mockDestPath)
	if err != nil {
		t.Fatalf("PlaceConfigYAML failed: %v", err)
	}

	if err := mockRunner.AssertCommandRun(expectedCmd); err != nil {
		t.Errorf("Expected command not run: %v", err)
	}

	if err := mockRunner.AssertNumberOfCommandsRun(1); err != nil {
		t.Errorf("Expected 1 command to be run, but got different: %v", err)
	}
}