package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
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
		targetAddr := bootstrapCmd.String("target-addr", "", "Target device address (e.g., user@host or host)")
		targetUser := bootstrapCmd.String("target-user", "root", "SSH user for the target device")
		sshPrivateKey := bootstrapCmd.String("ssh-private-key", "", "Path to the SSH private key (required)")
		configRepo := bootstrapCmd.String("config-repo", "", "URL of the configuration Git repository (required)")
		configPath := bootstrapCmd.String("config-path", "", "Path to the directory containing the config spec file")
		configSpec := bootstrapCmd.String("config-spec", "", "Name of the config spec file")
		edgeCDRepo := bootstrapCmd.String("edge-cd-repo", "https://github.com/alexandremahdhaoui/edge-cd.git", "URL of the edge-cd Git repository")
		edgeCDBranch := bootstrapCmd.String("edgecd-branch", "main", "Branch name for the edge-cd repository (default: main)")
		configBranch := bootstrapCmd.String("config-branch", "main", "Branch name for the config repository (default: main)")
		packages := bootstrapCmd.String("packages", "", "Comma-separated list of packages to install")
		serviceManager := bootstrapCmd.String("service-manager", "prodc", "Service manager to use (e.g., 'prodc', 'systemd')")
		packageManager := bootstrapCmd.String("package-manager", "opkg", "Package manager to use (e.g., 'opkg', 'apt')")
		edgeCDRepoDestPath := bootstrapCmd.String("edge-cd-repo-dest", "/usr/local/src/edge-cd", "Destination path for edge-cd repository on target device")
		userConfigRepoDestPath := bootstrapCmd.String("user-config-repo-dest", "/usr/local/src/edge-cd-config", "Destination path for user config repository on target device")
		prependCmd := bootstrapCmd.String("inject-prepend-cmd", "", "Command to prepend to privileged operations (e.g., 'sudo')")
		injectEnv := bootstrapCmd.String("inject-env", "", "Environment variables to inject to target (e.g., 'GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no')")

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
			log.Fatalf("Failed to create SSH client: %v", err)
		}

		// Define remote paths (from flags or defaults)
		remoteEdgeCDRepoDestPath := *edgeCDRepoDestPath
		userConfigRepoPath := *userConfigRepoDestPath

		// Clone edge-cd repo locally to get package manager configs
		localEdgeCDRepoTempDir, err := os.MkdirTemp("", "edgectl-local-edge-cd-repo-")
		if err != nil {
			log.Fatalf("Failed to create temporary directory for local edge-cd repo: %v", err)
		}
		defer os.RemoveAll(localEdgeCDRepoTempDir) // Clean up temp directory

		localCloneCmd := exec.Command("git", "clone", "-b", *edgeCDBranch, *edgeCDRepo, localEdgeCDRepoTempDir)
		localCloneCmd.Stdout = os.Stderr
		localCloneCmd.Stderr = os.Stderr
		// Use the environment as-is, which should include GIT_SSH_COMMAND if set by the caller
		// If GIT_SSH_COMMAND is not set, add a default with SSH options
		gitSSHCmd := os.Getenv("GIT_SSH_COMMAND")
		if gitSSHCmd == "" {
			gitSSHCmd = "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
		}
		localCloneCmd.Env = append(
			os.Environ(),
			fmt.Sprintf("GIT_SSH_COMMAND=%s", gitSSHCmd),
		)
		if err := localCloneCmd.Run(); err != nil {
			log.Fatalf("Failed to clone edge-cd repository locally: %v", err)
		}

		// Prepare GIT_SSH_COMMAND for the remote git operations (gitSSHCmd already defined above)
		gitsshEnv := fmt.Sprintf("GIT_SSH_COMMAND='%s'", gitSSHCmd)

		// Merge with injected environment variables if provided
		if *injectEnv != "" {
			gitsshEnv = gitsshEnv + " " + *injectEnv
		}

		// Package Provisioning
		pkgs := strings.Split(*packages, ",")
		if len(pkgs) > 0 {
			if err := provision.ProvisionPackagesWithEnv(sshClient, pkgs, *packageManager, localEdgeCDRepoTempDir, *edgeCDRepo, remoteEdgeCDRepoDestPath, gitsshEnv); err != nil {
				log.Fatalf("Failed to provision packages: %v", err)
			}
		}

		// Repo Cloning (only user config repo needs to be cloned here, edge-cd repo is handled in ProvisionPackages)
		if err := provision.CloneOrPullRepoWithBranchAndEnv(sshClient, *configRepo, userConfigRepoPath, *configBranch, gitsshEnv, *prependCmd); err != nil {
			log.Fatalf("Failed to clone user config repo: %v", err)
		}

		// Config Placement
		var configContent string
		if *configPath != "" && *configSpec != "" {
			configContent, err = provision.ReadLocalConfig(*configPath, *configSpec)
			if err != nil {
				log.Fatalf("Failed to read local config: %v", err)
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
				log.Fatalf("Failed to render config template: %v", err)
			}
		}

		if _, _, err := sshClient.Run("sudo mkdir -p /etc/edge-cd && sudo chown ubuntu:ubuntu /etc/edge-cd && sudo chmod 755 /etc/edge-cd"); err != nil {
			// ignore error if directory already exists
		}

		if err := provision.PlaceConfigYAML(sshClient, configContent, "/etc/edge-cd/config.yaml"); err != nil {
			log.Fatalf("Failed to place config.yaml: %v", err)
		}

		// Service Setup
		if err := provision.SetupEdgeCDService(sshClient, *serviceManager, remoteEdgeCDRepoDestPath, *prependCmd); err != nil {
			log.Fatalf("Failed to setup edge-cd service: %v", err)
		}

		fmt.Println("Bootstrap completed successfully!")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		rootCmd.Usage()
		os.Exit(1)
	}
}
