package ssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

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
