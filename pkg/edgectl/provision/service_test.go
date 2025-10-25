package provision

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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
		prependCmd       string
		expectedCommands []string
	}{
		{
			name:           "systemd service setup without prepend command",
			serviceManager: "systemd",
			prependCmd:     "",
			expectedCommands: []string{
				fmt.Sprintf("test -L %s && readlink %s | grep -q '/dev/null'", "/etc/systemd/system/edge-cd.service", "/etc/systemd/system/edge-cd.service"),
				"rm /etc/systemd/system/edge-cd.service",
				"systemctl daemon-reload",
				"systemctl unmask edge-cd.service",
				fmt.Sprintf("cp %s /etc/systemd/system/edge-cd.service", filepath.Join(repoPath, "cmd/edge-cd/service-managers/systemd/service")),
				"systemctl enable edge-cd.service",
				"systemctl start edge-cd.service",
			},
		},
		{
			name:           "systemd service setup with sudo prepend command",
			serviceManager: "systemd",
			prependCmd:     "sudo",
			expectedCommands: []string{
				fmt.Sprintf("test -L %s && readlink %s | grep -q '/dev/null'", "/etc/systemd/system/edge-cd.service", "/etc/systemd/system/edge-cd.service"),
				"sudo rm /etc/systemd/system/edge-cd.service",
				"sudo systemctl daemon-reload",
				"sudo systemctl unmask edge-cd.service",
				fmt.Sprintf("sudo cp %s /etc/systemd/system/edge-cd.service", filepath.Join(repoPath, "cmd/edge-cd/service-managers/systemd/service")),
				"sudo systemctl enable edge-cd.service",
				"sudo systemctl start edge-cd.service",
			},
		},
		{
			name:           "procd service setup without prepend command",
			serviceManager: "procd",
			prependCmd:     "",
			expectedCommands: []string{
				fmt.Sprintf("cp %s /etc/init.d/edge-cd", filepath.Join(repoPath, "cmd/edge-cd/service-managers/procd/service")),
				"/etc/init.d/edge-cd enable",
				"/etc/init.d/edge-cd start",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := ssh.NewMockRunner()

			// Mock the checkMaskedSymlinkCmd to return success (no error) for systemd case
			if tt.serviceManager == "systemd" {
				mockRunner.SetResponse(
					fmt.Sprintf("test -L %s && readlink %s | grep -q '/dev/null'", "/etc/systemd/system/edge-cd.service", "/etc/systemd/system/edge-cd.service"),
					"", "", nil,
				)
			}

			err := SetupEdgeCDService(mockRunner, tt.serviceManager, repoPath, tt.prependCmd)
			if err != nil {
				t.Fatalf("SetupEdgeCDService failed: %v", err)
			}

			if len(mockRunner.Commands) != len(tt.expectedCommands) {
				t.Fatalf("Expected %d commands to be run, but got %d. Commands: %v", len(tt.expectedCommands), len(mockRunner.Commands), mockRunner.Commands)
			}

			for i, cmd := range mockRunner.Commands {
				if cmd != tt.expectedCommands[i] {
					t.Errorf("Command %d mismatch: got %q, want %q", i, cmd, tt.expectedCommands[i])
				}
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
