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

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
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
	ServerAddr     string
	SSHPort        int
	HostKeyPath    string
	AuthorizedKeys []string
	BaseDir        string
	Repo           []Repo

	// -- private
	authorizedKeysFile string
	buildDir           string
	gitDir             string
	imgName            string
	runningContainer   string
}

func (s *Server) Run() error {
	if err := s.init(); err != nil {
		return fmt.Errorf("failed to initialize git server: %v", err)
	}

	if err := s.buildImage(); err != nil {
		return fmt.Errorf("failed to build git server image: %v", err)
	}

	if err := s.setupAuthorizedKeys(); err != nil {
		return fmt.Errorf("failed to set up authorized keys: %w", err)
	}

	if err := s.start(); err != nil {
		return fmt.Errorf("failed to start git server: %w", err)
	}

	// -- init ssh client
	sshClient, err := s.sshClient()
	if err != nil {
		return fmt.Errorf("failed to ssh to git server: %w", err)
	}

	for _, repo := range s.Repo {
		srcPath, err := s.fetchRepo(repo)
		if err != nil {
			return fmt.Errorf("failed to start fetch git repo: %w", err)
		}

		if err := s.initAndPushRepo(sshClient, repo.Name, srcPath); err != nil {
			return fmt.Errorf("failed to init repo: %w", err)
		}
	}

	return nil
}

func (s *Server) init() error {
	s.imgName = "gitserver"
	baseDir := s.BaseDir
	if baseDir == "" {
		var err error
		baseDir, err = os.MkdirTemp("", "git-server")
		if err != nil {
			return err
		}
	}
	sshDir := filepath.Join(baseDir, "ssh")
	s.authorizedKeysFile = filepath.Join(sshDir, "authorized_keys")
	s.buildDir = filepath.Join(s.BaseDir, "build")
	s.gitDir = filepath.Join(s.BaseDir, "git")
	for _, s := range []string{sshDir, s.buildDir, s.gitDir} {
		if err := os.MkdirAll(s, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) buildImage() error {
	dockerfilePath := filepath.Join(s.buildDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(Dockerfile), 0o644); err != nil {
		return err
	}
	entrypointPath := filepath.Join(s.buildDir, "entrypoint.sh")
	if err := os.WriteFile(entrypointPath, []byte(Entrypoint), 0o644); err != nil {
		return err
	}
	// Build the Git server Docker image
	cmd := exec.Command(
		"docker",
		"build",
		"-t",
		s.imgName,
		s.buildDir,
	)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *Server) setupAuthorizedKeys() error {
	// Write public keys to authorized_keys
	authKeysContent := strings.Join(s.AuthorizedKeys, "\n")
	authKeysPath := s.authorizedKeysFile
	if err := os.WriteFile(authKeysPath, []byte(authKeysContent), 0o644); err != nil {
		return fmt.Errorf("failed to write authorized_keys: %w", err)
	}

	return nil
}

func (s *Server) start() error {
	// Run the Git server container
	containerName := fmt.Sprintf("gitserver-%d", time.Now().UnixNano())
	cmd := exec.Command(
		"docker", "run", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("%d:22", s.SSHPort),
		"-v", fmt.Sprintf("%s:/srv/git:ro", s.gitDir),
		"-v", fmt.Sprintf("%s:/tmp/authorized_keys:ro", s.authorizedKeysFile),
		s.imgName,
	)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	s.runningContainer = containerName
	return nil
}

func (s *Server) Teardown() error {
	if s.runningContainer == "" {
		return errors.New("failed to teardown git server: no container running")
	}
	errs := exec.Command("docker", "stop", s.runningContainer).Run()
	err := exec.Command("docker", "rm", s.runningContainer).Run()
	errs = errors.Join(errs, err)
	if errs != nil {
		slog.Error("encountered unexpected error while tearing down git server",
			"error", err.Error())
	}
	s.runningContainer = ""
	return nil
}

func (s *Server) sshClient() (*ssh.Client, error) {
	gitServerAddr := fmt.Sprintf("%s:%d", s.ServerAddr, s.SSHPort)
	sshClient, err := ssh.NewClient(gitServerAddr, "git", s.HostKeyPath, "22")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client for Git server: %w", err)
	}
	if err := sshClient.AwaitServer(30 * time.Second); err != nil {
		return nil, fmt.Errorf("Git server did not become ready in time: %w", err)
	}
	return sshClient, nil
}

func (s *Server) fetchRepo(repo Repo) (string, error) {
	switch repo.Source.Type {
	default:
		panic("not implemented")
	case LocalSource:
		return filepath.Join(s.gitDir, repo.Name), nil
	case GitUrlSource:
		panic("not implemented either")
	}
}

func (s *Server) initAndPushRepo(sshClient *ssh.Client, repoName, srcPath string) error {
	// Initialize a bare Git repository on the server
	initCmd := fmt.Sprintf("git init --bare /srv/git/%s", repoName)
	if _, _, err := sshClient.Run(initCmd); err != nil {
		return fmt.Errorf("failed to initialize bare repository on Git server: %w", err)
	}

	destPath := filepath.Join(s.gitDir, repoName)
	cmd := exec.Command("cp", "-r", filepath.Join(srcPath, "."), destPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy package manager configs: %w\nOutput: %s", err, output)
	}

	if err := os.RemoveAll(filepath.Join(destPath, ".git")); err != nil {
		return err
	}

	// Initialize local git repo and push
	cmd = exec.Command("git", "init")
	cmd.Dir = destPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git init local repo: %w\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = destPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git add local repo: %w\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = destPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git commit local repo: %w\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "remote", "add", "origin", s.GetRepoUrl(repoName))
	cmd.Dir = destPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add git remote: %w\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = destPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git push to remote: %w\nOutput: %s", err, output)
	}

	return nil
}

func (s *Server) GetRepoUrl(repoName string) string {
	return fmt.Sprintf("ssh://git@%s/srv/git/%s", s.ServerAddr, repoName)
}
