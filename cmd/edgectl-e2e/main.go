package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	te2e "github.com/alexandremahdhaoui/edge-cd/pkg/test/e2e"
)

func main() {
	// Create a new flag set for this tool
	fs := flag.NewFlagSet("edgectl-e2e", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: edgectl-e2e [command] [options]

Commands:
  create             Create a new test environment
  get <test-id>      Get information about a test environment
  run <test-id>      Run tests in an existing environment
  delete <test-id>   Cleanup and destroy a test environment
  list               List all known test environments and their status
  test               One-shot test (create → run → delete)

Environment Variables:
  E2E_ARTIFACTS_DIR  Override artifact storage location (default: ~/.edge-cd/e2e/)

Examples:
  # Create test environment
  edgectl-e2e create

  # Get environment information
  edgectl-e2e get e2e-20231025-abc123

  # Run tests in that environment
  edgectl-e2e run e2e-20231025-abc123

  # Cleanup when done
  edgectl-e2e delete e2e-20231025-abc123

  # List all environments
  edgectl-e2e list

  # One-shot test
  edgectl-e2e test
`)
	}

	if len(os.Args) < 2 {
		fs.Usage()
		os.Exit(1)
	}

	command := os.Args[1]
	artifactStoreDir := getArtifactDir()

	execCtx := execcontext.New(make(map[string]string), []string{})

	switch command {
	case "create":
		cmdCreate(execCtx, artifactStoreDir)
	case "get":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: 'get' requires a test ID\n")
			fmt.Fprintf(os.Stderr, "Usage: edgectl-e2e get <test-id>\n")
			os.Exit(1)
		}
		cmdGet(execCtx, artifactStoreDir, os.Args[2])
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: 'run' requires a test ID\n")
			fmt.Fprintf(os.Stderr, "Usage: edgectl-e2e run <test-id>\n")
			os.Exit(1)
		}
		cmdRun(execCtx, artifactStoreDir, os.Args[2])
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: 'delete' requires a test ID\n")
			fmt.Fprintf(os.Stderr, "Usage: edgectl-e2e delete <test-id>\n")
			os.Exit(1)
		}
		cmdDelete(execCtx, artifactStoreDir, os.Args[2])
	case "list":
		cmdList(execCtx, artifactStoreDir)
	case "test":
		cmdTest(execCtx, artifactStoreDir)
	case "-h", "--help", "help":
		fs.Usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		fs.Usage()
		os.Exit(1)
	}
}

// getArtifactDir returns the artifact storage directory
func getArtifactDir() string {
	if dir := os.Getenv("E2E_ARTIFACTS_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.ExpandEnv("$HOME"), ".edge-cd", "e2e")
}

// getEdgeCDRepoPath returns the path to the edge-cd repository
func getEdgeCDRepoPath() string {
	// Try to find the repo root relative to current directory
	cwd, err := os.Getwd()
	if err == nil {
		// If we're in a subdirectory, try to find the git root
		for {
			if _, err := os.Stat(filepath.Join(cwd, ".git")); err == nil {
				return cwd
			}
			parent := filepath.Dir(cwd)
			if parent == cwd {
				break // reached filesystem root
			}
			cwd = parent
		}
	}

	// Fallback: use go command to find repo root
	// This assumes the CLI is run from within the edge-cd repository
	// or as a compiled binary in a specific location
	return "."
}

// cmdCreate creates and provisions a complete test environment with VMs
func cmdCreate(
	execCtx execcontext.Context,
	artifactStoreDir string,
) {
	// Get paths
	cacheDir := filepath.Join(os.TempDir(), "edgectl")
	edgeCDRepoPath := getEdgeCDRepoPath()

	// Setup configuration
	setupConfig := te2e.SetupConfig{
		ArtifactDir:    filepath.Join(artifactStoreDir, "artifacts"),
		ImageCacheDir:  cacheDir,
		EdgeCDRepoPath: edgeCDRepoPath,
		DownloadImages: true,
	}

	// Create test environment with VMs
	fmt.Fprintf(os.Stderr, "Creating test environment...\n")
	testEnv, err := te2e.SetupTestEnvironment(execCtx, setupConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create test environment: %v\n", err)
		os.Exit(1)
	}

	// Save to artifact store
	if err := os.MkdirAll(artifactStoreDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create artifact store directory: %v\n", err)
		os.Exit(1)
	}

	store := te2e.NewJSONArtifactStore(filepath.Join(artifactStoreDir, "artifacts.json"))
	if err := store.Save(execCtx, testEnv); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save environment to artifact store: %v\n", err)
		os.Exit(1)
	}

	// Output environment ID (this is the primary output for scripting)
	fmt.Println(testEnv.ID)

	// Print summary if not piped
	if !isPiped() {
		fmt.Fprintf(os.Stderr, "\n✅ Test environment created and provisioned: %s\n", testEnv.ID)
		fmt.Fprintf(os.Stderr, "   Status: %s\n", testEnv.Status)
		fmt.Fprintf(os.Stderr, "   Artifacts dir: %s\n", testEnv.ArtifactPath)
		fmt.Fprintf(os.Stderr, "\n=== Target VM ===\n")
		fmt.Fprintf(os.Stderr, "   Name: %s\n", testEnv.TargetVM.Name)
		fmt.Fprintf(os.Stderr, "   IP: %s\n", testEnv.TargetVM.IP)
		fmt.Fprintf(
			os.Stderr,
			"   SSH: ssh -i %s ubuntu@%s\n",
			testEnv.SSHKeys.HostKeyPath,
			testEnv.TargetVM.IP,
		)
		fmt.Fprintf(os.Stderr, "\n=== Git Server VM ===\n")
		fmt.Fprintf(os.Stderr, "   Name: %s\n", testEnv.GitServerVM.Name)
		fmt.Fprintf(os.Stderr, "   IP: %s\n", testEnv.GitServerVM.IP)
		fmt.Fprintf(
			os.Stderr,
			"   SSH: ssh -i %s git@%s\n",
			testEnv.SSHKeys.HostKeyPath,
			testEnv.GitServerVM.IP,
		)
		fmt.Fprintf(os.Stderr, "\n=== Git Repositories ===\n")
		for repoName, repoURL := range testEnv.GitSSHURLs {
			fmt.Fprintf(os.Stderr, "   %s: %s\n", repoName, repoURL)
		}
		fmt.Fprintf(os.Stderr, "\nNext: Run tests with:\n")
		fmt.Fprintf(os.Stderr, "   edgectl-e2e run %s\n", testEnv.ID)
	}
}

// cmdRun executes bootstrap tests in an existing environment
func cmdRun(ctx execcontext.Context, artifactStoreDir string, testID string) {
	artifactStoreFile := filepath.Join(artifactStoreDir, "artifacts.json")
	store := te2e.NewJSONArtifactStore(artifactStoreFile)

	// Load environment
	env, err := store.Load(ctx, testID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load test environment: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running tests in environment: %s\n", env.ID)
	fmt.Printf("Status: %s\n", env.Status)
	fmt.Printf("Artifact Path: %s\n", env.ArtifactPath)

	// Validate environment has VMs
	if env.TargetVM.Name == "" {
		fmt.Fprintf(os.Stderr, "Error: Target VM not found in environment\n")
		os.Exit(1)
	}
	if env.GitServerVM.Name == "" {
		fmt.Fprintf(os.Stderr, "Error: Git server VM not found in environment\n")
		os.Exit(1)
	}

	fmt.Printf("Target VM: %s (IP: %s)\n", env.TargetVM.Name, env.TargetVM.IP)
	fmt.Printf("Git Server: %s (IP: %s)\n", env.GitServerVM.Name, env.GitServerVM.IP)

	// Build edgectl binary
	binaryPath, err := te2e.BuildEdgectlBinary("./cmd/edgectl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to build edgectl binary: %v\n", err)
		os.Exit(1)
	}

	// Execute bootstrap test
	executorConfig := te2e.ExecutorConfig{
		EdgectlBinaryPath: binaryPath,
		ConfigPath:        "./test/edgectl/e2e/config",
		ConfigSpec:        "config.yaml",
		Packages:          "git,curl,openssh-client",
		ServiceManager:    "systemd",
		PackageManager:    "apt",
	}

	fmt.Printf("Executing bootstrap tests...\n")
	if err := te2e.ExecuteBootstrapTest(ctx, env, executorConfig); err != nil {
		env.Status = "failed"
		store.Save(ctx, env)
		fmt.Fprintf(os.Stderr, "Error: bootstrap tests failed: %v\n", err)
		os.Exit(1)
	}

	// Update status to passed
	env.Status = "passed"
	if err := store.Save(ctx, env); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update environment status: %v\n", err)
	}

	if !isPiped() {
		fmt.Fprintf(os.Stderr, "\n✅ Tests passed in environment: %s\n", testID)
	}
}

// cmdDelete destroys a test environment and cleans up all resources
func cmdDelete(ctx execcontext.Context, artifactStoreDir string, testID string) {
	artifactStoreFile := filepath.Join(artifactStoreDir, "artifacts.json")
	store := te2e.NewJSONArtifactStore(artifactStoreFile)

	// Load environment
	env, err := store.Load(ctx, testID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load test environment: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deleting test environment: %s\n", env.ID)

	// Safety check: verify temp directory exists and is marked as managed
	if env.TempDirRoot != "" {
		fmt.Printf("  Validating temp directory: %s\n", env.TempDirRoot)
		if !te2e.IsManagedTempDirectory(env.TempDirRoot) {
			fmt.Fprintf(
				os.Stderr,
				"Warning: temp directory is not marked as managed (%s), proceeding with caution\n",
				env.TempDirRoot,
			)
		}
	}

	// Audit: log what will be deleted
	if len(env.ManagedResources) > 0 {
		fmt.Printf("  Deleting %d managed resources\n", len(env.ManagedResources))
		for _, resource := range env.ManagedResources {
			debugf("  - %s\n", resource)
		}
	}

	// Use the reusable teardown function with logging
	// Important: capture the error to determine if cleanup was successful
	teardownErr := te2e.TeardownTestEnvironmentWithLogging(ctx, env)

	// Determine deletion strategy based on teardown success
	if teardownErr == nil {
		// SUCCESS: All cleanup operations completed without errors
		// Safe to delete from artifact store
		if err := store.Delete(ctx, testID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete environment from store: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Environment removed from store\n")
		fmt.Printf("\n✅ Test environment %s has been fully deleted\n", testID)

	} else {
		// PARTIAL/COMPLETE FAILURE: Some cleanup operations failed
		// Do NOT delete from artifact store, instead mark as partially_deleted
		// so the user knows cleanup was incomplete
		fmt.Fprintf(os.Stderr, "\n⚠️  Cleanup encountered errors. Marking environment as partially_deleted.\n")
		fmt.Fprintf(os.Stderr, "Error details:\n%v\n", teardownErr)

		env.Status = te2e.StatusPartiallyDeleted
		env.UpdatedAt = time.Now()
		if err := store.Save(ctx, env); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to update environment status: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("  ✓ Environment marked as %s in store\n", te2e.StatusPartiallyDeleted)
		fmt.Printf("\n⚠️  Test environment %s marked as partially_deleted - cleanup had errors\n", testID)
		fmt.Fprintf(os.Stderr, "Please review the errors above and manually clean up if necessary.\n")
		fmt.Fprintf(os.Stderr, "To retry cleanup, run: edgectl-e2e delete %s\n", testID)

		os.Exit(1)
	}
}

// cmdGet displays complete information about a test environment
func cmdGet(ctx execcontext.Context, artifactStoreDir string, testID string) {
	artifactStoreFile := filepath.Join(artifactStoreDir, "artifacts.json")
	store := te2e.NewJSONArtifactStore(artifactStoreFile)

	// Load environment
	env, err := store.Load(ctx, testID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load test environment: %v\n", err)
		os.Exit(1)
	}

	// Display environment information
	fmt.Fprintf(os.Stderr, "\n=== Test Environment: %s ===\n", env.ID)
	fmt.Fprintf(os.Stderr, "Status: %s\n", env.Status)
	fmt.Fprintf(os.Stderr, "Created: %s\n", env.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(os.Stderr, "Artifacts: %s\n\n", env.ArtifactPath)

	fmt.Fprintf(os.Stderr, "=== Target VM ===\n")
	fmt.Fprintf(os.Stderr, "Name: %s\n", env.TargetVM.Name)
	fmt.Fprintf(os.Stderr, "IP: %s\n", env.TargetVM.IP)
	if env.TargetVM.IP != "" {
		fmt.Fprintf(
			os.Stderr,
			"SSH: ssh -i %s ubuntu@%s\n",
			env.SSHKeys.HostKeyPath,
			env.TargetVM.IP,
		)
	}
	fmt.Fprintf(os.Stderr, "Memory: %dMiB\n", env.TargetVM.MemoryMB)
	fmt.Fprintf(os.Stderr, "vCPUs: %d\n\n", env.TargetVM.VCPUs)

	fmt.Fprintf(os.Stderr, "=== Git Server VM ===\n")
	fmt.Fprintf(os.Stderr, "Name: %s\n", env.GitServerVM.Name)
	fmt.Fprintf(os.Stderr, "IP: %s\n", env.GitServerVM.IP)
	if env.GitServerVM.IP != "" {
		fmt.Fprintf(
			os.Stderr,
			"SSH: ssh -i %s git@%s\n",
			env.SSHKeys.HostKeyPath,
			env.GitServerVM.IP,
		)
	}
	fmt.Fprintf(os.Stderr, "Memory: %dMiB\n", env.GitServerVM.MemoryMB)
	fmt.Fprintf(os.Stderr, "vCPUs: %d\n\n", env.GitServerVM.VCPUs)

	if len(env.GitSSHURLs) > 0 {
		fmt.Fprintf(os.Stderr, "=== Git Repositories ===\n")
		for repoName, repoURL := range env.GitSSHURLs {
			fmt.Fprintf(os.Stderr, "%s: %s\n", repoName, repoURL)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(os.Stderr, "=== SSH Keys ===\n")
	fmt.Fprintf(os.Stderr, "Host Key: %s\n", env.SSHKeys.HostKeyPath)
	fmt.Fprintf(os.Stderr, "Host Pub: %s\n", env.SSHKeys.HostKeyPubPath)
}

// cmdList lists all test environments
func cmdList(ctx execcontext.Context, artifactStoreDir string) {
	artifactStoreFile := filepath.Join(artifactStoreDir, "artifacts.json")
	store := te2e.NewJSONArtifactStore(artifactStoreFile)

	// Load all environments
	envs, err := store.ListAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list environments: %v\n", err)
		os.Exit(1)
	}

	if len(envs) == 0 {
		fmt.Println("No test environments found")
		return
	}

	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tStatus\tCreated\tTarget VM\tGit Server VM")
	fmt.Fprintln(w, "--\t--\t--\t--\t--")

	for _, env := range envs {
		createdStr := env.CreatedAt.Format("2006-01-02 15:04:05")
		targetVM := env.TargetVM.Name
		if targetVM == "" {
			targetVM = "(none)"
		}
		gitServerVM := env.GitServerVM.Name
		if gitServerVM == "" {
			gitServerVM = "(none)"
		}
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\n",
			env.ID,
			env.Status,
			createdStr,
			targetVM,
			gitServerVM,
		)
	}

	w.Flush()
}

// cmdTest runs a one-shot test (create → run → delete)
func cmdTest(ctx execcontext.Context, artifactStoreDir string) {
	fmt.Println("Running one-shot e2e test...")

	// Get paths
	cacheDir := filepath.Join(os.TempDir(), "edgectl")
	edgeCDRepoPath := getEdgeCDRepoPath()

	// Step 1: Create
	fmt.Println("\n[1/3] Creating test environment...")
	setupConfig := te2e.SetupConfig{
		ArtifactDir:    filepath.Join(artifactStoreDir, "artifacts"),
		ImageCacheDir:  cacheDir,
		EdgeCDRepoPath: edgeCDRepoPath,
		DownloadImages: true,
	}

	testEnv, err := te2e.SetupTestEnvironment(ctx, setupConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create test environment: %v\n", err)
		os.Exit(1)
	}

	// Save to artifact store
	if err := os.MkdirAll(artifactStoreDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create artifact store directory: %v\n", err)
		os.Exit(1)
	}

	store := te2e.NewJSONArtifactStore(filepath.Join(artifactStoreDir, "artifacts.json"))
	if err := store.Save(ctx, testEnv); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save environment to artifact store: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Test environment created: %s\n", testEnv.ID)

	// Cleanup at the end
	defer func() {
		fmt.Println("\n[3/3] Deleting test environment...")
		if err := te2e.TeardownTestEnvironmentWithLogging(ctx, testEnv); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: encountered errors during cleanup: %v\n", err)
		}
		if err := store.Delete(ctx, testEnv.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete environment from store: %v\n", err)
		}
	}()

	// Step 2: Run tests
	fmt.Println("\n[2/3] Running tests...")

	// Build edgectl binary
	binaryPath, err := te2e.BuildEdgectlBinary("./cmd/edgectl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to build edgectl binary: %v\n", err)
		os.Exit(1)
	}

	// Execute bootstrap test
	executorConfig := te2e.ExecutorConfig{
		EdgectlBinaryPath: binaryPath,
		ConfigPath:        "./test/edgectl/e2e/config",
		ConfigSpec:        "config.yaml",
		Packages:          "git,curl,openssh-client",
		ServiceManager:    "systemd",
		PackageManager:    "apt",
	}

	if err := te2e.ExecuteBootstrapTest(ctx, testEnv, executorConfig); err != nil {
		testEnv.Status = "failed"
		store.Save(ctx, testEnv)
		fmt.Fprintf(os.Stderr, "Error: bootstrap tests failed: %v\n", err)
		os.Exit(1)
	}

	testEnv.Status = "passed"
	store.Save(ctx, testEnv)

	fmt.Println("\n✅ One-shot e2e test completed successfully!")
}

// printEnvironmentJSON prints environment as JSON for parsing by other tools
func printEnvironmentJSON(env *te2e.TestEnvironment) {
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal environment: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// debugf prints debug messages to stderr if DEBUG is set
func debugf(format string, a ...interface{}) {
	if os.Getenv("EDGECTL_E2E_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format, a...)
	}
}

// isPiped returns true if stdout is piped to another process
func isPiped() bool {
	stat, _ := os.Stdout.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
