package pkgmgr

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewPackageManager_Opkg(t *testing.T) {
	// Tests run from package directory, so repo root is ../../../
	pm, err := NewPackageManager("opkg", "../../..")
	if err != nil {
		t.Fatalf("Failed to create opkg package manager: %v", err)
	}

	if pm == nil {
		t.Fatal("Expected package manager instance, got nil")
	}

	// Verify config was loaded (check internal state via type assertion)
	concrete, ok := pm.(*packageManager)
	if !ok {
		t.Fatal("Expected *packageManager type")
	}

	if concrete.name != "opkg" {
		t.Errorf("Expected name 'opkg', got '%s'", concrete.name)
	}

	if len(concrete.config.Update) == 0 {
		t.Error("Expected update command to be loaded")
	}
	if len(concrete.config.Install) == 0 {
		t.Error("Expected install command to be loaded")
	}
	if len(concrete.config.Upgrade) == 0 {
		t.Error("Expected upgrade command to be loaded")
	}
}

func TestNewPackageManager_Apt(t *testing.T) {
	// Tests run from package directory, so repo root is ../../../
	pm, err := NewPackageManager("apt", "../../..")
	if err != nil {
		t.Fatalf("Failed to create apt package manager: %v", err)
	}

	if pm == nil {
		t.Fatal("Expected package manager instance, got nil")
	}

	// Verify config was loaded
	concrete, ok := pm.(*packageManager)
	if !ok {
		t.Fatal("Expected *packageManager type")
	}

	if concrete.name != "apt" {
		t.Errorf("Expected name 'apt', got '%s'", concrete.name)
	}

	if len(concrete.config.Update) == 0 {
		t.Error("Expected update command to be loaded")
	}
	if len(concrete.config.Install) == 0 {
		t.Error("Expected install command to be loaded")
	}
	if len(concrete.config.Upgrade) == 0 {
		t.Error("Expected upgrade command to be loaded")
	}

	// Verify apt has expected command structure (sudo apt-get ...)
	if concrete.config.Update[0] != "sudo" {
		t.Errorf("Expected first update arg to be 'sudo', got '%s'", concrete.config.Update[0])
	}
}

func TestNewPackageManager_UnknownManager(t *testing.T) {
	_, err := NewPackageManager("nonexistent", "../../..")
	if err == nil {
		t.Fatal("Expected error for nonexistent package manager, got nil")
	}
}

func TestInstall_EmptyPackageList(t *testing.T) {
	// Create a mock package manager with test config
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"echo", "update"},
			Install: []string{"echo", "install"},
			Upgrade: []string{"echo", "upgrade"},
		},
	}

	// Should return nil without error for empty list
	err := pm.Install([]string{})
	if err != nil {
		t.Errorf("Expected no error for empty package list, got: %v", err)
	}
}

func TestUpgrade_EmptyPackageList(t *testing.T) {
	// Create a mock package manager with test config
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"echo", "update"},
			Install: []string{"echo", "install"},
			Upgrade: []string{"echo", "upgrade"},
		},
	}

	// Should return nil without error for empty list
	err := pm.Upgrade([]string{})
	if err != nil {
		t.Errorf("Expected no error for empty package list, got: %v", err)
	}
}

func TestUpdate_ExecutesCommand(t *testing.T) {
	// Create a package manager that uses 'true' command (always succeeds)
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update: []string{"true"},
		},
	}

	err := pm.Update()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUpdate_CommandFailure(t *testing.T) {
	// Create a package manager that uses 'false' command (always fails)
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update: []string{"false"},
		},
	}

	err := pm.Update()
	if err == nil {
		t.Error("Expected error for failed update command, got nil")
	}
}

func TestInstall_ExecutesCommand(t *testing.T) {
	// Create a package manager using echo command for testing
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"true"},
			Install: []string{"echo", "install"},
		},
	}

	err := pm.Install([]string{"pkg1", "pkg2"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestInstall_UpdateFailurePropagates(t *testing.T) {
	// Create a package manager where update fails
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"false"},
			Install: []string{"echo", "install"},
		},
	}

	err := pm.Install([]string{"pkg1"})
	if err == nil {
		t.Error("Expected error when update fails, got nil")
	}
}

func TestUpgrade_ExecutesCommand(t *testing.T) {
	// Create a package manager using echo command for testing
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"true"},
			Upgrade: []string{"echo", "upgrade"},
		},
	}

	err := pm.Upgrade([]string{"pkg1", "pkg2"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUpgrade_UpdateFailurePropagates(t *testing.T) {
	// Create a package manager where update fails
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"false"},
			Upgrade: []string{"echo", "upgrade"},
		},
	}

	err := pm.Upgrade([]string{"pkg1"})
	if err == nil {
		t.Error("Expected error when update fails, got nil")
	}
}

func TestUpdate_MissingCommand(t *testing.T) {
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update: []string{},
		},
	}

	err := pm.Update()
	if err == nil {
		t.Error("Expected error for missing update command, got nil")
	}
}

func TestInstall_MissingCommand(t *testing.T) {
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"true"},
			Install: []string{},
		},
	}

	err := pm.Install([]string{"pkg1"})
	if err == nil {
		t.Error("Expected error for missing install command, got nil")
	}
}

func TestUpgrade_MissingCommand(t *testing.T) {
	pm := &packageManager{
		name: "test",
		config: &PackageManagerConfig{
			Update:  []string{"true"},
			Upgrade: []string{},
		},
	}

	err := pm.Upgrade([]string{"pkg1"})
	if err == nil {
		t.Error("Expected error for missing upgrade command, got nil")
	}
}

// TestConfigParsing verifies YAML parsing works correctly
func TestConfigParsing(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "pkgmgr-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `update: ["cmd1", "arg1"]
install: ["cmd2", "arg2", "arg3"]
upgrade: ["cmd3"]
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Parse it
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	var config PackageManagerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Verify parsing
	if len(config.Update) != 2 || config.Update[0] != "cmd1" || config.Update[1] != "arg1" {
		t.Errorf("Update command not parsed correctly: %v", config.Update)
	}

	if len(config.Install) != 3 || config.Install[0] != "cmd2" {
		t.Errorf("Install command not parsed correctly: %v", config.Install)
	}

	if len(config.Upgrade) != 1 || config.Upgrade[0] != "cmd3" {
		t.Errorf("Upgrade command not parsed correctly: %v", config.Upgrade)
	}
}
