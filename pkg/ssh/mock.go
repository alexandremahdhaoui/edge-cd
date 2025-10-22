package ssh

import (
	"fmt"
	"sync"
)

// MockRunner is a mock implementation of the Runner interface for testing.
type MockRunner struct {
	mu        sync.Mutex
	Commands  []string // Stores commands that were run
	Responses map[string]struct {
		Stdout string
		Stderr string
		Err    error
	}
	DefaultStdout string
	DefaultStderr string
	DefaultErr    error
}

// NewMockRunner creates a new MockRunner.
func NewMockRunner() *MockRunner {
	return &MockRunner{
		Responses: make(map[string]struct {
			Stdout string
			Stderr string
			Err    error
		}),
	}
}

// Run records the command and returns a predefined response or a default.
func (m *MockRunner) Run(cmd string) (stdout, stderr string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Commands = append(m.Commands, cmd)

	if resp, ok := m.Responses[cmd]; ok {
		return resp.Stdout, resp.Stderr, resp.Err
	}

	return m.DefaultStdout, m.DefaultStderr, m.DefaultErr
}

// SetResponse sets a specific response for a given command.
func (m *MockRunner) SetResponse(cmd, stdout, stderr string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Responses[cmd] = struct {
		Stdout string
		Stderr string
		Err    error
	}{Stdout: stdout, Stderr: stderr, Err: err}
}

// AssertCommandRun asserts that a specific command was run.
func (m *MockRunner) AssertCommandRun(cmd string, otherCmds ...string) error {
	cmds := map[string]struct{}{cmd: {}}
	for _, s := range otherCmds {
		cmds[s] = struct{}{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.Commands {
		if _, ok := cmds[c]; ok {
			return nil
		}
	}
	return fmt.Errorf("command %q was not run", cmd)
}

// AssertNumberOfCommandsRun asserts the total number of commands run.
func (m *MockRunner) AssertNumberOfCommandsRun(expected int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Commands) != expected {
		return fmt.Errorf("expected %d commands to be run, but got %d", expected, len(m.Commands))
	}
	return nil
}
