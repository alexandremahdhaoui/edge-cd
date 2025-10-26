package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewJSONArtifactStore verifies store creation
func TestNewJSONArtifactStore(t *testing.T) {
	store := NewJSONArtifactStore("/tmp/artifacts.json")

	require.NotNil(t, store)
	assert.Equal(t, "/tmp/artifacts.json", store.GetStorePath())
	assert.Equal(t, "1.0", store.schemaVersion)
	assert.NotNil(t, store.environments)
	assert.Equal(t, 0, len(store.environments))
}

// TestSaveAndLoad verifies persistence to disk
func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")

	ctx := context.Background()

	// Create and save
	store1 := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:        "e2e-20231025-abc123",
		Status:    "running",
		Notes:     "Test environment",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := store1.Save(ctx, env)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Load and verify
	store2 := NewJSONArtifactStore(filePath)
	retrieved, err := store2.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)
	assert.Equal(t, "e2e-20231025-abc123", retrieved.ID)
	assert.Equal(t, "running", retrieved.Status)
	assert.Equal(t, "Test environment", retrieved.Notes)
}

// TestSaveNilEnvironment verifies error on nil environment
func TestSaveNilEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))
	ctx := context.Background()

	err := store.Save(ctx, nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}

// TestSaveEmptyIDEnvironment verifies error on empty ID
func TestSaveEmptyIDEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))
	ctx := context.Background()

	env := &TestEnvironment{ID: ""}
	err := store.Save(ctx, env)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}

// TestLoadByID verifies retrieval by ID
func TestLoadByID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:     "e2e-20231025-abc123",
		Status: "passed",
	}
	err := store.Save(ctx, env)
	require.NoError(t, err)

	retrieved, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)
	assert.Equal(t, "e2e-20231025-abc123", retrieved.ID)
	assert.Equal(t, "passed", retrieved.Status)
}

// TestLoadNotFound verifies error on non-existent ID
func TestLoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))
	ctx := context.Background()

	retrieved, err := store.Load(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.True(t, errors.Is(err, ErrNotFound))
}

// TestLoadEmptyID verifies error on empty ID
func TestLoadEmptyID(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))
	ctx := context.Background()

	retrieved, err := store.Load(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}

// TestLoadReturnsCopy verifies modifications don't affect store
func TestLoadReturnsCopy(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:         "e2e-20231025-abc123",
		Status:     "setup",
		GitSSHURLs: make(map[string]string),
	}
	err := store.Save(ctx, env)
	require.NoError(t, err)

	// Retrieve and modify
	retrieved, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)
	retrieved.Status = "modified"
	retrieved.GitSSHURLs["test"] = "test-url"

	// Verify internal state unchanged
	retrieved2, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)
	assert.Equal(t, "setup", retrieved2.Status)
	assert.NotContains(t, retrieved2.GitSSHURLs, "test")
}

// TestListAll verifies listing all environments
func TestListAll(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)

	// Create multiple environments
	for i := 0; i < 3; i++ {
		env := &TestEnvironment{
			ID:     "e2e-20231025-abc" + string(rune(i)),
			Status: "running",
		}
		err := store.Save(ctx, env)
		require.NoError(t, err)
	}

	// List all
	envs, err := store.ListAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(envs))
}

// TestListAllEmpty verifies empty list when no environments
func TestListAllEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))
	ctx := context.Background()

	envs, err := store.ListAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(envs))
	assert.NotNil(t, envs) // Should be empty slice, not nil
}

// TestListAllReturnsCopies verifies modifications don't affect store
func TestListAllReturnsCopies(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:         "e2e-20231025-abc123",
		Status:     "setup",
		GitSSHURLs: make(map[string]string),
	}
	err := store.Save(ctx, env)
	require.NoError(t, err)

	// List and modify
	envs, err := store.ListAll(ctx)
	require.NoError(t, err)
	envs[0].Status = "modified"
	envs[0].GitSSHURLs["test"] = "test-url"

	// Verify internal state unchanged
	retrieved, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)
	assert.Equal(t, "setup", retrieved.Status)
	assert.NotContains(t, retrieved.GitSSHURLs, "test")
}

// TestDelete verifies environment deletion
func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:     "e2e-20231025-abc123",
		Status: "running",
	}
	err := store.Save(ctx, env)
	require.NoError(t, err)

	// Verify exists
	_, err = store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)

	// Delete
	err = store.Delete(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)

	// Verify gone
	_, err = store.Load(ctx, "e2e-20231025-abc123")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}

// TestDeleteNotFound verifies error on non-existent ID
func TestDeleteNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))
	ctx := context.Background()

	err := store.Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}

// TestDeleteEmptyID verifies error on empty ID
func TestDeleteEmptyID(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))
	ctx := context.Background()

	err := store.Delete(ctx, "")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}

// TestDeletePersistsToDisk verifies deletion is persisted
func TestDeletePersistsToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	// Save with store1
	store1 := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{ID: "e2e-20231025-abc123"}
	err := store1.Save(ctx, env)
	require.NoError(t, err)

	// Delete with store1
	err = store1.Delete(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)

	// Load with store2 and verify deleted
	store2 := NewJSONArtifactStore(filePath)
	envs, err := store2.ListAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(envs))
}

// TestStoreUpdateEnvironment verifies updating existing environment in store
func TestStoreUpdateEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)

	// Save initial
	env := &TestEnvironment{
		ID:     "e2e-20231025-abc123",
		Status: "setup",
		Notes:  "Initial",
	}
	err := store.Save(ctx, env)
	require.NoError(t, err)

	// Update
	env.Status = "running"
	env.Notes = "Updated"
	err = store.Save(ctx, env)
	require.NoError(t, err)

	// Verify update
	retrieved, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)
	assert.Equal(t, "running", retrieved.Status)
	assert.Equal(t, "Updated", retrieved.Notes)
}

// TestJSONSchema verifies JSON structure written to disk
func TestJSONSchema(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:        "e2e-20231025-abc123",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
	}
	err := store.Save(ctx, env)
	require.NoError(t, err)

	// Read raw JSON
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// Parse JSON
	var schema ArtifactStoreSchema
	err = json.Unmarshal(data, &schema)
	require.NoError(t, err)

	// Verify schema fields
	assert.Equal(t, "1.0", schema.Version)
	assert.NotZero(t, schema.LastUpdated)
	assert.NotNil(t, schema.Environments)
	assert.Contains(t, schema.Environments, "e2e-20231025-abc123")
}

// TestJSONSchemaInvalid verifies error on invalid JSON
func TestJSONSchemaInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")

	// Write invalid JSON
	err := os.WriteFile(filePath, []byte("invalid json"), 0o644)
	require.NoError(t, err)

	ctx := context.Background()
	store := NewJSONArtifactStore(filePath)
	_, err = store.Load(ctx, "any")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}

// TestJSONSchemaMissingVersion verifies error on missing version
func TestJSONSchemaMissingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")

	// Write JSON without version
	schema := ArtifactStoreSchema{
		Version:      "",
		Environments: make(map[string]*TestEnvironment),
	}
	data, err := json.Marshal(schema)
	require.NoError(t, err)
	err = os.WriteFile(filePath, data, 0o644)
	require.NoError(t, err)

	ctx := context.Background()
	store := NewJSONArtifactStore(filePath)
	_, err = store.Load(ctx, "any")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSchema))
}

// TestMultipleSaveLoads verifies persistence across multiple operations
func TestMultipleSaveLoads(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	// Create, save, and load multiple times
	for i := 0; i < 3; i++ {
		store := NewJSONArtifactStore(filePath)
		env := &TestEnvironment{
			ID:     "e2e-20231025-abc" + string(rune(48+i)),
			Status: "running",
		}
		err := store.Save(ctx, env)
		require.NoError(t, err)
	}

	// Final load should have all 3
	store := NewJSONArtifactStore(filePath)
	envs, err := store.ListAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(envs))
}

// TestGetStorePathMethod verifies GetStorePath() returns correct value
func TestGetStorePathMethod(t *testing.T) {
	filePath := "/tmp/test-artifacts.json"
	store := NewJSONArtifactStore(filePath)

	assert.Equal(t, filePath, store.GetStorePath())
}

// TestCloseMethod verifies Close() works without error
func TestCloseMethod(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))

	err := store.Close()
	assert.NoError(t, err)
}

// TestEnvironmentWithGitSSHURLs verifies Git SSH URLs persistence
func TestEnvironmentWithGitSSHURLs(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:         "e2e-20231025-abc123",
		Status:     "running",
		GitSSHURLs: make(map[string]string),
	}
	env.GitSSHURLs["edge-cd"] = "ssh://git@192.168.1.1/srv/git/edge-cd.git"
	env.GitSSHURLs["user-config"] = "ssh://git@192.168.1.1/srv/git/user-config.git"

	err := store.Save(ctx, env)
	require.NoError(t, err)

	retrieved, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)

	assert.Equal(t, "ssh://git@192.168.1.1/srv/git/edge-cd.git", retrieved.GitSSHURLs["edge-cd"])
	assert.Equal(t, "ssh://git@192.168.1.1/srv/git/user-config.git", retrieved.GitSSHURLs["user-config"])
}

// TestEnvironmentWithVMMetadata verifies VM metadata persistence
func TestEnvironmentWithVMMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:     "e2e-20231025-abc123",
		Status: "running",
		TargetVM: vmm.VMMetadata{
			Name:      "e2e-target-abc123",
			IP:        "192.168.1.100",
			DomainXML: "<domain>...</domain>",
			SSHPort:   22,
			MemoryMB:  2048,
			VCPUs:     2,
		},
	}

	err := store.Save(ctx, env)
	require.NoError(t, err)

	retrieved, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)

	assert.Equal(t, "e2e-target-abc123", retrieved.TargetVM.Name)
	assert.Equal(t, "192.168.1.100", retrieved.TargetVM.IP)
	assert.Equal(t, "<domain>...</domain>", retrieved.TargetVM.DomainXML)
	assert.Equal(t, uint(2048), retrieved.TargetVM.MemoryMB)
	assert.Equal(t, uint(2), retrieved.TargetVM.VCPUs)
}

// TestContextCancellationSave verifies context cancellation in Save
func TestContextCancellationSave(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	env := &TestEnvironment{ID: "test"}
	err := store.Save(ctx, env)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestContextCancellationLoad verifies context cancellation in Load
func TestContextCancellationLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.Load(ctx, "any")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestContextCancellationListAll verifies context cancellation in ListAll
func TestContextCancellationListAll(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.ListAll(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestContextCancellationDelete verifies context cancellation in Delete
func TestContextCancellationDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONArtifactStore(filepath.Join(tmpDir, "artifacts.json"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := store.Delete(ctx, "any")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// BenchmarkSave measures save performance
func BenchmarkSave(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := &TestEnvironment{
			ID:     "e2e-20231025-abc" + string(rune(48 + (i % 10))),
			Status: "running",
		}
		_ = store.Save(ctx, env)
	}
}

// BenchmarkLoad measures load performance
func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{ID: "e2e-20231025-abc123"}
	_ = store.Save(ctx, env)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Load(ctx, "e2e-20231025-abc123")
	}
}

// BenchmarkListAll measures list performance
func BenchmarkListAll(b *testing.B) {
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	for i := 0; i < 100; i++ {
		env := &TestEnvironment{ID: "e2e-20231025-abc" + string(rune(48 + (i % 10)))}
		_ = store.Save(ctx, env)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.ListAll(ctx)
	}
}

// TestEnvironmentWithManagedResources verifies ManagedResources persistence
func TestEnvironmentWithManagedResources(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "artifacts.json")
	ctx := context.Background()

	store := NewJSONArtifactStore(filePath)
	env := &TestEnvironment{
		ID:               "e2e-20231025-abc123",
		Status:           "running",
		ManagedResources: []string{},
	}

	// Add some resources
	env.ManagedResources = append(env.ManagedResources, "/tmp/e2e-20231025-abc123/vmm/test-vm.qcow2")
	env.ManagedResources = append(env.ManagedResources, "/tmp/e2e-20231025-abc123/vmm/test-vm-cloud-init.iso")
	env.ManagedResources = append(env.ManagedResources, "/tmp/e2e-20231025-abc123/gitserver/id_rsa_gitserver")

	err := store.Save(ctx, env)
	require.NoError(t, err)

	retrieved, err := store.Load(ctx, "e2e-20231025-abc123")
	require.NoError(t, err)

	assert.Equal(t, 3, len(retrieved.ManagedResources))
	assert.Contains(t, retrieved.ManagedResources, "/tmp/e2e-20231025-abc123/vmm/test-vm.qcow2")
	assert.Contains(t, retrieved.ManagedResources, "/tmp/e2e-20231025-abc123/vmm/test-vm-cloud-init.iso")
	assert.Contains(t, retrieved.ManagedResources, "/tmp/e2e-20231025-abc123/gitserver/id_rsa_gitserver")
}
