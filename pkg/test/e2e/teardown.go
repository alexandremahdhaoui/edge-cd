package e2e

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/vmm"
)

// TeardownTestEnvironment destroys a test environment and cleans up all associated resources.
// This is the single source of truth for test cleanup and is used by both the test harness and CLI.
//
// This function is best-effort - it attempts to clean up all resources even if some operations fail.
// Returns a combined error if any cleanup operations failed, but other cleanup continues.
func TeardownTestEnvironment(ctx execcontext.Context, env *TestEnvironment) error {

	if env == nil || env.ID == "" {
		return fmt.Errorf("invalid test environment: nil or empty ID")
	}

	var combinedErr error

	// Destroy target VM
	if env.TargetVM.Name != "" {
		if err := destroyVMByName(ctx, env.TargetVM.Name); err != nil {
			combinedErr = errors.Join(combinedErr, fmt.Errorf("failed to destroy target VM: %w", err))
		}
	}

	// Destroy git server VM
	if env.GitServerVM.Name != "" {
		if err := destroyVMByName(ctx, env.GitServerVM.Name); err != nil {
			combinedErr = errors.Join(combinedErr, fmt.Errorf("failed to destroy git server VM: %w", err))
		}
	}

	// Clean up entire temp directory root (contains all component subdirs)
	// Only remove if it's a managed temp directory (has marker file) for safety
	if env.TempDirRoot != "" {
		if IsManagedTempDirectory(env.TempDirRoot) {
			if err := os.RemoveAll(env.TempDirRoot); err != nil {
				combinedErr = errors.Join(combinedErr, fmt.Errorf("failed to remove temp directory root: %w", err))
			}
		} else {
			combinedErr = errors.Join(combinedErr, fmt.Errorf("temp directory root is not marked as managed, skipping deletion: %s", env.TempDirRoot))
		}
	}

	// Clean up artifacts directory (backward compat, separate from TempDirRoot)
	if env.ArtifactPath != "" {
		if err := os.RemoveAll(env.ArtifactPath); err != nil {
			combinedErr = errors.Join(combinedErr, fmt.Errorf("failed to remove artifact directory: %w", err))
		}
	}

	return combinedErr
}

// destroyVMByName destroys a VM by name via libvirt.
// It handles both running and stopped VMs.
// If the VM doesn't exist, it returns nil (not an error) since the goal is cleanup.
func destroyVMByName(ctx execcontext.Context, vmName string) error {
	// Connect to libvirt
	vmManager, err := vmm.NewVMM()
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer vmManager.Close()

	// Check if VM exists
	exists, err := vmManager.DomainExists(ctx, vmName)
	if err != nil {
		return fmt.Errorf("failed to check if VM exists: %w", err)
	}

	// If VM doesn't exist, that's OK for cleanup purposes
	if !exists {
		return nil
	}

	// Destroy the VM (this handles both running and stopped states)
	if err := vmManager.DestroyVM(ctx, vmName); err != nil {
		return fmt.Errorf("failed to destroy VM %s: %w", vmName, err)
	}

	return nil
}

// TeardownTestEnvironmentWithLogging is like TeardownTestEnvironment but logs all cleanup operations.
// Useful for CLI tools that want to show progress to the user.
func TeardownTestEnvironmentWithLogging(ctx execcontext.Context, env *TestEnvironment) error {

	if env == nil || env.ID == "" {
		return fmt.Errorf("invalid test environment: nil or empty ID")
	}

	var combinedErr error

	// Destroy target VM
	if env.TargetVM.Name != "" {
		fmt.Printf("Destroying target VM: %s\n", env.TargetVM.Name)
		if err := destroyVMByName(ctx, env.TargetVM.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to destroy target VM: %v\n", err)
			combinedErr = errors.Join(combinedErr, err)
		} else {
			fmt.Println("  ✓ Target VM destroyed")
		}
	}

	// Destroy git server VM
	if env.GitServerVM.Name != "" {
		fmt.Printf("Destroying git server VM: %s\n", env.GitServerVM.Name)
		if err := destroyVMByName(ctx, env.GitServerVM.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to destroy git server VM: %v\n", err)
			combinedErr = errors.Join(combinedErr, err)
		} else {
			fmt.Println("  ✓ Git server VM destroyed")
		}
	}

	// Clean up entire temp directory root (contains all component subdirs)
	// Only remove if it's a managed temp directory (has marker file) for safety
	if env.TempDirRoot != "" {
		fmt.Printf("Removing temp directory: %s\n", env.TempDirRoot)
		if !IsManagedTempDirectory(env.TempDirRoot) {
			err := fmt.Errorf("temp directory root is not marked as managed, skipping deletion: %s", env.TempDirRoot)
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			combinedErr = errors.Join(combinedErr, err)
		} else if err := os.RemoveAll(env.TempDirRoot); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory: %v\n", err)
			combinedErr = errors.Join(combinedErr, err)
		} else {
			fmt.Println("  ✓ Temp directory removed")
		}
	}

	// Clean up artifacts directory (backward compat, separate from TempDirRoot)
	if env.ArtifactPath != "" {
		fmt.Printf("Removing artifacts from: %s\n", env.ArtifactPath)
		if err := os.RemoveAll(env.ArtifactPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove artifact directory: %v\n", err)
			combinedErr = errors.Join(combinedErr, err)
		} else {
			fmt.Println("  ✓ Artifacts removed")
		}
	}

	if combinedErr != nil {
		slog.Error(
			"encountered errors while tearing down test environment",
			"environment_id", env.ID,
			"error", combinedErr.Error(),
		)
	}

	return combinedErr
}
