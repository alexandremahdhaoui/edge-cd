package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
		target := bootstrapCmd.String("target", "", "Target device (user@host) (required)")
		user := bootstrapCmd.String("user", "root", "SSH user for the target device")
		key := bootstrapCmd.String("key", "~/.ssh/id_rsa", "Path to the SSH private key")
		configRepo := bootstrapCmd.String("config-repo", "", "URL of the configuration Git repository (required)")
		configPath := bootstrapCmd.String("config-path", "", "Path to the directory containing the config spec file")
		configSpec := bootstrapCmd.String("config-spec", "", "Name of the config spec file")
		edgeCDRepo := bootstrapCmd.String("edge-cd-repo", "https://github.com/alexandremahdhaoui/edge-cd.git", "URL of the edge-cd Git repository")
		packages := bootstrapCmd.String("packages", "", "Comma-separated list of packages to install")
		serviceManager := bootstrapCmd.String("service-manager", "procd", "Service manager to use (e.g., 'procd', 'systemd')")
		packageManager := bootstrapCmd.String("package-manager", "opkg", "Package manager to use (e.g., 'opkg', 'apt')")

		bootstrapCmd.Usage = func() {
			fmt.Fprintf(bootstrapCmd.Output(), "Usage of %s bootstrap:\n", os.Args[0])
			fmt.Fprintf(bootstrapCmd.Output(), "  Bootstrap an edge device.\n\n")
			fmt.Fprintf(bootstrapCmd.Output(), "Flags:\n")
			bootstrapCmd.PrintDefaults()
		}
		bootstrapCmd.Parse(rootCmd.Args()[1:])

		// Validate required flags
		if *target == "" {
			fmt.Fprintf(os.Stderr, "Error: --target is required\n")
			bootstrapCmd.Usage()
			os.Exit(1)
		}

		// Split target into user and host
		targetParts := strings.Split(*target, "@")
		var targetUser, targetHost string
		if len(targetParts) == 2 {
			targetUser = targetParts[0]
			targetHost = targetParts[1]
		} else {
			targetUser = *user
			targetHost = *target
		}

		// SSH Client
		sshClient, err := ssh.NewClient(targetHost, targetUser, *key, "22")
		if err != nil {
			log.Fatalf("Failed to create SSH client: %v", err)
		}

		// Package Provisioning
		pkgs := strings.Split(*packages, ",")
		if len(pkgs) > 0 {
			if err := provision.ProvisionPackages(sshClient, pkgs, *packageManager, "./cmd/edge-cd/package-managers"); err != nil {
				log.Fatalf("Failed to provision packages: %v", err)
			}
		}

		// Repo Cloning
		const edgeCDRepoPath = "/opt/edge-cd"
		const userConfigRepoPath = "/opt/user-config"

		if err := provision.CloneOrPullRepo(sshClient, *edgeCDRepo, edgeCDRepoPath); err != nil {
			log.Fatalf("Failed to clone edge-cd repo: %v", err)
		}
		if err := provision.CloneOrPullRepo(sshClient, *configRepo, userConfigRepoPath); err != nil {
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

		if err := sshClient.Run("mkdir -p /etc/edge-cd"); err != nil {
			// ignore error if directory already exists
		}

		if err := provision.PlaceConfigYAML(sshClient, configContent, "/etc/edge-cd/config.yaml"); err != nil {
			log.Fatalf("Failed to place config.yaml: %v", err)
		}

		// Service Setup
		if err := provision.SetupEdgeCDService(sshClient, *serviceManager, edgeCDRepoPath); err != nil {
			log.Fatalf("Failed to setup edge-cd service: %v", err)
		}

		fmt.Println("Bootstrap completed successfully!")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		rootCmd.Usage()
		os.Exit(1)
	}
}