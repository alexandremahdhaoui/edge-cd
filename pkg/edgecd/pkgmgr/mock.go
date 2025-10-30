package pkgmgr

// MockPackageManager is a mock implementation of PackageManager for testing
type MockPackageManager struct {
	UpdateFunc  func() error
	InstallFunc func(packages []string) error
	UpgradeFunc func(packages []string) error
}

// Update calls the mock UpdateFunc if set, otherwise returns nil
func (m *MockPackageManager) Update() error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc()
	}
	return nil
}

// Install calls the mock InstallFunc if set, otherwise returns nil
func (m *MockPackageManager) Install(packages []string) error {
	if m.InstallFunc != nil {
		return m.InstallFunc(packages)
	}
	return nil
}

// Upgrade calls the mock UpgradeFunc if set, otherwise returns nil
func (m *MockPackageManager) Upgrade(packages []string) error {
	if m.UpgradeFunc != nil {
		return m.UpgradeFunc(packages)
	}
	return nil
}
