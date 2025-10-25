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

// Run executes a command on the remote host via SSH.
func (c *Client) Run(cmd string) (stdout, stderr string, err error) {
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

// RunWithBuilder executes a command built with CommandBuilder on the remote host via SSH.
// It properly injects both prepend commands and environment variables into the remote execution.
//
// The builder's command is constructed with a prepend prefix and environment variables.
// This method extracts both and composes them into a proper SSH command:
// "ENV1=val1 ENV2=val2 ... sh -c 'prepend_command base_command'"
func (c *Client) RunWithBuilder(builder *execution.CommandBuilder) (stdout, stderr string, err error) {
	// Build the command to get the full command string and environment variables
	builtCmd := builder.Build()

	// Extract the command string from the built command
	// builtCmd.Args is ["sh", "-c", "actual_command_with_prepend"]
	if len(builtCmd.Args) < 3 {
		return "", "", fmt.Errorf("invalid command structure from builder")
	}
	commandStr := builtCmd.Args[2]

	// Extract environment variables added by the builder (not in os.Environ())
	addedEnvVars := extractAddedEnvVars(builtCmd.Env)

	// Construct the full command with environment variables prefixed
	fullCmd := composeSSHCommand(addedEnvVars, commandStr)

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

// composeSSHCommand combines environment variables and a command into a single
// SSH-executable command string. Environment variables are prefixed to the command.
//
// Example:
//   addedEnvVars: ["GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no"]
//   commandStr: "sudo git clone https://example.com/repo /dest"
//   Result: "GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no' sudo git clone https://example.com/repo /dest"
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
