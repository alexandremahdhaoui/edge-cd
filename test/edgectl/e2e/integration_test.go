package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	te2e "github.com/alexandremahdhaoui/edge-cd/pkg/test/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentManagerIntegration tests the full lifecycle of environment management
func TestEnvironmentManagerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create manager
	manager := te2e.NewManager(tmpDir)

	// Create an environment
	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)
	require.NotNil(t, env)
	require.NotEmpty(t, env.ID)
	assert.Equal(t, "setup", env.Status)

	// Verify we can retrieve it
	retrieved, err := manager.GetEnvironment(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, env.ID, retrieved.ID)

	// Update the environment
	env.Status = "running"
	err = manager.UpdateEnvironment(ctx, env)
	require.NoError(t, err)

	// Verify update persisted
	retrieved, err = manager.GetEnvironment(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, "running", retrieved.Status)

	// List environments
	envs, err := manager.ListEnvironments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(envs))

	// Delete environment
	err = manager.DeleteEnvironment(ctx, env.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = manager.GetEnvironment(ctx, env.ID)
	assert.Error(t, err)
}

// TestArtifactStorePersistence tests that artifact store persists across instances
func TestArtifactStorePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	// Create first store instance and save environment
	store1 := te2e.NewJSONArtifactStore(storeFile)
	env := &te2e.TestEnvironment{
		ID:     "e2e-20231025-persist123",
		Status: "created",
	}
	require.NoError(t, store1.Save(ctx, env))
	require.NoError(t, store1.Close())

	// Create second store instance and load environment
	store2 := te2e.NewJSONArtifactStore(storeFile)
	loaded, err := store2.Load(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, env.ID, loaded.ID)
	assert.Equal(t, env.Status, loaded.Status)
}

// TestEnvironmentIDFormat tests that generated IDs follow the correct format
func TestEnvironmentIDFormat(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	manager := te2e.NewManager(tmpDir)

	// Create multiple environments and check ID format
	for i := 0; i < 5; i++ {
		env, err := manager.CreateEnvironment(ctx)
		require.NoError(t, err)

		// Verify format: e2e-YYYYMMDD-XXXXXXXX
		assert.Regexp(t, `^e2e-\d{8}-[a-zA-Z0-9]{8}$`, env.ID)

		// Clean up
		_ = manager.DeleteEnvironment(ctx, env.ID)
	}
}

// TestMultipleEnvironments tests managing multiple environments simultaneously
func TestMultipleEnvironments(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	manager := te2e.NewManager(tmpDir)
	store := te2e.NewJSONArtifactStore(storeFile)

	// Create multiple environments
	var envIDs []string
	for i := 0; i < 3; i++ {
		env, err := manager.CreateEnvironment(ctx)
		require.NoError(t, err)
		envIDs = append(envIDs, env.ID)

		// Save to store
		require.NoError(t, store.Save(ctx, env))
	}

	// Verify all can be retrieved
	for _, id := range envIDs {
		_, err := store.Load(ctx, id)
		require.NoError(t, err)
	}

	// Verify list shows all
	all, err := store.ListAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(all))
}

// TestEnvironmentUpdateFlow tests updating environment status through lifecycle
func TestEnvironmentUpdateFlow(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	manager := te2e.NewManager(tmpDir)
	store := te2e.NewJSONArtifactStore(storeFile)

	// Create environment
	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Simulate test lifecycle: setup -> running -> passed
	statuses := []string{"setup", "running", "passed"}
	for _, status := range statuses {
		env.Status = status
		require.NoError(t, manager.UpdateEnvironment(ctx, env))
		require.NoError(t, store.Save(ctx, env))

		// Verify status is persisted
		loaded, err := store.Load(ctx, env.ID)
		require.NoError(t, err)
		assert.Equal(t, status, loaded.Status)
	}
}

// TestErrorHandling tests error conditions
func TestErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	manager := te2e.NewManager(tmpDir)
	store := te2e.NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))

	// Test getting non-existent environment
	_, err := manager.GetEnvironment(ctx, "nonexistent")
	assert.Error(t, err)

	// Test loading non-existent from store
	_, err = store.Load(ctx, "nonexistent")
	assert.Error(t, err)

	// Test deleting non-existent
	err = manager.DeleteEnvironment(ctx, "nonexistent")
	assert.Error(t, err)

	// Test saving nil environment
	err = store.Save(ctx, nil)
	assert.Error(t, err)
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	manager := te2e.NewManager(tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Operations with cancelled context should fail
	_, err := manager.CreateEnvironment(ctx)
	assert.Error(t, err)
}

// TestArtifactPath tests that artifact paths can be set and used
func TestArtifactPath(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	manager := te2e.NewManager(tmpDir)
	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Set artifact path
	artifactDir := filepath.Join(tmpDir, "artifacts", env.ID)
	env.ArtifactPath = artifactDir

	// Create the directory
	require.NoError(t, os.MkdirAll(artifactDir, 0755))

	// Verify directory exists
	stat, err := os.Stat(artifactDir)
	require.NoError(t, err)
	assert.True(t, stat.IsDir())
}

// TestSSHKeyPathsStorage tests that SSH key paths can be stored
func TestSSHKeyPathsStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := te2e.NewJSONArtifactStore(storeFile)
	env := &te2e.TestEnvironment{
		ID: "e2e-20231025-keys123",
		SSHKeys: te2e.SSHKeyInfo{
			HostKeyPath:      "/tmp/id_rsa_host",
			HostKeyPubPath:   "/tmp/id_rsa_host.pub",
			TargetKeyPath:    "/tmp/id_rsa_target",
			TargetKeyPubPath: "/tmp/id_rsa_target.pub",
		},
	}

	require.NoError(t, store.Save(ctx, env))

	loaded, err := store.Load(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, env.SSHKeys.HostKeyPath, loaded.SSHKeys.HostKeyPath)
	assert.Equal(t, env.SSHKeys.TargetKeyPath, loaded.SSHKeys.TargetKeyPath)
}

// TestGitSSHURLs tests that Git SSH URLs can be stored and retrieved
func TestGitSSHURLs(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := te2e.NewJSONArtifactStore(storeFile)
	env := &te2e.TestEnvironment{
		ID: "e2e-20231025-git123",
		GitSSHURLs: map[string]string{
			"edge-cd":    "ssh://git@192.168.1.1/srv/git/edge-cd.git",
			"user-config": "ssh://git@192.168.1.1/srv/git/user-config.git",
		},
	}

	require.NoError(t, store.Save(ctx, env))

	loaded, err := store.Load(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(loaded.GitSSHURLs))
	assert.Equal(t, env.GitSSHURLs["edge-cd"], loaded.GitSSHURLs["edge-cd"])
}
