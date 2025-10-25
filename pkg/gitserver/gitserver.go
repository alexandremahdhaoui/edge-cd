package gitserver

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/cloudinit"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
)

//go:embed Dockerfile
var Dockerfile string

//go:embed entrypoint.sh
var Entrypoint string

type SourceType int

const (
	LocalSource SourceType = iota
	GitUrlSource
)

type Source struct {
	Type      SourceType
	LocalPath string
	GitUrl    string
}

type Repo struct {
	Name   string
	Source Source
}

type Server struct {
	name               string
	ServerAddr         string
	SSHPort            int
	AuthorizedKeys     []string
	BaseDir            string
	Repo               []Repo
	initPrivateKeyPath string

	// -- VM related fields
	vmm            *vmm.VMM
	vmConfig       vmm.VMConfig
	vmIPAddress    string
	tempDir        string // Temporary directory for SSH keys and other temporary files
	imageQCOW2Path string // Path to the base QCOW2 image for the VM

	// -- Docker related fields (to be removed later)
	authorizedKeysFile string
	buildDir           string
	gitDir             string
	runningContainer   string
}

func NewServer(baseDir, imageQCOW2Path string, repo []Repo) *Server {
	return &Server{
		name:           fmt.Sprintf("gitserver-%d", time.Now().UnixNano()),
		ServerAddr:     "localhost",
		SSHPort:        22,
		AuthorizedKeys: []string{},
		BaseDir:        baseDir,
		Repo:           repo,
		tempDir:        baseDir,
		imageQCOW2Path: imageQCOW2Path,
	}
}

func (s *Server) Run() error {
	if err := s.initVM(); err != nil {
		return fmt.Errorf("failed to initialize VM: %v", err)
	}

	var err error
	s.vmm, err = vmm.NewVMM()
	if err != nil {
		return fmt.Errorf("failed to create VMM: %v", err)
	}

	if _, err := s.vmm.CreateVM(s.vmConfig); err != nil {
		return fmt.Errorf("failed to create VM: %v", err)
	}

	s.vmIPAddress, err = s.vmm.GetVMIPAddress(s.vmConfig.Name)
	if err != nil {
		return fmt.Errorf("failed to get VM IP address: %v", err)
	}
	s.ServerAddr = s.vmIPAddress

	if len(s.Repo) > 0 {
		sshClient, err := s.sshClient()
		if err != nil {
			return fmt.Errorf("failed to create ssh client for initAndPushRepo: %w", err)
		}

		for _, repo := range s.Repo {
			if repo.Source.Type != LocalSource {
				return fmt.Errorf("unsupported repo source type for repo %s", repo.Name)
			}
			if err := s.initAndPushRepo(sshClient, repo.Name, repo.Source.LocalPath); err != nil {
				return fmt.Errorf("failed to init and push repo %s: %w", repo.Name, err)
			}
		}
	}

	return nil
}

func (s *Server) init() error {
	baseDir := s.BaseDir
	if baseDir == "" {
		var err error
		baseDir, err = os.MkdirTemp("", "git-server")
		if err != nil {
			return err
		}
	}
	// Initialize tempDir if it's not already set (e.g., by a test)
	if s.tempDir == "" {
		s.tempDir = baseDir
	}

	sshDir := filepath.Join(baseDir, "ssh")
	s.authorizedKeysFile = filepath.Join(sshDir, "authorized_keys")
	s.buildDir = filepath.Join(s.BaseDir, "build")
	s.gitDir = filepath.Join(s.BaseDir, "git")
	for _, dir := range []string{sshDir, s.buildDir, s.gitDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// initVM prepares the vmm.VMConfig and cloudinit.UserData for the Git server VM.
func (s *Server) initVM() error {
	// 1. Generate a new SSH key pair for the Git server VM
	s.initPrivateKeyPath = filepath.Join(s.tempDir, "id_rsa_gitserver")
	sshPublicKeyPath := s.initPrivateKeyPath + ".pub"

	cmd := exec.Command(
		"ssh-keygen",
		"-t",
		"rsa",
		"-b",
		"2048",
		"-f",
		s.initPrivateKeyPath,
		"-N",
		"",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(
			"failed to generate SSH key pair for Git server VM: %w\nOutput: %s",
			err,
			output,
		)
	}
	if err := os.Chmod(s.initPrivateKeyPath, 0o600); err != nil {
		return fmt.Errorf("failed to set permissions on Git server VM SSH private key: %v", err)
	}

	// 2. Configure cloud-init UserData
	publicKeyBytes, err := os.ReadFile(sshPublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read Git server VM SSH public key: %w", err)
	}

	authorizedKeys := append(s.AuthorizedKeys, strings.TrimSpace(string(publicKeyBytes)))

	gitUser := cloudinit.NewUserWithAuthorizedKeys("git", authorizedKeys)
	gitUser.HomeDir = "/srv/git"

	userData := cloudinit.UserData{
		Hostname:      s.name,
		PackageUpdate: true,
		Packages:      []string{"git", "openssh-server", "qemu-guest-agent"},
		Users:         []cloudinit.User{gitUser},
		RunCommands: []string{
			"mkdir -p /srv/git",
			"chown -R git:git /srv/git",
			"chmod -R 755 /srv/git",
			"sed -i 's/^#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config",
			"sed -i 's/^PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config",
			"sed -i 's/^#PermitRootLogin prohibit-password/PermitRootLogin no/' /etc/ssh/sshd_config",
			"sed -i 's/^PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config",
			"systemctl restart sshd",
			"chsh -s /usr/bin/git-shell git",
		},
	}

	// 3. Populate s.vmConfig
	s.vmConfig = vmm.NewVMConfig(s.name, s.imageQCOW2Path, userData)

	return nil
}

func (s *Server) Teardown() error {
	if s.vmm == nil {
		return nil // Nothing to do if VMM was not initialized
	}

	var errs error
	if err := s.vmm.DestroyVM(s.vmConfig.Name); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to destroy VM: %w", err))
	}

	if err := s.vmm.Close(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to close VMM connection: %w", err))
	}

	if err := os.RemoveAll(s.tempDir); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to remove temp dir: %w", err))
	}

	if errs != nil {
		slog.Error(
			"encountered unexpected error while tearing down git server",
			"error",
			errs.Error(),
		)
	}

	return errs
}

func (s *Server) sshClient() (*ssh.Client, error) {
	sshClient, err := ssh.NewClient(
		s.ServerAddr,
		"git",
		s.initPrivateKeyPath,
		fmt.Sprintf("%d", s.SSHPort),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client for Git server: %w", err)
	}
	if err := sshClient.AwaitServer(30 * time.Second); err != nil {
		return nil, fmt.Errorf("Git server did not become ready in time: %w", err)
	}
	return sshClient, nil
}

func (s *Server) initAndPushRepo(sshClient *ssh.Client, repoName, srcPath string) error {
	// Initialize a bare Git repository on the server
	initCmd := fmt.Sprintf("git init --bare /srv/git/%s", repoName)
	if stdout, stderr, err := sshClient.Run(initCmd); err != nil {
		return fmt.Errorf(
			"failed to initialize bare repository on Git server: stdout=%s; stderr=%s; %w",
			stdout, stderr, err,
		)
	}

	// Push from the local source repository to the Git server
	// Create a temporary directory for the local repo working copy
	tempLocalRepoDir, err := os.MkdirTemp("", fmt.Sprintf("gitpush-%s-", repoName))
	if err != nil {
		return fmt.Errorf("failed to create temp local repo dir: %w", err)
	}
	defer os.RemoveAll(tempLocalRepoDir)

	// Copy the source repo to the temp directory
	cpCmd := exec.Command("cp", "-r", srcPath, filepath.Join(tempLocalRepoDir, "repo"))
	if output, err := cpCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy repo: %w\nOutput: %s", err, output)
	}

	tempRepoDirPath := filepath.Join(tempLocalRepoDir, "repo")

	// Initialize git if not already initialized
	if _, err := os.Stat(filepath.Join(tempRepoDirPath, ".git")); os.IsNotExist(err) {
		initGitCmd := exec.Command("git", "init")
		initGitCmd.Dir = tempRepoDirPath
		if output, err := initGitCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to git init local repo: %w\nOutput: %s", err, output)
		}

		// Configure git user
		configCmds := [][]string{
			{"git", "config", "user.email", "gitserver@example.com"},
			{"git", "config", "user.name", "Git Server"},
		}
		for _, args := range configCmds {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = tempRepoDirPath
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to configure git: %w\nOutput: %s", err, output)
			}
		}

		// Add and commit all files
		addCmd := exec.Command("git", "add", ".")
		addCmd.Dir = tempRepoDirPath
		if output, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to git add: %w\nOutput: %s", err, output)
		}

		commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
		commitCmd.Dir = tempRepoDirPath
		if output, err := commitCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to git commit: %w\nOutput: %s", err, output)
		}
	}

	// Add remote and push
	remoteURL := fmt.Sprintf("ssh://git@%s:%d/srv/git/%s", s.vmIPAddress, s.SSHPort, repoName)

	// Remove existing origin remote if it exists
	cmd = exec.Command("git", "remote", "remove", "origin")
	cmd.Dir = tempRepoDirPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// ignore errors, remote might not exist
	}

	// Add new remote
	cmd = exec.Command("git", "remote", "add", "origin", remoteURL)
	cmd.Dir = tempRepoDirPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add git remote: %w\nOutput: %s", err, output)
	}

	// Commit any uncommitted changes
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tempRepoDirPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// ignore errors, files might already be added
	}

	cmd = exec.Command("git", "commit", "-m", "Sync from source", "--allow-empty")
	cmd.Dir = tempRepoDirPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// ignore errors, might have no changes
	}

	// Push to the server
	cmd = exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = tempRepoDirPath
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf(
			"GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
			s.initPrivateKeyPath,
		),
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push repo %s to server: %w\nOutput: %s", repoName, err, output)
	}

	return nil
}

func (s *Server) GetRepoUrl(repoName string) string {
	return fmt.Sprintf("ssh://git@%s/srv/git/%s", s.ServerAddr, repoName)
}

func (s *Server) GetVMIPAddress() string {
	return s.vmIPAddress
}
