package provision_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

func TestProvisionPackages(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "prov-pkg-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	aptYaml := `
update: ["apt-get", "update"]
install: ["apt-get", "install", "-y"]
`
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "apt.yaml"), []byte(aptYaml), 0644); err != nil {
		t.Fatalf("Failed to write apt.yaml: %v", err)
	}

	t.Run("should install multiple packages with apt", func(t *testing.T) {
		mock := &ssh.MockRunner{
			RunResponses: []ssh.RunResponse{
				{Err: nil}, // update
				{Err: nil}, // install
			},
		}
		packages := []string{"git", "curl"}
		if err := provision.ProvisionPackages(mock, packages, "apt", tmpDir); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			"apt-get update",
			"apt-get install -y git curl",
		}

		if len(mock.RunCommands) != len(expectedCommands) {
			t.Fatalf("expected %d commands, got %d", len(expectedCommands), len(mock.RunCommands))
		}

		for i, cmd := range mock.RunCommands {
			if cmd != expectedCommands[i] {
				t.Errorf("expected command '%s' at index %d, got '%s'", expectedCommands[i], i, cmd)
			}
		}
	})

	t.Run("should do nothing if no packages are provided", func(t *testing.T) {
		mock := &ssh.MockRunner{
			RunResponses: []ssh.RunResponse{
				{Err: nil}, // update
			},
		}
		var packages []string
		if err := provision.ProvisionPackages(mock, packages, "apt", tmpDir); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expectedCommands := []string{
			"apt-get update",
		}

		if len(mock.RunCommands) != len(expectedCommands) {
			t.Fatalf("expected %d commands, got %d", len(expectedCommands), len(mock.RunCommands))
		}
	})
}
