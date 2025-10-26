package e2e

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateEnvironment verifies basic environment creation
func TestCreateEnvironment(t *testing.T) {
	manager := NewManager("/tmp/artifacts")

	ctx := execcontext.New(make(map[string]string), []string{})
	env, err := manager.CreateEnvironment(ctx)

	require.NoError(t, err)
	require.NotNil(t, env)
	assert.NotEmpty(t, env.ID)
	assert.Equal(t, "setup", env.Status)
	assert.NotZero(t, env.CreatedAt)
	assert.NotZero(t, env.UpdatedAt)
	assert.NotNil(t, env.GitSSHURLs)
	assert.Equal(t, 0, len(env.GitSSHURLs))
}

// TestIDFormat verifies ID generation produces correct format
func TestIDFormat(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Format should be: e2e-YYYYMMDD-XXXXXXXX
	pattern := `^e2e-\d{8}-[a-zA-Z0-9]{8}$`
	matched, err := regexp.MatchString(pattern, env.ID)
	require.NoError(t, err)
	assert.True(t, matched, fmt.Sprintf("ID %s doesn't match pattern %s", env.ID, pattern))
}

// TestIDContainsCurrentDate verifies date portion of ID is current date
func TestIDContainsCurrentDate(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Extract date portion from ID
	parts := strings.Split(env.ID, "-")
	require.Equal(t, 3, len(parts), "ID should have 3 parts separated by dashes")

	dateStr := parts[1]
	idDate, err := time.Parse("20060102", dateStr)
	require.NoError(t, err)

	now := time.Now().UTC()
	expectedDate := now.Format("20060102")

	assert.Equal(t, expectedDate, dateStr)
	// Verify the parsed date is today (within 1 day to account for day boundaries)
	assert.True(t, idDate.After(now.Add(-24*time.Hour)))
	assert.True(t, idDate.Before(now.Add(24*time.Hour)))
}

// TestIDUniqueness verifies that rapidly generated IDs are unique
func TestIDUniqueness(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		env, err := manager.CreateEnvironment(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, env.ID)

		// Check for duplicates
		assert.False(t, ids[env.ID], "Duplicate ID generated: %s", env.ID)
		ids[env.ID] = true
	}

	assert.Equal(t, 100, len(ids), "Should have 100 unique IDs")
}

// TestGetEnvironment verifies retrieval of existing environment
func TestGetEnvironment(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	// Create an environment
	created, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Retrieve it
	retrieved, err := manager.GetEnvironment(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Verify contents match
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Status, retrieved.Status)
	assert.Equal(t, created.CreatedAt, retrieved.CreatedAt)
}

// TestGetEnvironmentNotFound verifies error when environment doesn't exist
func TestGetEnvironmentNotFound(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	retrieved, err := manager.GetEnvironment(ctx, "e2e-nonexistent")
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Contains(t, err.Error(), "not found")
}

// TestGetEnvironmentReturnsCopy verifies modifications don't affect internal state
func TestGetEnvironmentReturnsCopy(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	created, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Retrieve environment
	retrieved, err := manager.GetEnvironment(ctx, created.ID)
	require.NoError(t, err)

	// Modify the retrieved copy
	retrieved.Status = "modified"
	retrieved.GitSSHURLs["test"] = "test-url"

	// Retrieve again and verify internal state is unchanged
	retrieved2, err := manager.GetEnvironment(ctx, created.ID)
	require.NoError(t, err)

	assert.Equal(t, "setup", retrieved2.Status)
	assert.NotContains(t, retrieved2.GitSSHURLs, "test")
}

// TestListEnvironments verifies listing all environments
func TestListEnvironments(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	// Create multiple environments
	ids := []string{}
	for i := 0; i < 5; i++ {
		env, err := manager.CreateEnvironment(ctx)
		require.NoError(t, err)
		ids = append(ids, env.ID)
	}

	// List all environments
	envs, err := manager.ListEnvironments(ctx)
	require.NoError(t, err)
	require.Equal(t, 5, len(envs))

	// Verify all created IDs are in the list
	idMap := make(map[string]bool)
	for _, env := range envs {
		idMap[env.ID] = true
	}

	for _, id := range ids {
		assert.True(t, idMap[id], "Created ID not found in list: %s", id)
	}
}

// TestListEnvironmentsEmpty verifies empty list when no environments
func TestListEnvironmentsEmpty(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	envs, err := manager.ListEnvironments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(envs))
}

// TestListEnvironmentsReturnsCopies verifies modifications don't affect internal state
func TestListEnvironmentsReturnsCopies(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	created, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// List environments
	envs, err := manager.ListEnvironments(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(envs))

	// Modify the returned copy
	envs[0].Status = "modified"
	envs[0].GitSSHURLs["test"] = "test-url"

	// Retrieve again and verify internal state is unchanged
	retrieved, err := manager.GetEnvironment(ctx, created.ID)
	require.NoError(t, err)

	assert.Equal(t, "setup", retrieved.Status)
	assert.NotContains(t, retrieved.GitSSHURLs, "test")
}

// TestUpdateEnvironment verifies updating an existing environment
func TestUpdateEnvironment(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	// Create environment
	created, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)
	originalCreatedAt := created.CreatedAt

	// Wait a bit to ensure UpdatedAt differs from CreatedAt
	time.Sleep(10 * time.Millisecond)

	// Update the environment
	created.Status = "running"
	created.Notes = "Test note"
	created.GitSSHURLs["edge-cd"] = "ssh://git@192.168.1.1/srv/git/edge-cd.git"

	err = manager.UpdateEnvironment(ctx, created)
	require.NoError(t, err)

	// Retrieve and verify changes
	retrieved, err := manager.GetEnvironment(ctx, created.ID)
	require.NoError(t, err)

	assert.Equal(t, "running", retrieved.Status)
	assert.Equal(t, "Test note", retrieved.Notes)
	assert.Equal(t, "ssh://git@192.168.1.1/srv/git/edge-cd.git", retrieved.GitSSHURLs["edge-cd"])
	assert.Equal(t, originalCreatedAt, retrieved.CreatedAt)
	assert.True(t, retrieved.UpdatedAt.After(originalCreatedAt))
}

// TestUpdateEnvironmentNotFound verifies error when updating non-existent environment
func TestUpdateEnvironmentNotFound(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	env := &TestEnvironment{ID: "nonexistent"}
	err := manager.UpdateEnvironment(ctx, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestUpdateEnvironmentNil verifies error when updating nil environment
func TestUpdateEnvironmentNil(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	err := manager.UpdateEnvironment(ctx, nil)
	assert.Error(t, err)
}

// TestDeleteEnvironment verifies deleting an environment
func TestDeleteEnvironment(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	// Create environment
	created, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Verify it exists
	_, err = manager.GetEnvironment(ctx, created.ID)
	require.NoError(t, err)

	// Delete it
	err = manager.DeleteEnvironment(ctx, created.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = manager.GetEnvironment(ctx, created.ID)
	assert.Error(t, err)
}

// TestDeleteEnvironmentNotFound verifies error when deleting non-existent environment
func TestDeleteEnvironmentNotFound(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	err := manager.DeleteEnvironment(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestDeleteEnvironmentRemovesFromList verifies environment is removed from list
func TestDeleteEnvironmentRemovesFromList(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	// Create 3 environments
	ids := []string{}
	for i := 0; i < 3; i++ {
		env, err := manager.CreateEnvironment(ctx)
		require.NoError(t, err)
		ids = append(ids, env.ID)
	}

	// Delete the middle one
	err := manager.DeleteEnvironment(ctx, ids[1])
	require.NoError(t, err)

	// List should have 2 remaining
	envs, err := manager.ListEnvironments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(envs))

	idMap := make(map[string]bool)
	for _, env := range envs {
		idMap[env.ID] = true
	}

	assert.True(t, idMap[ids[0]])
	assert.False(t, idMap[ids[1]])
	assert.True(t, idMap[ids[2]])
}

// TestGetArtifactDir verifies artifact directory is returned
func TestGetArtifactDir(t *testing.T) {
	artifactDir := "/tmp/test-artifacts"
	manager := NewManager(artifactDir)

	result := manager.GetArtifactDir()
	assert.Equal(t, artifactDir, result)
}

// TestConcurrentOperations verifies thread-safety with concurrent operations
func TestConcurrentOperations(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	var wg sync.WaitGroup
	const goroutines = 10
	const opsPerGoroutine = 10

	// Concurrent creates
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				_, err := manager.CreateEnvironment(ctx)
				require.NoError(t, err)
			}
		}()
	}

	wg.Wait()

	// Verify all environments were created
	envs, err := manager.ListEnvironments(ctx)
	require.NoError(t, err)
	assert.Equal(t, goroutines*opsPerGoroutine, len(envs))
}

// TestConcurrentUpdates verifies thread-safety with concurrent updates
func TestConcurrentUpdates(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	// Create one environment
	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	const goroutines = 5

	// Concurrent updates with different statuses
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			envCopy, err := manager.GetEnvironment(ctx, env.ID)
			require.NoError(t, err)

			envCopy.Status = fmt.Sprintf("status-%d", index)
			err = manager.UpdateEnvironment(ctx, envCopy)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Final status should be one of the set statuses (last one wins)
	final, err := manager.GetEnvironment(ctx, env.ID)
	require.NoError(t, err)
	assert.Contains(t, final.Status, "status-")
}

// TestTimestampAccuracy verifies timestamps are set correctly
func TestTimestampAccuracy(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	beforeCreate := time.Now().UTC()
	env, err := manager.CreateEnvironment(ctx)
	afterCreate := time.Now().UTC()

	require.NoError(t, err)
	assert.True(t, env.CreatedAt.After(beforeCreate.Add(-1*time.Second)))
	assert.True(t, env.CreatedAt.Before(afterCreate.Add(1*time.Second)))
	assert.Equal(t, env.CreatedAt, env.UpdatedAt)
}

// TestEnvironmentStructFields verifies all fields in created environment
func TestEnvironmentStructFields(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Check all required fields are initialized
	assert.NotEmpty(t, env.ID)
	assert.NotZero(t, env.CreatedAt)
	assert.NotZero(t, env.UpdatedAt)
	assert.Equal(t, "setup", env.Status)
	assert.NotNil(t, env.GitSSHURLs)
	assert.Empty(t, env.ArtifactPath)    // Should be empty initially
	assert.Empty(t, env.Notes)            // Should be empty initially
}

// BenchmarkCreateEnvironment measures performance of environment creation
func BenchmarkCreateEnvironment(b *testing.B) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.CreateEnvironment(ctx)
		if err != nil {
			b.Fatalf("CreateEnvironment failed: %v", err)
		}
	}
}

// BenchmarkIDGeneration measures ID generation performance
func BenchmarkIDGeneration(b *testing.B) {
	manager := NewManager("/tmp/artifacts")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.generateID()
	}
}

// TestManagedResourcesTracking verifies ManagedResources are properly initialized
func TestManagedResourcesTracking(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// ManagedResources should be initialized as empty slice
	assert.NotNil(t, env.ManagedResources)
	assert.Equal(t, 0, len(env.ManagedResources))
}

// TestManagedResourcesPersistence verifies ManagedResources can be updated and persisted
func TestManagedResourcesPersistence(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Add some managed resources
	env.ManagedResources = append(env.ManagedResources, "/tmp/e2e-xxx/vmm/disk.qcow2")
	env.ManagedResources = append(env.ManagedResources, "/tmp/e2e-xxx/vmm/cloud-init.iso")

	// Update environment
	err = manager.UpdateEnvironment(ctx, env)
	require.NoError(t, err)

	// Retrieve and verify resources were persisted
	retrieved, err := manager.GetEnvironment(ctx, env.ID)
	require.NoError(t, err)

	assert.Equal(t, 2, len(retrieved.ManagedResources))
	assert.Contains(t, retrieved.ManagedResources, "/tmp/e2e-xxx/vmm/disk.qcow2")
	assert.Contains(t, retrieved.ManagedResources, "/tmp/e2e-xxx/vmm/cloud-init.iso")
}

// TestParseIDExtractsDate verifies that date can be extracted from ID
func TestParseIDExtractsDate(t *testing.T) {
	manager := NewManager("/tmp/artifacts")
	ctx := execcontext.New(make(map[string]string), []string{})

	env, err := manager.CreateEnvironment(ctx)
	require.NoError(t, err)

	// Extract and parse date from ID
	parts := strings.Split(env.ID, "-")
	require.Equal(t, 3, len(parts))

	dateStr := parts[1]
	parsedDate, err := strconv.Atoi(dateStr)
	require.NoError(t, err)

	// Verify it's a valid date in YYYYMMDD format
	year := parsedDate / 10000
	month := (parsedDate % 10000) / 100
	day := parsedDate % 100

	assert.True(t, year >= 2020, "Year should be >= 2020")
	assert.True(t, month >= 1 && month <= 12, "Month should be 1-12")
	assert.True(t, day >= 1 && day <= 31, "Day should be 1-31")
}
