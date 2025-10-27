package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// PrepareLibvirtDir creates a subdirectory within the given parent directory
// with permissions that allow libvirt to access VM disk files.
// This is necessary because t.TempDir() creates directories with 0700 permissions
// which prevent libvirt (running as qemu/libvirt-qemu user) from accessing files.
//
// The function:
// 1. Creates a subdirectory with 0755 permissions
// 2. Changes parent directory to 0755 permissions (so libvirt can traverse it)
// 3. Attempts to set ownership to the libvirt user if needed
// 4. Returns the path to the libvirt-accessible directory
func PrepareLibvirtDir(t *testing.T, parentDir, subdirName string) string {
	// Create subdirectory path
	libvirtDir := filepath.Join(parentDir, subdirName)

	// Create directory with 0755 permissions (readable/executable by all)
	if err := os.MkdirAll(libvirtDir, 0755); err != nil {
		t.Fatalf("failed to create libvirt directory %q: %v", libvirtDir, err)
	}

	// Try to make parent and ancestor directories executable so libvirt can traverse them
	// This is critical: even if subdirectory has 0755, libvirt needs +x on all ancestors
	// We chmod all directories from parentDir up to /tmp
	currentDir := parentDir
	for {
		if err := os.Chmod(currentDir, 0755); err != nil {
			t.Logf("Warning: failed to chmod directory %q: %v", currentDir, err)
		}

		// Stop at /tmp or filesystem root
		if currentDir == "/tmp" || currentDir == "/" {
			break
		}

		// Move to parent directory
		nextDir := filepath.Dir(currentDir)
		if nextDir == currentDir {
			// Reached filesystem root
			break
		}
		currentDir = nextDir
	}

	// Grant libvirt-related groups access to the directory
	// This allows libvirt (running as libvirt-qemu or qemu user) to access VM disk files
	libvirtGroups := detectLibvirtGroups(t)
	if len(libvirtGroups) > 0 {
		// Set ACL for all detected libvirt-related groups
		for _, group := range libvirtGroups {
			setfaclCmd := exec.Command("sudo", "setfacl", "-m", fmt.Sprintf("g:%s:rwx", group), libvirtDir)
			if output, err := setfaclCmd.CombinedOutput(); err != nil {
				t.Logf("Warning: failed to set ACL for %q to allow group %q: %v\nOutput: %s",
					libvirtDir, group, err, output)
			} else {
				// Also set default ACL so files/dirs created in this directory inherit the permission
				defaultAclCmd := exec.Command("sudo", "setfacl", "-d", "-m", fmt.Sprintf("g:%s:rwx", group), libvirtDir)
				if output, err := defaultAclCmd.CombinedOutput(); err != nil {
					t.Logf("Warning: failed to set default ACL for group %q: %v\nOutput: %s",
						group, err, output)
				}
				t.Logf("Successfully set ACL on %q to allow group %q access", libvirtDir, group)
			}
		}
	}

	return libvirtDir
}

// detectLibvirtGroups attempts to detect all groups that libvirt might use
// by checking system configuration and common patterns
func detectLibvirtGroups(t *testing.T) []string {
	groups := make(map[string]bool)

	// Try to get the group from /etc/libvirt/qemu.conf
	if data, err := os.ReadFile("/etc/libvirt/qemu.conf"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "group = ") {
				group := strings.Trim(strings.TrimPrefix(line, "group = "), "\"")
				if group != "" {
					groups[group] = true
					t.Logf("Detected libvirt group from config: %s", group)
				}
			}
		}
	}

	// Try all common group names - we grant access to all that exist
	commonGroups := []string{"libvirt", "libvirt-qemu", "kvm", "qemu"}
	for _, group := range commonGroups {
		cmd := exec.Command("getent", "group", group)
		if err := cmd.Run(); err == nil {
			if !groups[group] {
				groups[group] = true
				t.Logf("Detected libvirt-related group: %s", group)
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(groups))
	for group := range groups {
		result = append(result, group)
	}

	if len(result) == 0 {
		t.Logf("Could not detect any libvirt groups, relying on permissions only")
	}

	return result
}
