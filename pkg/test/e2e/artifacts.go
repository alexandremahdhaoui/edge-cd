package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
)

// Error types for artifact store operations
var (
	// ErrNotFound is returned when a requested artifact or environment is not found
	ErrNotFound = errors.New("artifact not found")

	// ErrInvalidSchema is returned when artifact data has invalid structure
	ErrInvalidSchema = errors.New("invalid artifact schema")

	// ErrStorageFull is returned when storage operations fail due to capacity
	ErrStorageFull = errors.New("artifact storage full")
)

// ArtifactStore provides persistent storage for test environment metadata
type ArtifactStore interface {
	// Save persists a test environment to storage
	// If environment already exists, it overwrites it
	// Must create parent directories if needed
	Save(ctx execcontext.Context, env *TestEnvironment) error

	// Load retrieves a test environment by ID
	// Returns ErrNotFound if environment doesn't exist
	Load(ctx execcontext.Context, id string) (*TestEnvironment, error)

	// ListAll returns all persisted environments
	// Returns empty slice (not nil) if no environments exist
	ListAll(ctx execcontext.Context) ([]*TestEnvironment, error)

	// Delete removes an environment from storage
	// Returns ErrNotFound if environment doesn't exist
	Delete(ctx execcontext.Context, id string) error

	// GetStorePath returns the path where artifacts are stored
	GetStorePath() string

	// Close performs any necessary cleanup (e.g., close file handles)
	Close() error
}

// ArtifactStoreSchema represents the JSON structure for persistent storage
type ArtifactStoreSchema struct {
	Version      string                       `json:"version"`
	LastUpdated  time.Time                    `json:"last_updated"`
	Environments map[string]*TestEnvironment `json:"environments"`
}

// JSONArtifactStore implements ArtifactStore using JSON file persistence
type JSONArtifactStore struct {
	mu            sync.RWMutex
	filePath      string
	environments  map[string]*TestEnvironment
	lastUpdated   time.Time
	schemaVersion string
}

// NewJSONArtifactStore creates a new JSONArtifactStore instance
// The filePath should be the path to a JSON file (directory must exist or be created)
func NewJSONArtifactStore(filePath string) *JSONArtifactStore {
	return &JSONArtifactStore{
		filePath:      filePath,
		environments:  make(map[string]*TestEnvironment),
		schemaVersion: "1.0",
		lastUpdated:   time.Now().UTC(),
	}
}

// Save persists a test environment to storage
func (j *JSONArtifactStore) Save(ctx execcontext.Context, env *TestEnvironment) error {

	if env == nil || env.ID == "" {
		return fmt.Errorf("%w: nil or empty environment", ErrInvalidSchema)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	// Ensure parent directories exist
	if err := os.MkdirAll(filepath.Dir(j.filePath), 0755); err != nil {
		return fmt.Errorf("create artifact directory: %w", err)
	}

	// Load existing data from disk if not already loaded
	if len(j.environments) == 0 {
		if err := j.loadUnlocked(); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("load existing artifacts: %w", err)
		}
	}

	// Update or add environment
	j.environments[env.ID] = copyEnvironment(env)
	j.lastUpdated = time.Now().UTC()

	// Write to file
	return j.flush()
}

// Load retrieves a test environment by ID
func (j *JSONArtifactStore) Load(ctx execcontext.Context, id string) (*TestEnvironment, error) {

	if id == "" {
		return nil, fmt.Errorf("%w: empty ID", ErrInvalidSchema)
	}

	// Try to load from disk first if not already loaded
	if err := j.loadIfNeeded(); err != nil {
		// If file doesn't exist, that's OK - return not found
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return nil, err
	}

	j.mu.RLock()
	defer j.mu.RUnlock()

	env, exists := j.environments[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	// Return a copy to prevent external modifications
	return copyEnvironment(env), nil
}

// ListAll returns all persisted environments
func (j *JSONArtifactStore) ListAll(ctx execcontext.Context) ([]*TestEnvironment, error) {

	// Try to load from disk first if not already loaded
	if err := j.loadIfNeeded(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	j.mu.RLock()
	defer j.mu.RUnlock()

	envs := make([]*TestEnvironment, 0, len(j.environments))
	for _, env := range j.environments {
		envs = append(envs, copyEnvironment(env))
	}
	return envs, nil
}

// Delete removes an environment from storage
func (j *JSONArtifactStore) Delete(ctx execcontext.Context, id string) error {

	if id == "" {
		return fmt.Errorf("%w: empty ID", ErrInvalidSchema)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	// Try to load from disk first if not already loaded
	if _, err := os.Stat(j.filePath); err == nil {
		if err := j.loadUnlocked(); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	if _, exists := j.environments[id]; !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	delete(j.environments, id)
	j.lastUpdated = time.Now().UTC()

	// Write changes to disk
	return j.flush()
}

// GetStorePath returns the file path where artifacts are stored
func (j *JSONArtifactStore) GetStorePath() string {
	return j.filePath
}

// Close performs any necessary cleanup
func (j *JSONArtifactStore) Close() error {
	// For file-based storage, no cleanup needed
	return nil
}

// loadIfNeeded loads from disk if not already loaded (must not hold lock)
func (j *JSONArtifactStore) loadIfNeeded() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// If already loaded (len > 0 or file was loaded), don't reload
	if len(j.environments) > 0 {
		return nil
	}

	return j.loadUnlocked()
}

// loadUnlocked loads from disk (must be called with lock held)
func (j *JSONArtifactStore) loadUnlocked() error {
	// If file doesn't exist, that's OK - just start empty
	if _, err := os.Stat(j.filePath); os.IsNotExist(err) {
		j.environments = make(map[string]*TestEnvironment)
		return nil
	}

	// Read file
	data, err := os.ReadFile(j.filePath)
	if err != nil {
		return fmt.Errorf("read artifact file: %w", err)
	}

	// Parse JSON
	var schema ArtifactStoreSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("%w: invalid JSON: %v", ErrInvalidSchema, err)
	}

	// Validate schema
	if schema.Version == "" {
		return fmt.Errorf("%w: missing version field", ErrInvalidSchema)
	}

	if schema.Environments == nil {
		schema.Environments = make(map[string]*TestEnvironment)
	}

	j.environments = schema.Environments
	j.lastUpdated = schema.LastUpdated
	return nil
}

// flush writes the current state to disk (must be called with lock held)
func (j *JSONArtifactStore) flush() error {
	schema := ArtifactStoreSchema{
		Version:       j.schemaVersion,
		LastUpdated:   j.lastUpdated,
		Environments:  j.environments,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	// Write to file with proper permissions
	if err := os.WriteFile(j.filePath, data, 0o644); err != nil {
		return fmt.Errorf("write artifact file: %w", err)
	}

	return nil
}
