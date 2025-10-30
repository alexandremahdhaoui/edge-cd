package svcmgr

// MockServiceManager is a mock implementation of ServiceManager for testing
type MockServiceManager struct {
	EnableFunc  func(serviceName string) error
	RestartFunc func(serviceName string) error
	StartFunc   func(serviceName string) error

	// Track calls for verification
	EnableCalls  []string
	RestartCalls []string
	StartCalls   []string
}

// Enable calls the mock function if provided, otherwise returns nil
func (m *MockServiceManager) Enable(serviceName string) error {
	m.EnableCalls = append(m.EnableCalls, serviceName)
	if m.EnableFunc != nil {
		return m.EnableFunc(serviceName)
	}
	return nil
}

// Restart calls the mock function if provided, otherwise returns nil
func (m *MockServiceManager) Restart(serviceName string) error {
	m.RestartCalls = append(m.RestartCalls, serviceName)
	if m.RestartFunc != nil {
		return m.RestartFunc(serviceName)
	}
	return nil
}

// Start calls the mock function if provided, otherwise returns nil
func (m *MockServiceManager) Start(serviceName string) error {
	m.StartCalls = append(m.StartCalls, serviceName)
	if m.StartFunc != nil {
		return m.StartFunc(serviceName)
	}
	return nil
}
