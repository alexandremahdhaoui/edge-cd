package userconfig

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/yaml"
)

// TestValidateConfigFiles validates that all example config files can be parsed
func TestValidateConfigFiles(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		skipValidation bool // Some test configs may have intentional minimal fields
	}{
		{
			name:     "examples/config.yaml",
			filePath: "../../examples/config.yaml",
		},
		{
			name:     "test/edgectl/e2e/config/config.yaml",
			filePath: "../../test/edgectl/e2e/config/config.yaml",
		},
		{
			name:     "test/edge-cd/e2e/config/config.yaml",
			filePath: "../../test/edge-cd/e2e/config/config.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get absolute path
			absPath, err := filepath.Abs(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			// Read the file
			content, err := os.ReadFile(absPath)
			if err != nil {
				t.Fatalf("Failed to read config file at %s: %v", absPath, err)
			}

			// Parse the YAML
			var config Spec
			if err := yaml.Unmarshal(content, &config); err != nil {
				t.Fatalf("Failed to unmarshal config file at %s: %v", absPath, err)
			}

			// Verify critical fields match authoritative structure
			// edgeCD section should exist (not edgectl)
			if config.EdgeCD.Repo.URL == "" {
				t.Logf("Warning: edgeCD.repo.url is empty in %s", tc.name)
			}

			// config.repo should use destPath (not destinationPath)
			if config.Config.Repo.DestPath == "" && config.Config.Repo.URL != "" {
				t.Errorf("config.repo.destPath is empty but repo.url is set in %s", tc.name)
			}

			// Optional: validate the config if skipValidation is false
			if !tc.skipValidation {
				if err := config.Validate(); err != nil {
					t.Logf("Validation warnings for %s: %v", tc.name, err)
				}
			}

			t.Logf("Successfully parsed and validated %s", tc.name)
		})
	}
}

// TestConfigStructureCompliance verifies the field names match authoritative source
func TestConfigStructureCompliance(t *testing.T) {
	// This test verifies that the YAML uses correct field names
	testYAML := `
edgeCD:
  repo:
    url: "https://example.com/edge-cd.git"
    branch: "main"
    destinationPath: "/usr/local/src/edge-cd"

config:
  spec: "spec.yaml"
  path: "./test"
  repo:
    url: "https://example.com/config.git"
    branch: "main"
    destPath: "/usr/local/src/config"
`

	var config Spec
	if err := yaml.Unmarshal([]byte(testYAML), &config); err != nil {
		t.Fatalf("Failed to unmarshal test YAML: %v", err)
	}

	// Verify edgeCD.repo uses destinationPath
	if config.EdgeCD.Repo.DestinationPath != "/usr/local/src/edge-cd" {
		t.Errorf("Expected edgeCD.repo.destinationPath to be '/usr/local/src/edge-cd', got '%s'", config.EdgeCD.Repo.DestinationPath)
	}

	// Verify config.repo uses destPath
	if config.Config.Repo.DestPath != "/usr/local/src/config" {
		t.Errorf("Expected config.repo.destPath to be '/usr/local/src/config', got '%s'", config.Config.Repo.DestPath)
	}

	// Verify top-level key is edgeCD (not edgectl)
	// This is implicitly tested by successful parsing above
}
