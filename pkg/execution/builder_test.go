package execution

import (
	"strings"
	"testing"
)

func TestNewCommandBuilder(t *testing.T) {
	builder := NewCommandBuilder("git clone https://example.com/repo /dest")
	if builder == nil {
		t.Fatal("expected builder, got nil")
	}
	if builder.baseCmd != "git clone https://example.com/repo /dest" {
		t.Fatalf("expected base command to be set, got: %s", builder.baseCmd)
	}
	if len(builder.envVars) != 0 {
		t.Fatalf("expected empty envVars, got: %v", builder.envVars)
	}
}

func TestCommandBuilder_BaseCommandOnly(t *testing.T) {
	builder := NewCommandBuilder("git clone https://example.com/repo /dest")
	cmd := builder.Build()

	if !strings.HasSuffix(cmd.Path, "sh") {
		t.Fatalf("expected shell path to end with 'sh', got: %s", cmd.Path)
	}

	// Check that the command is properly composed
	foundCmd := false
	for _, arg := range cmd.Args {
		if strings.Contains(arg, "git clone https://example.com/repo /dest") {
			foundCmd = true
			break
		}
	}
	if !foundCmd {
		t.Fatalf("expected base command in args, got: %v", cmd.Args)
	}

	// Verify no extra prepend was added
	if len(cmd.Env) == 0 {
		t.Fatal("expected environment variables to be inherited from os.Environ()")
	}
}

func TestCommandBuilder_WithPrependCmd(t *testing.T) {
	builder := NewCommandBuilder("git clone https://example.com/repo /dest")
	builder.WithPrependCmd("sudo")
	cmd := builder.Build()

	// Check that prepend is in the command
	foundPrepend := false
	for _, arg := range cmd.Args {
		if strings.Contains(arg, "sudo") && strings.Contains(arg, "git clone") {
			foundPrepend = true
			break
		}
	}
	if !foundPrepend {
		t.Fatalf("expected 'sudo' prepended to command, got: %v", cmd.Args)
	}
}

func TestCommandBuilder_WithEnvironment(t *testing.T) {
	builder := NewCommandBuilder("git clone https://example.com/repo /dest")
	builder.WithEnvironment("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no")
	cmd := builder.Build()

	// Check that environment variable is set
	foundEnv := false
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "GIT_SSH_COMMAND=") && strings.Contains(env, "ssh -o StrictHostKeyChecking=no") {
			foundEnv = true
			break
		}
	}
	if !foundEnv {
		t.Fatalf("expected GIT_SSH_COMMAND environment variable, got: %v", cmd.Env)
	}
}

func TestCommandBuilder_WithMultipleEnvironmentVariables(t *testing.T) {
	builder := NewCommandBuilder("git clone https://example.com/repo /dest")
	builder.WithEnvironment("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no")
	builder.WithEnvironment("SSH_AUTH_SOCK", "/tmp/ssh-agent.sock")
	cmd := builder.Build()

	envMap := make(map[string]bool)
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "GIT_SSH_COMMAND=") {
			envMap["GIT_SSH_COMMAND"] = true
		}
		if strings.HasPrefix(env, "SSH_AUTH_SOCK=") {
			envMap["SSH_AUTH_SOCK"] = true
		}
	}

	if !envMap["GIT_SSH_COMMAND"] || !envMap["SSH_AUTH_SOCK"] {
		t.Fatalf("expected both environment variables, got: %v", envMap)
	}
}

func TestCommandBuilder_WithPrependAndEnvironment(t *testing.T) {
	builder := NewCommandBuilder("git clone https://example.com/repo /dest")
	builder.WithPrependCmd("sudo")
	builder.WithEnvironment("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no")
	cmd := builder.Build()

	// Check command has prepend
	foundPrepend := false
	for _, arg := range cmd.Args {
		if strings.Contains(arg, "sudo") && strings.Contains(arg, "git clone") {
			foundPrepend = true
			break
		}
	}
	if !foundPrepend {
		t.Fatalf("expected 'sudo' prepended to command, got: %v", cmd.Args)
	}

	// Check environment variable is set
	foundEnv := false
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "GIT_SSH_COMMAND=") {
			foundEnv = true
			break
		}
	}
	if !foundEnv {
		t.Fatalf("expected GIT_SSH_COMMAND environment variable, got: %v", cmd.Env)
	}
}

func TestCommandBuilder_Chaining(t *testing.T) {
	// Test that method chaining works
	cmd := NewCommandBuilder("git clone https://example.com/repo /dest").
		WithPrependCmd("sudo").
		WithEnvironment("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no").
		Build()

	if cmd == nil {
		t.Fatal("expected command after chaining, got nil")
	}

	// Verify both prepend and env are present
	foundPrepend := false
	for _, arg := range cmd.Args {
		if strings.Contains(arg, "sudo") && strings.Contains(arg, "git clone") {
			foundPrepend = true
			break
		}
	}
	if !foundPrepend {
		t.Fatal("expected prepend in chained command")
	}

	foundEnv := false
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "GIT_SSH_COMMAND=") {
			foundEnv = true
			break
		}
	}
	if !foundEnv {
		t.Fatal("expected environment variable in chained command")
	}
}

func TestCommandBuilder_EmptyPrependCmd(t *testing.T) {
	builder := NewCommandBuilder("git clone https://example.com/repo /dest")
	builder.WithPrependCmd("")
	cmd := builder.Build()

	// Should just be the base command without extra spaces
	commandStr := cmd.Args[len(cmd.Args)-1] // Last arg is the command for "sh -c"
	if strings.Contains(commandStr, "  ") {
		t.Fatalf("expected no double spaces in command, got: %s", commandStr)
	}
	if !strings.Contains(commandStr, "git clone") {
		t.Fatalf("expected base command, got: %s", commandStr)
	}
}

func TestComposeCommand(t *testing.T) {
	tests := []struct {
		name       string
		baseCmd    string
		prependCmd string
		expected   string
	}{
		{
			name:       "base command only",
			baseCmd:    "git clone https://example.com/repo /dest",
			prependCmd: "",
			expected:   "git clone https://example.com/repo /dest",
		},
		{
			name:       "with prepend",
			baseCmd:    "git clone https://example.com/repo /dest",
			prependCmd: "sudo",
			expected:   "sudo git clone https://example.com/repo /dest",
		},
		{
			name:       "with doas",
			baseCmd:    "cp /src /dst",
			prependCmd: "doas",
			expected:   "doas cp /src /dst",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewCommandBuilder(tt.baseCmd)
			if tt.prependCmd != "" {
				builder.WithPrependCmd(tt.prependCmd)
			}
			composed := builder.composeCommand()
			if composed != tt.expected {
				t.Fatalf("expected: %s, got: %s", tt.expected, composed)
			}
		})
	}
}
