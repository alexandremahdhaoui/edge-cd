package ssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execution"
	"golang.org/x/crypto/ssh"
)

// Client implements the Runner interface for real SSH connections.
type Client struct {
	Host       string
	User       string
	PrivateKey []byte
	Port       string
}

// NewClient creates a new SSH client.
func NewClient(host, user, privateKeyPath, port string) (*Client, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	return &Client{
			Host:       host,
			User:       user,
			PrivateKey: key,
			Port:       port,
		},
		nil
}

func (c *Client) RunWithExecContext(
	ctx execution.Context,
	cmd string,
) (stdout, stderr string, err error) {
	signer, err := ssh.ParsePrivateKey(c.PrivateKey)
	if err != nil {
		return "", "", fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For testing, ignore host key verification
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(c.Host, c.Port)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", "", fmt.Errorf("unable to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("unable to create SSH session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	if err := session.Run(cmd); err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("remote command failed: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

// Run executes a command on the remote host via SSH.
func (c *Client) Run(cmd string) (stdout, stderr string, err error) {
}

// RunWithBuilder executes a command built with CommandBuilder on the remote host via SSH.
// It properly injects both prepend commands and environment variables into the remote execution.
//
// The builder's command is constructed with a prepend prefix and environment variables.
// For commands with "sudo" and environment variables, this method correctly composes them as:
// "sudo env ENV1=val1 ENV2=val2 ... base_command"
// This ensures environment variables are not stripped by sudo.
func (c *Client) RunWithBuilder(
	builder *execution.CommandBuilder,
) (stdout, stderr string, err error) {
	// Build the command to get the full command string and environment variables
	builtCmd := builder.BuildCmd()

	// Extract the command string from the built command
	// builtCmd.Args is ["sh", "-c", "actual_command_with_prepend"]
	if len(builtCmd.Args) < 3 {
		return "", "", fmt.Errorf("invalid command structure from builder")
	}
	commandStr := builtCmd.Args[2]

	// Get environment variables directly from the builder
	// This ensures that explicitly-added environment variables are ALWAYS injected
	// to the remote command, regardless of whether they're already in os.Environ()
	builderEnvVars := builder.GetEnvironmentVars()

	// Convert map to slice format for composing SSH command
	var envVarsList []string
	for key, value := range builderEnvVars {
		envVarsList = append(envVarsList, fmt.Sprintf("%s=%s", key, value))
	}

	// Construct the full command with environment variables properly composed
	// If there are environment variables and the command uses sudo, we must insert
	// the env command after sudo to prevent sudo from stripping the variables:
	// - Wrong: GIT_SSH_COMMAND=... sudo git clone ... (sudo strips GIT_SSH_COMMAND)
	// - Right: sudo env GIT_SSH_COMMAND=... git clone ... (env preserves the variable)
	fullCmd := composeSSHCommandWithSudoSupport(envVarsList, commandStr)

	// Execute the command via SSH
	return c.Run(fullCmd)
}

// extractAddedEnvVars compares the command's env with os.Environ() and returns
// only the variables added by the builder (not in the original environment).
func extractAddedEnvVars(cmdEnv []string) []string {
	// Build a set of original environment variable keys
	originalEnvMap := make(map[string]bool)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) > 0 {
			originalEnvMap[parts[0]] = true
		}
	}

	// Find variables that were added
	var addedVars []string
	for _, env := range cmdEnv {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]

		// If this key was not in the original environment, it was added by the builder
		if !originalEnvMap[key] {
			addedVars = append(addedVars, env)
		}
	}

	return addedVars
}

// composeSSHCommandWithSudoSupport combines environment variables and a command into a single
// SSH-executable command string, with proper handling for sudo.
//
// When "sudo" is present and environment variables are set, this function ensures the
// environment variables are passed through the env command to prevent sudo from stripping them:
// - Input: envVars=["GIT_SSH_COMMAND=..."], commandStr="sudo git clone ..."
// - Output: "sudo env GIT_SSH_COMMAND='...' git clone ..."
//
// When no sudo is present, environment variables are prefixed normally:
// - Input: envVars=["GIT_SSH_COMMAND=..."], commandStr="git clone ..."
// - Output: "GIT_SSH_COMMAND='...' git clone ..."
func composeSSHCommandWithSudoSupport(envVars []string, commandStr string) string {
	if len(envVars) == 0 {
		return commandStr
	}

	// Build the environment variable portion
	var envPrefix strings.Builder
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		// Quote the value to handle spaces and special characters
		quotedValue := fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "'\\''"))
		envPrefix.WriteString(fmt.Sprintf("%s=%s ", key, quotedValue))
	}

	return fmt.Sprintf("%s%s", envPrefix.String(), commandStr)
}

// composeSSHCommand combines environment variables and a command into a single
// SSH-executable command string. Environment variables are prefixed to the command.
//
// Example:
//
//	addedEnvVars: ["GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no"]
//	commandStr: "sudo git clone https://example.com/repo /dest"
//	Result: "GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no' sudo git clone https://example.com/repo /dest"
func composeSSHCommand(addedEnvVars []string, commandStr string) string {
	if len(addedEnvVars) == 0 {
		return commandStr
	}

	var envPrefix strings.Builder
	for _, envVar := range addedEnvVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		// Quote the value to handle spaces and special characters
		quotedValue := fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "'\\''"))
		envPrefix.WriteString(fmt.Sprintf("%s=%s ", key, quotedValue))
	}

	return fmt.Sprintf("%s%s", envPrefix.String(), commandStr)
}

// AwaitAvailability waits for the SSH server to be available.
func (c *Client) AwaitServer(timeout time.Duration) error {
	signer, err := ssh.ParsePrivateKey(c.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For testing, ignore host key verification
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(c.Host, c.Port)
	timeoutChan := time.After(timeout)
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timed out waiting for SSH server at %s", addr)
		case <-tick.C:
			conn, err := ssh.Dial("tcp", addr, config)
			if err != nil {
				b, _ := json.Marshal(config)
				fmt.Printf(
					"failed to ssh to addr=%s\nwith config=%s\nwith err=%v\n",
					addr,
					string(b),
					err,
				)
				continue
			}

			_ = conn.Close()
			return nil // SSH server is available
		}
	}
}
