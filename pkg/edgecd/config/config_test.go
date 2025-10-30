package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigValue(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		yamlValue    string
		defaultValue string
		want         string
	}{
		{
			name:         "env variable takes precedence",
			envVar:       "TEST_VAR_1",
			envValue:     "from-env",
			yamlValue:    "from-yaml",
			defaultValue: "from-default",
			want:         "from-env",
		},
		{
			name:         "yaml value when no env",
			envVar:       "TEST_VAR_2",
			envValue:     "",
			yamlValue:    "from-yaml",
			defaultValue: "from-default",
			want:         "from-yaml",
		},
		{
			name:         "default value when no env or yaml",
			envVar:       "TEST_VAR_3",
			envValue:     "",
			yamlValue:    "",
			defaultValue: "from-default",
			want:         "from-default",
		},
		{
			name:         "empty string for all",
			envVar:       "TEST_VAR_4",
			envValue:     "",
			yamlValue:    "",
			defaultValue: "",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				os.Unsetenv(tt.envVar)
			}

			got := getConfigValue(tt.envVar, tt.yamlValue, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getConfigValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig_MissingConfigPath(t *testing.T) {
	// Ensure CONFIG_PATH is not set
	os.Unsetenv("CONFIG_PATH")

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when CONFIG_PATH not set, got nil")
	}

	if err != nil && err.Error() != "CONFIG_PATH environment variable must be set" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Set CONFIG_PATH to non-existent directory
	os.Setenv("CONFIG_PATH", "nonexistent")
	defer os.Unsetenv("CONFIG_PATH")

	os.Setenv("CONFIG_REPO_DEST_PATH", "/tmp")
	defer os.Unsetenv("CONFIG_REPO_DEST_PATH")

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when config file not found, got nil")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temp directory with invalid YAML
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "test-config")
	os.MkdirAll(configDir, 0755)

	configFile := filepath.Join(configDir, "spec.yaml")
	os.WriteFile(configFile, []byte("invalid: yaml: content: ["), 0644)

	os.Setenv("CONFIG_PATH", "test-config")
	defer os.Unsetenv("CONFIG_PATH")

	os.Setenv("CONFIG_REPO_DEST_PATH", tempDir)
	defer os.Unsetenv("CONFIG_REPO_DEST_PATH")

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_Success(t *testing.T) {
	// Create temp directory with valid config
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "test-device")
	os.MkdirAll(configDir, 0755)

	validConfig := `
edgeCD:
  repo:
    url: https://github.com/test/edge-cd.git
    branch: main
    destinationPath: /opt/edge-cd

config:
  spec: spec.yaml
  path: test-device
  repo:
    url: https://github.com/test/config.git
    branch: main
    destPath: /opt/config

pollingIntervalSecond: 30

serviceManager:
  name: systemd

packageManager:
  name: apt
  autoUpgrade: true
  requiredPackages:
    - git
    - curl
`

	configFile := filepath.Join(configDir, "spec.yaml")
	os.WriteFile(configFile, []byte(validConfig), 0644)

	os.Setenv("CONFIG_PATH", "test-device")
	defer os.Unsetenv("CONFIG_PATH")

	os.Setenv("CONFIG_REPO_DEST_PATH", tempDir)
	defer os.Unsetenv("CONFIG_REPO_DEST_PATH")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify basic fields
	if cfg.Spec == nil {
		t.Fatal("Spec is nil")
	}

	if cfg.Spec.EdgeCD.Repo.URL != "https://github.com/test/edge-cd.git" {
		t.Errorf("EdgeCD URL = %v, want %v", cfg.Spec.EdgeCD.Repo.URL, "https://github.com/test/edge-cd.git")
	}

	if cfg.Spec.ServiceManager.Name != "systemd" {
		t.Errorf("ServiceManager = %v, want systemd", cfg.Spec.ServiceManager.Name)
	}

	if cfg.Spec.PackageManager.Name != "apt" {
		t.Errorf("PackageManager = %v, want apt", cfg.Spec.PackageManager.Name)
	}

	if cfg.Spec.PollingInterval != 30 {
		t.Errorf("PollingInterval = %v, want 30", cfg.Spec.PollingInterval)
	}

	// Verify computed paths
	if cfg.LockPath == "" {
		t.Error("LockPath not set")
	}

	if cfg.EdgeCDRepoPath == "" {
		t.Error("EdgeCDRepoPath not set")
	}

	if cfg.ConfigRepoPath != tempDir {
		t.Errorf("ConfigRepoPath = %v, want %v", cfg.ConfigRepoPath, tempDir)
	}

	if cfg.ConfigSpecPath != configFile {
		t.Errorf("ConfigSpecPath = %v, want %v", cfg.ConfigSpecPath, configFile)
	}
}

func TestLoadConfig_EnvironmentOverridesYAML(t *testing.T) {
	// Create temp directory with valid config
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "test-device")
	os.MkdirAll(configDir, 0755)

	validConfig := `
edgeCD:
  repo:
    url: https://github.com/test/edge-cd.git
    branch: main
    destinationPath: /opt/edge-cd

config:
  spec: spec.yaml
  path: test-device
  repo:
    url: https://github.com/test/config.git
    branch: main
    destPath: /opt/config

pollingIntervalSecond: 30

serviceManager:
  name: systemd

packageManager:
  name: apt
`

	configFile := filepath.Join(configDir, "spec.yaml")
	os.WriteFile(configFile, []byte(validConfig), 0644)

	os.Setenv("CONFIG_PATH", "test-device")
	defer os.Unsetenv("CONFIG_PATH")

	os.Setenv("CONFIG_REPO_DEST_PATH", tempDir)
	defer os.Unsetenv("CONFIG_REPO_DEST_PATH")

	// Set environment variable to override YAML value
	os.Setenv("EDGE_CD_REPO_DESTINATION_PATH", "/custom/edge-cd")
	defer os.Unsetenv("EDGE_CD_REPO_DESTINATION_PATH")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Environment variable should override YAML value
	if cfg.EdgeCDRepoPath != "/custom/edge-cd" {
		t.Errorf("EdgeCDRepoPath = %v, want /custom/edge-cd (env should override yaml)", cfg.EdgeCDRepoPath)
	}
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Create temp directory with minimal config (no optional fields)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "test-device")
	os.MkdirAll(configDir, 0755)

	minimalConfig := `
edgeCD:
  repo:
    url: https://github.com/test/edge-cd.git
    branch: main
    destinationPath: /opt/edge-cd

config:
  spec: spec.yaml
  path: test-device
  repo:
    url: https://github.com/test/config.git
    branch: main
    destPath: /opt/config

serviceManager:
  name: systemd

packageManager:
  name: apt
`

	configFile := filepath.Join(configDir, "spec.yaml")
	os.WriteFile(configFile, []byte(minimalConfig), 0644)

	os.Setenv("CONFIG_PATH", "test-device")
	defer os.Unsetenv("CONFIG_PATH")

	os.Setenv("CONFIG_REPO_DEST_PATH", tempDir)
	defer os.Unsetenv("CONFIG_REPO_DEST_PATH")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// YAML value should be used when env is not set
	if cfg.EdgeCDRepoPath != "/opt/edge-cd" {
		t.Errorf("EdgeCDRepoPath = %v, want /opt/edge-cd (from YAML)", cfg.EdgeCDRepoPath)
	}

	// Default values should be applied for optional fields not in YAML
	if cfg.LockPath != "/tmp/edge-cd/edge-cd.lock" {
		t.Errorf("LockPath = %v, want /tmp/edge-cd/edge-cd.lock (default)", cfg.LockPath)
	}

	if cfg.EdgeCDCommitPath != "/tmp/edge-cd/edge-cd-last-synchronized-commit.txt" {
		t.Errorf("EdgeCDCommitPath = %v, want default", cfg.EdgeCDCommitPath)
	}

	if cfg.ConfigCommitPath != "/tmp/edge-cd/config-last-synchronized-commit.txt" {
		t.Errorf("ConfigCommitPath = %v, want default", cfg.ConfigCommitPath)
	}
}
