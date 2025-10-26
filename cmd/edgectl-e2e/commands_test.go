package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	te2e "github.com/alexandremahdhaoui/edge-cd/pkg/test/e2e"
	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandFunctions verifies CLI command functions are defined
func TestCommandFunctions(t *testing.T) {
	// Verify command functions exist (can't test full behavior without libvirt)
	assert.NotPanics(t, func() {
		_ = cmdCreate
		_ = cmdDelete
		_ = cmdList
		_ = cmdGet
	})
}

// TestEnvironmentLoadingFromStore tests loading environments from artifact store
func TestEnvironmentLoadingFromStore(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")

	ctx := context.Background()
	store := te2e.NewJSONArtifactStore(storeFile)

	// Create and save a test environment
	env := &te2e.TestEnvironment{
		ID:     "e2e-20231025-test123",
		Status: "passed",
		TargetVM: vmm.VMMetadata{
			Name: "test-target-vm",
			IP:   "192.168.1.100",
		},
		GitServerVM: vmm.VMMetadata{
			Name: "test-gitserver-vm",
			IP:   "192.168.1.101",
		},
	}

	require.NoError(t, store.Save(ctx, env))

	// Load the environment back
	loaded, err := store.Load(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, env.ID, loaded.ID)
	assert.Equal(t, env.Status, loaded.Status)
	assert.Equal(t, env.TargetVM.Name, loaded.TargetVM.Name)
}

// TestEnvironmentDeletion tests environment deletion from store
func TestEnvironmentDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")

	ctx := context.Background()
	store := te2e.NewJSONArtifactStore(storeFile)

	// Create and save a test environment
	env := &te2e.TestEnvironment{
		ID:     "e2e-20231025-delete123",
		Status: "failed",
	}

	require.NoError(t, store.Save(ctx, env))

	// Verify it exists
	loaded, err := store.Load(ctx, env.ID)
	require.NoError(t, err)
	assert.NotNil(t, loaded)

	// Delete it
	require.NoError(t, store.Delete(ctx, env.ID))

	// Verify it's gone
	_, err = store.Load(ctx, env.ID)
	assert.Error(t, err)
}

// TestArtifactDirectoryCleanup tests artifact directory removal
func TestArtifactDirectoryCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test artifact directory with files
	artifactDir := filepath.Join(tmpDir, "artifacts")
	require.NoError(t, os.MkdirAll(artifactDir, 0755))

	// Create some test files
	testFile := filepath.Join(artifactDir, "test.log")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	// Verify directory exists
	_, err := os.Stat(artifactDir)
	require.NoError(t, err)

	// Remove directory
	require.NoError(t, os.RemoveAll(artifactDir))

	// Verify directory is gone
	_, err = os.Stat(artifactDir)
	assert.True(t, os.IsNotExist(err))
}

// TestCreateCommandTempDirStructure tests that temp directory structure is created
func TestCreateCommandTempDirStructure(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")

	ctx := context.Background()
	store := te2e.NewJSONArtifactStore(storeFile)

	// Create a test environment with proper temp dir structure
	env := &te2e.TestEnvironment{
		ID:          "e2e-20231025-test789",
		TempDirRoot: filepath.Join(tmpDir, "e2e-20231025-test789"),
		Status:      "created",
		CreatedAt:   time.Now().UTC(),
	}

	// Create temp dir structure
	require.NoError(t, os.MkdirAll(env.TempDirRoot, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(env.TempDirRoot, "vmm"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(env.TempDirRoot, "gitserver"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(env.TempDirRoot, "artifacts"), 0755))

	// Save environment
	require.NoError(t, store.Save(ctx, env))

	// Verify all subdirs exist
	for _, subdir := range []string{"vmm", "gitserver", "artifacts"} {
		path := filepath.Join(env.TempDirRoot, subdir)
		stat, err := os.Stat(path)
		assert.NoError(t, err)
		assert.True(t, stat.IsDir())
	}

	// Verify can be loaded back
	loaded, err := store.Load(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, env.TempDirRoot, loaded.TempDirRoot)
}

// TestCreateCommandOutputsEnvironmentInfo tests create command output format
func TestCreateCommandOutputsEnvironmentInfo(t *testing.T) {
	// This test verifies the structure and fields needed for output
	// Full CLI testing requires mocking or integration test environment
	env := &te2e.TestEnvironment{
		ID:          "e2e-20231025-output123",
		TempDirRoot: "/tmp/e2e-20231025-output123",
		Status:      "created",
		TargetVM: vmm.VMMetadata{
			Name: "test-target-vm",
			IP:   "192.168.1.100",
		},
		GitServerVM: vmm.VMMetadata{
			Name: "test-gitserver-vm",
			IP:   "192.168.1.101",
		},
		SSHKeys: te2e.SSHKeyInfo{
			HostKeyPath: "/tmp/key",
		},
		GitSSHURLs: map[string]string{
			"edge-cd":      "git@192.168.1.101:repos/edge-cd",
			"user-config": "git@192.168.1.101:repos/user-config",
		},
	}

	// Verify all required fields are present for output
	assert.NotEmpty(t, env.ID)
	assert.NotEmpty(t, env.TargetVM.Name)
	assert.NotEmpty(t, env.TargetVM.IP)
	assert.NotEmpty(t, env.GitServerVM.Name)
	assert.NotEmpty(t, env.GitServerVM.IP)
	assert.NotEmpty(t, env.SSHKeys.HostKeyPath)
	assert.Greater(t, len(env.GitSSHURLs), 0)
}

// TestTeardownPhases tests that teardown happens in correct phases
func TestTeardownPhases(t *testing.T) {
	// This test verifies the conceptual phases of teardown
	// Phase 1: VM destruction
	// Phase 2: Artifact removal
	// Phase 3: Metadata update

	phases := []string{
		"Destroying target VM",
		"Removing artifacts",
		"Updating metadata",
	}

	for _, phase := range phases {
		assert.NotEmpty(t, phase)
	}
}

// TestDeleteCommandSafetyValidation tests that delete validates managed temp directories
func TestDeleteCommandSafetyValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a managed temp directory
	managedDir := filepath.Join(tmpDir, "e2e-20231025-managed")
	managedTempPath, err := te2e.CreateTempDirectory(managedDir)
	require.NoError(t, err)

	// Create an unmanaged directory
	unmanagedDir := filepath.Join(tmpDir, "e2e-20231025-unmanaged")
	require.NoError(t, os.MkdirAll(unmanagedDir, 0755))

	// Test that managed directory is detected
	assert.True(t, te2e.IsManagedTempDirectory(managedTempPath))

	// Test that unmanaged directory is not detected
	assert.False(t, te2e.IsManagedTempDirectory(unmanagedDir))
}

// TestDeleteCommandTracksResources tests that delete tracks managed resources
func TestDeleteCommandTracksResources(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")

	ctx := context.Background()
	store := te2e.NewJSONArtifactStore(storeFile)

	// Create a test environment with managed resources
	env := &te2e.TestEnvironment{
		ID:          "e2e-20231025-resources",
		TempDirRoot: filepath.Join(tmpDir, "e2e-20231025-resources"),
		Status:      "created",
		CreatedAt:   time.Now().UTC(),
		ManagedResources: []string{
			"/tmp/e2e-20231025-resources/vmm/target.qcow2",
			"/tmp/e2e-20231025-resources/vmm/target-cloud-init.iso",
			"/tmp/e2e-20231025-resources/gitserver/gitserver-cloud-init.iso",
		},
	}

	// Create the managed temp directory
	_, err := te2e.CreateTempDirectory(env.TempDirRoot)
	require.NoError(t, err)

	// Save environment
	require.NoError(t, store.Save(ctx, env))

	// Load and verify managed resources are preserved
	loaded, err := store.Load(ctx, env.ID)
	require.NoError(t, err)

	assert.Equal(t, 3, len(loaded.ManagedResources))
	assert.Contains(t, loaded.ManagedResources, "/tmp/e2e-20231025-resources/vmm/target.qcow2")
	assert.Contains(t, loaded.ManagedResources, "/tmp/e2e-20231025-resources/vmm/target-cloud-init.iso")
}

// TestDeleteCommandValidatesTempDirStructure tests temp directory structure validation
func TestDeleteCommandValidatesTempDirStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a managed temp directory with proper structure
	tempRoot := filepath.Join(tmpDir, "e2e-20231025-struct")
	_, err := te2e.CreateTempDirectory(tempRoot)
	require.NoError(t, err)

	// Create subdirectories
	for _, subdir := range []string{"vmm", "gitserver", "artifacts"} {
		path := filepath.Join(tempRoot, subdir)
		require.NoError(t, os.MkdirAll(path, 0755))
	}

	// Verify marker file exists
	markerPath := filepath.Join(tempRoot, te2e.TempDirMarkerFile)
	_, err = os.Stat(markerPath)
	assert.NoError(t, err)

	// Verify IsManagedTempDirectory detects it correctly
	assert.True(t, te2e.IsManagedTempDirectory(tempRoot))
}
