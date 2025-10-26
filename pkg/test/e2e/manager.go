package e2e

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
)

// Status constants for test environments
const (
	StatusCreated           = "created"
	StatusSetup             = "setup"
	StatusRunning           = "running"
	StatusPassed            = "passed"
	StatusFailed            = "failed"
	StatusPartiallyDeleted  = "partially_deleted"
)

// TestEnvironment represents a complete test execution context
type TestEnvironment struct {
	ID               string            // Unique identifier (e.g., "e2e-20231025-abc123")
	CreatedAt        time.Time         // When the environment was created
	UpdatedAt        time.Time         // Last time environment was updated
	TargetVM         vmm.VMMetadata    // Target VM being tested
	GitServerVM      vmm.VMMetadata    // Git server VM for config repos
	ArtifactPath     string            // Root directory for all test artifacts
	TempDirRoot      string            // Root temp directory: /tmp/e2e-<test-id>. All component subdirs created here
	SSHKeys          SSHKeyInfo        // Paths to SSH keys used in this environment
	Status           string            // Current status: "setup", "running", "passed", "failed", "cleanup"
	Notes            string            // Optional notes for this environment
	GitSSHURLs       map[string]string // Git repository SSH URLs, keyed by repo name
	ManagedResources []string          // List of files/directories created during test (for audit and cleanup)
	TempDirs         []string          // Deprecated: kept for backward compatibility. Use TempDirRoot instead.
}

// SSHKeyInfo stores paths to SSH key files
type SSHKeyInfo struct {
	HostKeyPath      string // Private key for edgectl -> target VM connection
	HostKeyPubPath   string // Public key corresponding to HostKeyPath
	TargetKeyPath    string // Private key for target VM -> git server connection
	TargetKeyPubPath string // Public key corresponding to TargetKeyPath
}

// TestEnvironmentManager handles the lifecycle of test environments
type TestEnvironmentManager interface {
	// CreateEnvironment creates a new test environment with unique ID
	// Returns populated TestEnvironment (except VM/artifact details)
	CreateEnvironment(ctx context.Context) (*TestEnvironment, error)

	// GetEnvironment retrieves an existing environment by ID
	GetEnvironment(ctx context.Context, id string) (*TestEnvironment, error)

	// ListEnvironments returns all known environments (optionally filtered by status)
	ListEnvironments(ctx context.Context) ([]*TestEnvironment, error)

	// UpdateEnvironment saves changes to an existing environment
	UpdateEnvironment(ctx context.Context, env *TestEnvironment) error

	// DeleteEnvironment removes environment from internal tracking
	// Note: Does NOT delete VMs or artifacts (that's caller's responsibility)
	DeleteEnvironment(ctx context.Context, id string) error

	// GetArtifactDir returns the base directory where artifacts should be stored
	GetArtifactDir() string
}

// Manager implements TestEnvironmentManager with in-memory storage
type Manager struct {
	mu           sync.RWMutex
	environments map[string]*TestEnvironment
	artifactDir  string
}

// NewManager creates a new Manager instance with the given artifact directory
func NewManager(artifactDir string) *Manager {
	return &Manager{
		environments: make(map[string]*TestEnvironment),
		artifactDir:  artifactDir,
	}
}

// generateID creates a unique test environment ID in format: e2e-YYYYMMDD-XXXXXXXX
// where YYYYMMDD is the current date and XXXXXXXX is 8-character random alphanumeric
func (m *Manager) generateID() string {
	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	randomStr := randString(8)
	return fmt.Sprintf("e2e-%s-%s", dateStr, randomStr)
}

// randString generates a random alphanumeric string of the given length using crypto/rand
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback: should not happen in practice with crypto/rand
			panic(fmt.Sprintf("crypto/rand failed: %v", err))
		}
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

// CreateEnvironment creates a new test environment with a unique ID
// Returns a TestEnvironment with ID, timestamps, and default status set
func (m *Manager) CreateEnvironment(ctx context.Context) (*TestEnvironment, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	id := m.generateID()

	// Ensure uniqueness (though extremely unlikely)
	for m.environmentExists(id) {
		id = m.generateID()
	}

	env := &TestEnvironment{
		ID:               id,
		CreatedAt:        now,
		UpdatedAt:        now,
		Status:           "setup",
		GitSSHURLs:       make(map[string]string),
		ManagedResources: make([]string, 0),
	}

	m.environments[id] = env
	return env, nil
}

// environmentExists checks if an environment with the given ID exists (must be called under lock)
func (m *Manager) environmentExists(id string) bool {
	_, exists := m.environments[id]
	return exists
}

// GetEnvironment retrieves an existing environment by ID
func (m *Manager) GetEnvironment(ctx context.Context, id string) (*TestEnvironment, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	env, exists := m.environments[id]
	if !exists {
		return nil, fmt.Errorf("environment not found: %s", id)
	}

	// Return a copy to prevent external modifications
	return copyEnvironment(env), nil
}

// ListEnvironments returns all known environments
func (m *Manager) ListEnvironments(ctx context.Context) ([]*TestEnvironment, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	envs := make([]*TestEnvironment, 0, len(m.environments))
	for _, env := range m.environments {
		envs = append(envs, copyEnvironment(env))
	}
	return envs, nil
}

// UpdateEnvironment saves changes to an existing environment
func (m *Manager) UpdateEnvironment(ctx context.Context, env *TestEnvironment) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if env == nil || env.ID == "" {
		return fmt.Errorf("invalid environment: nil or empty ID")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.environments[env.ID]; !exists {
		return fmt.Errorf("environment not found: %s", env.ID)
	}

	env.UpdatedAt = time.Now().UTC()
	m.environments[env.ID] = copyEnvironment(env)
	return nil
}

// DeleteEnvironment removes environment from internal tracking
// Note: Does NOT delete VMs or artifacts (that's caller's responsibility)
func (m *Manager) DeleteEnvironment(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.environments[id]; !exists {
		return fmt.Errorf("environment not found: %s", id)
	}

	delete(m.environments, id)
	return nil
}

// GetArtifactDir returns the base directory where artifacts should be stored
func (m *Manager) GetArtifactDir() string {
	return m.artifactDir
}

// copyEnvironment creates a deep copy of a TestEnvironment to prevent external modifications
func copyEnvironment(env *TestEnvironment) *TestEnvironment {
	if env == nil {
		return nil
	}

	// Copy the main environment struct
	copy := *env

	// Deep copy the GitSSHURLs map
	if env.GitSSHURLs != nil {
		copy.GitSSHURLs = make(map[string]string)
		for k, v := range env.GitSSHURLs {
			copy.GitSSHURLs[k] = v
		}
	}

	return &copy
}
