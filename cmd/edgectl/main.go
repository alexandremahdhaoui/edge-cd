package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
)

var (
	errCreateSSHClient     = errors.New("failed to create SSH client")
	errCreateTempDir       = errors.New("failed to create temporary directory")
	errCloneLocalRepo      = errors.New("failed to clone edge-cd repository locally")
	errProvisionPackages   = errors.New("failed to provision packages")
	errCloneUserConfigRepo = errors.New("failed to clone user config repo")
	errReadLocalConfig     = errors.New("failed to read local config")
	errRenderConfig        = errors.New("failed to render config template")
	errPlaceConfig         = errors.New("failed to place config.yaml")
	errSetupService        = errors.New("failed to setup edge-cd service")
)

func main() {
	// Define a new FlagSet for the root command
	rootCmd := flag.NewFlagSet("edgectl", flag.ExitOnError)
	rootCmd.Usage = func() {
		fmt.Fprintf(rootCmd.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(rootCmd.Output(), "  %s <command> [arguments]\n", os.Args[0])
		fmt.Fprintf(rootCmd.Output(), "The commands are:\n")
		fmt.Fprintf(rootCmd.Output(), "  bootstrap   Bootstrap an edge device\n")
		rootCmd.PrintDefaults()
	}

	// Parse the root command flags (if any, though none are defined yet)
	rootCmd.Parse(os.Args[1:])

	// Check if a subcommand was provided
	if rootCmd.NArg() == 0 {
		rootCmd.Usage()
		os.Exit(1)
	}

	// Get the subcommand
	cmd := rootCmd.Arg(0)

	switch cmd {
	case "bootstrap":
		bootstrapCmd := flag.NewFlagSet("bootstrap", flag.ExitOnError)

		// Define flags for the bootstrap command
		targetAddr := bootstrapCmd.String(
			"target-addr",
			"",
			"Target device address (e.g., user@host or host)",
		)
		targetUser := bootstrapCmd.String("target-user", "root", "SSH user for the target device")
		sshPrivateKey := bootstrapCmd.String(
			"ssh-private-key",
			"",
			"Path to the SSH private key (required)",
		)
		configRepo := bootstrapCmd.String(
			"config-repo",
			"",
			"URL of the configuration Git repository (required)",
		)
		configPath := bootstrapCmd.String(
			"config-path",
			"",
			"Path to the directory containing the config spec file",
		)
		configSpec := bootstrapCmd.String("config-spec", "", "Name of the config spec file")
		edgeCDRepo := bootstrapCmd.String(
			"edge-cd-repo",
			"https://github.com/alexandremahdhaoui/edge-cd.git",
			"URL of the edge-cd Git repository",
		)
		edgeCDBranch := bootstrapCmd.String(
			"edgecd-branch",
			"main",
			"Branch name for the edge-cd repository (default: main)",
		)
		configBranch := bootstrapCmd.String(
			"config-branch",
			"main",
			"Branch name for the config repository (default: main)",
		)
		packages := bootstrapCmd.String(
			"packages",
			"",
			"Comma-separated list of packages to install",
		)
		serviceManager := bootstrapCmd.String(
			"service-manager",
			"prodc",
			"Service manager to use (e.g., 'prodc', 'systemd')",
		)
		packageManager := bootstrapCmd.String(
			"package-manager",
			"opkg",
			"Package manager to use (e.g., 'opkg', 'apt')",
		)
		edgeCDRepoDestPath := bootstrapCmd.String(
			"edge-cd-repo-dest",
			"/usr/local/src/edge-cd",
			"Destination path for edge-cd repository on target device",
		)
		userConfigRepoDestPath := bootstrapCmd.String(
			"user-config-repo-dest",
			"/usr/local/src/edge-cd-config",
			"Destination path for user config repository on target device",
		)
		injectEnv := bootstrapCmd.String(
			"inject-env",
			"",
			"Environment variables to inject to target (e.g., 'GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no')",
		)

		bootstrapCmd.Usage = func() {
			fmt.Fprintf(bootstrapCmd.Output(), "Usage of %s bootstrap:\n", os.Args[0])
			fmt.Fprintf(bootstrapCmd.Output(), "  Bootstrap an edge device.\n\n")
			fmt.Fprintf(bootstrapCmd.Output(), "Flags:\n")
			bootstrapCmd.PrintDefaults()
		}
		bootstrapCmd.Parse(rootCmd.Args()[1:])

		// Validate required flags
		if *targetAddr == "" {
			fmt.Fprintf(os.Stderr, "Error: --target-addr is required\n")
			bootstrapCmd.Usage()
			os.Exit(1)
		}

		if *configRepo == "" {
			fmt.Fprintf(os.Stderr, "Error: --config-repo is required\n")
			bootstrapCmd.Usage()
			os.Exit(1)
		}

		if *sshPrivateKey == "" {
			fmt.Fprintf(os.Stderr, "Error: --ssh-private-key is required\n")
			bootstrapCmd.Usage()
			os.Exit(1)
		}

		// SSH Client
		sshClient, err := ssh.NewClient(*targetAddr, *targetUser, *sshPrivateKey, "22")
		if err != nil {
			slog.Error(
				"bootstrap failed",
				"error",
				flaterrors.Join(err, errCreateSSHClient).Error(),
			)
			os.Exit(1)
		}

		// Define remote paths (from flags or defaults)
		remoteEdgeCDRepoDestPath := *edgeCDRepoDestPath
		userConfigRepoPath := *userConfigRepoDestPath

		// Create execution contexts
		// Build environment variables map
		targetInjectedEnvs := make(map[string]string)

		// Add injected environment variables if provided
		if *injectEnv != "" {
			envKey, envValue := parseEnvFromFlag(*injectEnv)
			if envKey != "" {
				targetInjectedEnvs[envKey] = envValue
			}
		}

		// Create contexts using the immutable factory function
		// targetExecCtx: for remote commands requiring privilege escalation (sudo -E)
		targetExecCtx := execcontext.New(targetInjectedEnvs, []string{"sudo", "-E"})

		// Clone edge-cd repo locally to get package manager configs
		localEdgeCDRepoTempDir, err := os.MkdirTemp("", "edgectl-local-edge-cd-repo-")
		if err != nil {
			slog.Error("bootstrap failed", "error", flaterrors.Join(err, errCreateTempDir).Error())
			os.Exit(1)
		}
		defer os.RemoveAll(localEdgeCDRepoTempDir) // Clean up temp directory

		localCloneCmd := exec.Command(
			"git",
			"clone",
			"-b",
			*edgeCDBranch,
			*edgeCDRepo,
			localEdgeCDRepoTempDir,
		)
		localCloneCmd.Stdout = os.Stderr
		localCloneCmd.Stderr = os.Stderr
		if err := localCloneCmd.Run(); err != nil {
			slog.Error("bootstrap failed", "error", flaterrors.Join(err, errCloneLocalRepo).Error())
			os.Exit(1)
		}

		// type yolo struct {
		//	TargetExecCtx execcontext.Context
		//	LocalExecCtx execcontext.Context
		//}

		// Package Provisioning
		pkgs := strings.Split(*packages, ",")
		if len(pkgs) > 0 {
			if err := provision.ProvisionPackages(targetExecCtx, sshClient, pkgs, *packageManager, localEdgeCDRepoTempDir, *edgeCDRepo, remoteEdgeCDRepoDestPath); err != nil {
				slog.Error(
					"bootstrap failed",
					"error",
					flaterrors.Join(err, errProvisionPackages).Error(),
				)
				os.Exit(1)
			}
		}

		configGitRepo := provision.GitRepo{
			URL:    *configRepo,
			Branch: *configBranch,
		}
		if err := provision.CloneOrPullRepo(targetExecCtx, sshClient, userConfigRepoPath, configGitRepo); err != nil {
			slog.Error(
				"bootstrap failed",
				"error",
				flaterrors.Join(err, errCloneUserConfigRepo).Error(),
			)
			os.Exit(1)
		}

		// Config Placement
		var configContent string
		if *configPath != "" && *configSpec != "" {
			configContent, err = provision.ReadLocalConfig(*configPath, *configSpec)
			if err != nil {
				slog.Error(
					"bootstrap failed",
					"error",
					flaterrors.Join(err, errReadLocalConfig).Error(),
				)
				os.Exit(1)
			}
		} else {
			configData := provision.ConfigTemplateData{
				EdgeCDRepoURL:      *edgeCDRepo,
				ConfigRepoURL:      *configRepo,
				ServiceManagerName: *serviceManager,
				PackageManagerName: *packageManager,
				RequiredPackages:   pkgs,
			}
			configContent, err = provision.RenderConfig(configData)
			if err != nil {
				slog.Error("bootstrap failed", "error", flaterrors.Join(err, errRenderConfig).Error())
				os.Exit(1)
			}
		}

		if err := provision.PlaceConfigYAML(targetExecCtx, sshClient, configContent, "/etc/edge-cd/config.yaml"); err != nil {
			slog.Error("bootstrap failed", "error", flaterrors.Join(err, errPlaceConfig).Error())
			os.Exit(1)
		}

		// Service Setup
		if err := provision.SetupEdgeCDService(targetExecCtx, sshClient, *serviceManager, localEdgeCDRepoTempDir, remoteEdgeCDRepoDestPath); err != nil {
			slog.Error("bootstrap failed", "error", flaterrors.Join(err, errSetupService).Error())
			os.Exit(1)
		}

		slog.Info("bootstrap completed successfully")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		rootCmd.Usage()
		os.Exit(1)
	}
}

// parseEnvFromFlag parses an environment variable string in the format "KEY=value"
// and returns the key and value separately. If the format is invalid, it returns empty strings.
func parseEnvFromFlag(envVar string) (key, value string) {
	parts := strings.SplitN(envVar, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
