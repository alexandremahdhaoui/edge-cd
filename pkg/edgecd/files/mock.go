package files

import "github.com/alexandremahdhaoui/edge-cd/pkg/userconfig"

// MockFileReconciler is a mock implementation of FileReconciler for testing.
type MockFileReconciler struct {
	ReconcileFilesFunc func(configRepoPath, configPath string, files []userconfig.FileSpec) (*ReconcileResult, error)
}

// ReconcileFiles calls the mock function if set, otherwise returns empty result.
func (m *MockFileReconciler) ReconcileFiles(configRepoPath, configPath string, files []userconfig.FileSpec) (*ReconcileResult, error) {
	if m.ReconcileFilesFunc != nil {
		return m.ReconcileFilesFunc(configRepoPath, configPath, files)
	}
	return &ReconcileResult{
		ServicesToRestart: []string{},
		RequiresReboot:    false,
	}, nil
}
