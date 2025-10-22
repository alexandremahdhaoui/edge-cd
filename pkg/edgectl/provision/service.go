package provision

import (
	"fmt"
	"path/filepath"

	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

// SetupEdgeCDService sets up and enables the edge-cd service on the remote device.
func SetupEdgeCDService(runner ssh.Runner, serviceManagerName, edgeCDRepoPath string) error {
	// Determine source and destination paths for the service file
	serviceSourcePath := filepath.Join(
		edgeCDRepoPath,
		"cmd",
		"edge-cd",
		"service-managers",
		serviceManagerName,
		"service",
	)
	var serviceDestPath string
	var enableCmd, startCmd string
	var stdout, stderr string // Declare once at the beginning
	var err error             // Declare once at the beginning

	switch serviceManagerName {
	case "systemd":
		serviceDestPath = "/etc/systemd/system/edge-cd.service"

		// Check if the service is masked by a symlink to /dev/null and remove it
		checkMaskedSymlinkCmd := fmt.Sprintf(
			"test -L %s && readlink %s | grep -q '/dev/null'",
			serviceDestPath,
			serviceDestPath,
		)
		_, _, err = runner.Run(checkMaskedSymlinkCmd)
		if err == nil { // If the symlink to /dev/null exists
			rmMaskedSymlinkCmd := fmt.Sprintf("rm %s", serviceDestPath)
			fmt.Printf("Removing masked symlink %s...\n", serviceDestPath)
			stdout, stderr, err = runner.Run(rmMaskedSymlinkCmd) // Assign using =
			if err != nil {
				return fmt.Errorf(
					"failed to remove masked symlink: %w. Stdout: %s, Stderr: %s",
					err,
					stdout,
					stderr,
				)
			}
		}

		// Reload daemon after potential symlink removal

		daemonReloadCmd := "systemctl daemon-reload"
		fmt.Printf("Reloading systemd daemon...\n")
		stdout, stderr, err = runner.Run(daemonReloadCmd)
		if err != nil {
			return fmt.Errorf(
				"failed to reload systemd daemon: %w. Stdout: %s, Stderr: %s",
				err,
				stdout,
				stderr,
			)
		}

		// Unmask the service (in case it was masked by other means)
		unmaskCmd := "systemctl unmask edge-cd.service"
		fmt.Printf("Unmasking service %s...\n", serviceManagerName)
		stdout, stderr, err = runner.Run(unmaskCmd)
		if err != nil {
			return fmt.Errorf(
				"failed to unmask service: %w. Stdout: %s, Stderr: %s",
				err,
				stdout,
				stderr,
			)
		}

		enableCmd = "systemctl enable edge-cd.service"
		startCmd = "systemctl start edge-cd.service"
	case "procd":
		serviceDestPath = "/etc/init.d/edge-cd"
		enableCmd = "/etc/init.d/edge-cd enable"
		startCmd = "/etc/init.d/edge-cd start"
	default:
		return fmt.Errorf("unsupported service manager: %s", serviceManagerName)
	}

	// Copy service file to destination
	copyCmd := fmt.Sprintf("cp %s %s", serviceSourcePath, serviceDestPath)
	fmt.Printf("Copying service file from %s to %s...\n", serviceSourcePath, serviceDestPath)
	stdout, stderr, err = runner.Run(copyCmd) // Assign using =
	if err != nil {
		return fmt.Errorf(
			"failed to copy service file: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	fmt.Printf("Enabling service %s...\n", serviceManagerName)
	stdout, stderr, err = runner.Run(enableCmd) // Assign using =
	if err != nil {
		return fmt.Errorf(
			"failed to enable service: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	fmt.Printf("Starting service %s...\n", serviceManagerName)
	stdout, stderr, err = runner.Run(startCmd) // Assign using =
	if err != nil {
		return fmt.Errorf(
			"failed to start service: %w. Stdout: %s, Stderr: %s",
			err,
			stdout,
			stderr,
		)
	}

	return nil
}
