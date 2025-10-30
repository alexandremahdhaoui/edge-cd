package git

// MockRepoManager is a mock implementation of RepoManager for testing
type MockRepoManager struct {
	CloneRepoFunc        func(url, branch, destPath string, sparseCheckoutPaths []string) error
	SyncRepoFunc         func(repoPath, branch string, sparseCheckoutPaths []string) error
	GetCurrentCommitFunc func(repoPath string) (string, error)
	GetCommitDiffFunc    func(repoPath, oldCommit, newCommit string) ([]string, error)
}

// CloneRepo delegates to CloneRepoFunc if set
func (m *MockRepoManager) CloneRepo(url, branch, destPath string, sparseCheckoutPaths []string) error {
	if m.CloneRepoFunc != nil {
		return m.CloneRepoFunc(url, branch, destPath, sparseCheckoutPaths)
	}
	return nil
}

// SyncRepo delegates to SyncRepoFunc if set
func (m *MockRepoManager) SyncRepo(repoPath, branch string, sparseCheckoutPaths []string) error {
	if m.SyncRepoFunc != nil {
		return m.SyncRepoFunc(repoPath, branch, sparseCheckoutPaths)
	}
	return nil
}

// GetCurrentCommit delegates to GetCurrentCommitFunc if set
func (m *MockRepoManager) GetCurrentCommit(repoPath string) (string, error) {
	if m.GetCurrentCommitFunc != nil {
		return m.GetCurrentCommitFunc(repoPath)
	}
	return "mock-commit-hash", nil
}

// GetCommitDiff delegates to GetCommitDiffFunc if set
func (m *MockRepoManager) GetCommitDiff(repoPath, oldCommit, newCommit string) ([]string, error) {
	if m.GetCommitDiffFunc != nil {
		return m.GetCommitDiffFunc(repoPath, oldCommit, newCommit)
	}
	return []string{}, nil
}
