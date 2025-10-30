package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexandremahdhaoui/edge-cd/pkg/userconfig"
	"gopkg.in/yaml.v3"
)

// Config holds the complete edge-cd configuration with computed paths.
type Config struct {
	// Parsed YAML specification
	Spec *userconfig.Spec

	// Computed runtime paths
	LockPath         string
	EdgeCDRepoPath   string
	EdgeCDCommitPath string
	ConfigRepoPath   string
	ConfigCommitPath string
	ConfigSpecPath   string
}

// LoadConfig reads configuration from environment variables and YAML file.
// It applies precedence rules: env > yaml > default.
//
// Required environment variables:
//   - CONFIG_PATH: Path within config repository
//
// Returns error if CONFIG_PATH is not set or if configuration is invalid.
func LoadConfig() (*Config, error) {
	// CONFIG_PATH is required
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, fmt.Errorf("CONFIG_PATH environment variable must be set")
	}

	// Read other values with precedence: env > yaml > default
	configSpecFile := getConfigValue("CONFIG_SPEC_FILE", "", "spec.yaml")
	configRepoDestPath := getConfigValue("CONFIG_REPO_DEST_PATH", "", "/usr/local/src/edge-cd-config")

	// Build config spec path
	configSpecPath := filepath.Join(configRepoDestPath, configPath, configSpecFile)

	// Parse YAML using userconfig.Spec
	data, err := os.ReadFile(configSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configSpecPath, err)
	}

	var spec userconfig.Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Build Config struct with computed values
	cfg := &Config{
		Spec:             &spec,
		LockPath:         filepath.Join(getConfigValue("LOCK_FILE_DIRNAME", "", "/tmp/edge-cd"), "edge-cd.lock"),
		EdgeCDRepoPath:   getConfigValue("EDGE_CD_REPO_DESTINATION_PATH", spec.EdgeCD.Repo.DestinationPath, "/usr/local/src/edge-cd"),
		EdgeCDCommitPath: getConfigValue("EDGE_CD_COMMIT_PATH", spec.EdgeCD.CommitPath, "/tmp/edge-cd/edge-cd-last-synchronized-commit.txt"),
		ConfigRepoPath:   configRepoDestPath,
		ConfigCommitPath: getConfigValue("CONFIG_COMMIT_PATH", spec.Config.CommitPath, "/tmp/edge-cd/config-last-synchronized-commit.txt"),
		ConfigSpecPath:   configSpecPath,
	}

	return cfg, nil
}

// getConfigValue reads a value with precedence: env > yaml > default.
//
// Parameters:
//   - envVar: Environment variable name to check first
//   - yamlValue: Value from YAML configuration (checked second)
//   - defaultValue: Default value if neither env nor yaml provide a value
//
// Returns the first non-empty value according to precedence.
func getConfigValue(envVar, yamlValue, defaultValue string) string {
	// 1. Environment variable (highest precedence)
	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}

	// 2. YAML value (middle precedence)
	if yamlValue != "" {
		return yamlValue
	}

	// 3. Default value (lowest precedence)
	return defaultValue
}
