package main

import (
	"flag"
	"fmt"
	"os"
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
		edgeCDRepo := bootstrapCmd.String("edge-cd-repo", "https://github.com/alexandremahdhaoui/edge-cd.git", "URL of the edge-cd Git repository")
		packages := bootstrapCmd.String("packages", "", "Comma-separated list of packages to install")
		serviceManager := bootstrapCmd.String("service-manager", "procd", "Service manager to use (e.g., 'procd', 'systemd')")

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
		if *configRepo == "" {
			fmt.Fprintf(os.Stderr, "Error: --config-repo is required\n")
			bootstrapCmd.Usage()
			os.Exit(1)
		}

		fmt.Printf("Running bootstrap command with flags:\n")
		fmt.Printf("  Target: %s\n", *target)
		fmt.Printf("  User: %s\n", *user)
		fmt.Printf("  Key: %s\n", *key)
		fmt.Printf("  Config Repo: %s\n", *configRepo)
		fmt.Printf("  Edge-CD Repo: %s\n", *edgeCDRepo)
		fmt.Printf("  Packages: %s\n", *packages)
		fmt.Printf("  Service Manager: %s\n", *serviceManager)

		// Actual bootstrap logic will go here in future tasks
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		rootCmd.Usage()
		os.Exit(1)
	}
}