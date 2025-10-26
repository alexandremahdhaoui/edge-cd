package provision

import (
	"fmt"
	"os"
	"path/filepath"
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
		expectedCommands []string
	}{
		{
			name:           "systemd service setup without prepend command",
			serviceManager: "systemd",
			prependCmd:     []string{},
			expectedCommands: []string{
				fmt.Sprintf("cp %s /etc/systemd/system/edge-cd.service", filepath.Join(repoPath, "cmd/edge-cd/service-managers/systemd/edge-cd.systemd")),
				"systemctl enable edge-cd",
				"systemctl start edge-cd",
			},
		},
		{
			name:           "systemd service setup with sudo prepend command",
			serviceManager: "systemd",
			prependCmd:     []string{"sudo", "-E"},
			expectedCommands: []string{
				fmt.Sprintf("sudo -E cp %s /etc/systemd/system/edge-cd.service", filepath.Join(repoPath, "cmd/edge-cd/service-managers/systemd/edge-cd.systemd")),
				"sudo -E systemctl enable edge-cd",
				"sudo -E systemctl start edge-cd",
			},
		},
		{
			name:           "procd service setup without prepend command",
			serviceManager: "procd",
			prependCmd:     []string{},
			expectedCommands: []string{
				fmt.Sprintf("cp %s /etc/init.d/edge-cd", filepath.Join(repoPath, "cmd/edge-cd/service-managers/procd/edge-cd.procd")),
				"/etc/init.d/edge-cd enable",
				"service edge-cd restart",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := ssh.NewMockRunner()

			// Create context with prepend command if provided
			envs := make(map[string]string)
			ctx := execcontext.New(envs, tt.prependCmd)

			err := SetupEdgeCDService(ctx, mockRunner, tt.serviceManager, repoPath)
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
