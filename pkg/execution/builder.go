package execution

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandBuilder provides a structured way to compose a base command with
// optional prepend prefix (e.g., "sudo") and environment variables.
// Both mechanisms can be used together - prepend is injected into the command string,
// environment variables are passed via exec.Command.Env.
type CommandBuilder struct {
	baseCmd  string
	prependCmd string
	envVars  map[string]string
}

// NewCommandBuilder creates a new CommandBuilder with the given base command.
// The base command is the actual command to execute (e.g., "git clone ...").
func NewCommandBuilder(baseCmd string) *CommandBuilder {
	return &CommandBuilder{
		baseCmd: baseCmd,
		envVars: make(map[string]string),
	}
}

// WithPrependCmd adds a prefix command to be prepended to the base command.
// This is typically used for privilege escalation (e.g., "sudo", "doas").
// Returns the builder for method chaining.
func (cb *CommandBuilder) WithPrependCmd(prefix string) *CommandBuilder {
	cb.prependCmd = prefix
	return cb
}

// WithEnvironment adds an environment variable that will be set when executing the command.
// Multiple calls to WithEnvironment can be made to set multiple variables.
// Returns the builder for method chaining.
func (cb *CommandBuilder) WithEnvironment(key, value string) *CommandBuilder {
	cb.envVars[key] = value
	return cb
}

// ComposeCommand returns the composed command string with optional prepend prefix,
// without building the full exec.Cmd. This is useful when you only need the command string
// (e.g., for use with SSH or shell execution) but don't need the full exec.Cmd structure.
//
// Order: [PREPENDCMD] [BASE_COMMAND]
// Example: "sudo git clone https://example.com/repo /dest"
func (cb *CommandBuilder) ComposeCommand() string {
	if cb.prependCmd == "" {
		return cb.baseCmd
	}

	// Prepend the prefix command with proper spacing
	return strings.TrimSpace(fmt.Sprintf("%s %s", cb.prependCmd, cb.baseCmd))
}

// Build constructs and returns a fully configured *exec.Cmd.
// The command is executed via /bin/sh -c with the base command,
// optionally prepended with the prefix command.
// Environment variables are set in the cmd.Env field.
//
// Command composition order: [PREPENDCMD] [BASE_COMMAND]
// Example: "sudo git clone https://example.com/repo /dest"
func (cb *CommandBuilder) Build() *exec.Cmd {
	// Compose the full command with optional prepend
	fullCmd := cb.composeCommand()

	// Create the shell command
	cmd := exec.Command("sh", "-c", fullCmd)

	// Set up environment: start with current environment
	cmd.Env = os.Environ()

	// Add any builder-provided environment variables
	for key, value := range cb.envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	return cmd
}

// composeCommand returns the full command string with optional prepend prefix.
// If prependCmd is empty, just returns the base command.
// If prependCmd is set, returns "[prependCmd] [baseCmd]" with proper spacing.
func (cb *CommandBuilder) composeCommand() string {
	if cb.prependCmd == "" {
		return cb.baseCmd
	}

	// Prepend the prefix command with proper spacing
	return strings.TrimSpace(fmt.Sprintf("%s %s", cb.prependCmd, cb.baseCmd))
}
