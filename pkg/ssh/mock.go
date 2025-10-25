package ssh

import (
	"fmt"
	"sync"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execution"
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

// RunWithBuilder records the command built from a CommandBuilder and returns a predefined response or a default.
// It extracts the command string and environment variables from the builder and composes them into a full command string,
// then delegates to Run() to record and return the response.
func (m *MockRunner) RunWithBuilder(builder *execution.CommandBuilder) (stdout, stderr string, err error) {
	// Build the command to get the full command string and environment variables
	builtCmd := builder.Build()

	// Extract the command string from the built command
	// builtCmd.Args is ["sh", "-c", "actual_command_with_prepend"]
	if len(builtCmd.Args) < 3 {
		return "", "", fmt.Errorf("invalid command structure from builder")
	}
	commandStr := builtCmd.Args[2]

	// For the mock runner, we only need the command string itself
	// The builder's environment variables are already in builtCmd.Env,
	// but for test mocking purposes, we just use the command string
	// as the tests set responses based on command strings, not env vars

	// Delegate to Run() to record the command and return the response
	return m.Run(commandStr)
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
