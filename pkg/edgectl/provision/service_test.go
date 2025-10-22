package provision

import (
	"fmt"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

func TestSetupEdgeCDService(t *testing.T) {
	tests := []struct {
		name             string
		serviceManager   string
		edgeCDRepoPath   string
		expectedCommands []string
	}{
		{
			name:           "systemd service setup",
			serviceManager: "systemd",
			edgeCDRepoPath: "/tmp/edge-cd-repo",
			expectedCommands: []string{
				fmt.Sprintf("test -L %s && readlink %s | grep -q '/dev/null'", "/etc/systemd/system/edge-cd.service", "/etc/systemd/system/edge-cd.service"),
				fmt.Sprintf("rm %s", "/etc/systemd/system/edge-cd.service"),
				"systemctl daemon-reload",
				"systemctl unmask edge-cd.service",
				"cp /tmp/edge-cd-repo/cmd/edge-cd/service-managers/systemd/service /etc/systemd/system/edge-cd.service",
				"systemctl enable edge-cd.service",
				"systemctl start edge-cd.service",
			},
		},
		{
			name:           "procd service setup",
			serviceManager: "procd",
			edgeCDRepoPath: "/tmp/edge-cd-repo",
			expectedCommands: []string{
				"cp /tmp/edge-cd-repo/cmd/edge-cd/service-managers/procd/service /etc/init.d/edge-cd",
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

			err := SetupEdgeCDService(mockRunner, tt.serviceManager, tt.edgeCDRepoPath)
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