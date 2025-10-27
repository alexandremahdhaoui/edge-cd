package provision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

func TestSetupEdgeCDService(t *testing.T) {
	// Find the actual edge-cd repository path
	repoPath, err := findEdgeCDRepoPath()
	if err != nil {
		t.Skipf("Skipping test: could not find edge-cd repository: %v", err)
	}

	tests := []struct {
		name             string
		serviceManager   string
		prependCmd       []string
		minExpectedCmds  int // Minimum expected commands (mkdir, sh -c, chmod, enable, start/restart)
	}{
		{
			name:            "systemd service setup without prepend command",
			serviceManager:  "systemd",
			prependCmd:      []string{},
			minExpectedCmds: 5, // mkdir, sh -c (base64), chmod, enable, start
		},
		{
			name:            "systemd service setup with sudo prepend command",
			serviceManager:  "systemd",
			prependCmd:      []string{"sudo", "-E"},
			minExpectedCmds: 5, // mkdir, sh -c (base64), chmod, enable, start
		},
		{
			name:            "procd service setup without prepend command",
			serviceManager:  "procd",
			prependCmd:      []string{},
			minExpectedCmds: 5, // mkdir, sh -c (base64), chmod, enable, restart
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := ssh.NewMockRunner()

			// Create context with prepend command if provided
			envs := make(map[string]string)
			ctx := execcontext.New(envs, tt.prependCmd)

			// Build template data for testing
			templateData := ServiceTemplateData{
				EdgeCDScriptPath: filepath.Join(repoPath, "cmd/edge-cd/edge-cd"),
				ConfigPath:       "/etc/edge-cd/config.yaml",
				User:             "",
				Group:            "",
				EnvironmentVars:  []EnvVar{},
				Args:             []string{},
			}

			err := SetupEdgeCDService(ctx, mockRunner, tt.serviceManager, repoPath, repoPath, templateData)
			if err != nil {
				t.Fatalf("SetupEdgeCDService failed: %v", err)
			}

			// Verify that we got at least the minimum expected commands
			if len(mockRunner.Commands) < tt.minExpectedCmds {
				t.Fatalf("Expected at least %d commands to be run, but got %d. Commands: %v", tt.minExpectedCmds, len(mockRunner.Commands), mockRunner.Commands)
			}

			// Verify key commands are present (enable and start/restart)
			hasEnable := false
			hasStartOrRestart := false
			for _, cmd := range mockRunner.Commands {
				if strings.Contains(cmd, "enable") {
					hasEnable = true
				}
				if strings.Contains(cmd, "start") || strings.Contains(cmd, "restart") {
					hasStartOrRestart = true
				}
			}

			if !hasEnable {
				t.Errorf("Expected 'enable' command but did not find it. Commands: %v", mockRunner.Commands)
			}
			if !hasStartOrRestart {
				t.Errorf("Expected 'start' or 'restart' command but did not find it. Commands: %v", mockRunner.Commands)
			}
		})
	}
}

// findEdgeCDRepoPath finds the edge-cd repository root by looking for the cmd/edge-cd directory
func findEdgeCDRepoPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up from current directory to find the repo root
	for {
		// Check if cmd/edge-cd exists
		if _, err := os.Stat(filepath.Join(cwd, "cmd", "edge-cd")); err == nil {
			return cwd, nil
		}

		// Move up one directory
		parent := filepath.Dir(cwd)
		if parent == cwd {
			// Reached the root
			return "", fmt.Errorf("could not find edge-cd repository root")
		}
		cwd = parent
	}
}
