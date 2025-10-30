package svcmgr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewServiceManager_Systemd(t *testing.T) {
	repoRoot := findRepoRoot(t)

	sm, err := NewServiceManager("systemd", repoRoot)
	if err != nil {
		t.Fatalf("NewServiceManager(systemd) failed: %v", err)
	}

	if sm == nil {
		t.Fatal("Expected non-nil ServiceManager")
	}

	// Check that config was loaded
	concrete, ok := sm.(*serviceManager)
	if !ok {
		t.Fatal("Expected *serviceManager")
	}

	if concrete.name != "systemd" {
		t.Errorf("Expected name=systemd, got %s", concrete.name)
	}

	if len(concrete.config.Commands.Enable) == 0 {
		t.Error("Expected enable command to be loaded")
	}

	if len(concrete.config.Commands.Restart) == 0 {
		t.Error("Expected restart command to be loaded")
	}

	if len(concrete.config.Commands.Start) == 0 {
		t.Error("Expected start command to be loaded")
	}

	expectedEnable := []string{"systemctl", "enable", "__SERVICE_NAME__"}
	if !slicesEqual(concrete.config.Commands.Enable, expectedEnable) {
		t.Errorf("Expected enable=%v, got %v", expectedEnable, concrete.config.Commands.Enable)
	}

	expectedRestart := []string{"systemctl", "restart", "__SERVICE_NAME__"}
	if !slicesEqual(concrete.config.Commands.Restart, expectedRestart) {
		t.Errorf("Expected restart=%v, got %v", expectedRestart, concrete.config.Commands.Restart)
	}
}

func TestNewServiceManager_Procd(t *testing.T) {
	repoRoot := findRepoRoot(t)

	sm, err := NewServiceManager("procd", repoRoot)
	if err != nil {
		t.Fatalf("NewServiceManager(procd) failed: %v", err)
	}

	if sm == nil {
		t.Fatal("Expected non-nil ServiceManager")
	}

	// Check that config was loaded
	concrete, ok := sm.(*serviceManager)
	if !ok {
		t.Fatal("Expected *serviceManager")
	}

	if concrete.name != "procd" {
		t.Errorf("Expected name=procd, got %s", concrete.name)
	}

	if len(concrete.config.Commands.Enable) == 0 {
		t.Error("Expected enable command to be loaded")
	}

	if len(concrete.config.Commands.Restart) == 0 {
		t.Error("Expected restart command to be loaded")
	}

	expectedEnable := []string{"/etc/init.d/__SERVICE_NAME__", "enable"}
	if !slicesEqual(concrete.config.Commands.Enable, expectedEnable) {
		t.Errorf("Expected enable=%v, got %v", expectedEnable, concrete.config.Commands.Enable)
	}

	expectedRestart := []string{"/etc/init.d/__SERVICE_NAME__", "restart"}
	if !slicesEqual(concrete.config.Commands.Restart, expectedRestart) {
		t.Errorf("Expected restart=%v, got %v", expectedRestart, concrete.config.Commands.Restart)
	}

	expectedStart := []string{"/etc/init.d/__SERVICE_NAME__", "start"}
	if !slicesEqual(concrete.config.Commands.Start, expectedStart) {
		t.Errorf("Expected start=%v, got %v", expectedStart, concrete.config.Commands.Start)
	}
}

func TestNewServiceManager_UnknownManager(t *testing.T) {
	repoRoot := findRepoRoot(t)

	_, err := NewServiceManager("nonexistent", repoRoot)
	if err == nil {
		t.Fatal("Expected error for unknown service manager, got nil")
	}
}

func TestReplaceServiceName(t *testing.T) {
	sm := &serviceManager{
		name: "test",
	}

	tests := []struct {
		name         string
		cmdTemplate  []string
		serviceName  string
		expected     []string
	}{
		{
			name:        "Single placeholder",
			cmdTemplate: []string{"systemctl", "restart", "__SERVICE_NAME__"},
			serviceName: "nginx",
			expected:    []string{"systemctl", "restart", "nginx"},
		},
		{
			name:        "Multiple placeholders",
			cmdTemplate: []string{"systemctl", "__SERVICE_NAME__", "start", "__SERVICE_NAME__"},
			serviceName: "apache2",
			expected:    []string{"systemctl", "apache2", "start", "apache2"},
		},
		{
			name:        "No placeholder",
			cmdTemplate: []string{"systemctl", "daemon-reload"},
			serviceName: "nginx",
			expected:    []string{"systemctl", "daemon-reload"},
		},
		{
			name:        "Empty template",
			cmdTemplate: []string{},
			serviceName: "nginx",
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.replaceServiceName(tt.cmdTemplate, tt.serviceName)
			if !slicesEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}


// Helper functions

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find repository root (no go.mod found)")
		}
		dir = parent
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
