package ssh

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
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

func (c *Client) Run(
	ctx execcontext.Context,
	cmd ...string,
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
	defer runFuncAndLogErr(conn.Close)

	session, err := conn.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("unable to create SSH session: %w", err)
	}
	defer runFuncAndLogErr(session.Close)

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Set environment variables for the SSH session
	for k, v := range ctx.Envs() {
		if err := session.Setenv(k, v); err != nil {
			// Log but don't fail - some servers don't support Setenv
			slog.Debug("unable to set environment variable", "key", k, "error", err)
		}
	}

	// Compose the final command with prepend commands
	finalCmd := strings.Join(cmd, " ")
	prependCmd := ctx.PrependCmd()
	if len(prependCmd) > 0 {
		// Prepend the command (e.g., "sudo", "-E")
		finalCmd = strings.Join(append(prependCmd, finalCmd), " ")
	}

	if err := session.Run(finalCmd); err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("remote command failed: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
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
				fmt.Printf(
					"failed to ssh to addr=%s\nwith err=%v\n",
					addr,
					err,
				)
				continue
			}

			_ = conn.Close()
			return nil // SSH server is available
		}
	}
}

func runFuncAndLogErr(f func() error) {
	if err := f(); err != nil {
		slog.Error(err.Error())
	}
}
