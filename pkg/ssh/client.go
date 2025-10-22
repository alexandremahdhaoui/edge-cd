package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client implements the Runner interface for real SSH connections.
type Client struct {
	Host       string
	User       string
	PrivateKey []byte
	Password   string
	Port       string
}

// NewClient creates a new SSH client.
func NewClient(host, user, privateKeyPath, password, port string) (*Client, error) {
	key, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	return &Client{
			Host:       host,
			User:       user,
			PrivateKey: key,
			Password:   password,
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
			ssh.Password(c.Password),
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
