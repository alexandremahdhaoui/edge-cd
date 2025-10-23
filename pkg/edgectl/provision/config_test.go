package provision_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
)

func TestReadLocalConfig(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "read-local-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := "hello: world"
	configSpec := "config.yaml"

	if err := ioutil.WriteFile(filepath.Join(tmpDir, configSpec), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	t.Run("should read local config file", func(t *testing.T) {
		content, err := provision.ReadLocalConfig(tmpDir, configSpec)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if content != configContent {
			t.Errorf("expected content '%s', got '%s'", configContent, content)
		}
	})

	t.Run("should return an error if file does not exist", func(t *testing.T) {
		_, err := provision.ReadLocalConfig(tmpDir, "nonexistent.yaml")
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}