package gitserver

import "github.com/alexandremahdhaoui/edge-cd/pkg/vmm"

// Status represents the complete state of a git server.
// This is the ONLY public way to query git server information.
// Status is read-only and provides all necessary details for callers
// to interact with and monitor the git server.
type Status struct {
	// VMMetadata contains information about the VM hosting the git server
	VMMetadata *vmm.VMMetadata

	// BaseDir is the root directory containing all server artifacts
	BaseDir string

	// GitSSHURLs maps repository names to their SSH clone URLs
	// Example: GitSSHURLs["edge-cd"] = "ssh://git@192.168.1.1:22/srv/git/edge-cd.git"
	GitSSHURLs map[string]string

	// ServicePort is the SSH port on which the git server listens
	// Typically 22 for standard SSH
	ServicePort int
}

// Copy returns a deep copy of the Status struct.
// This prevents external callers from accidentally modifying the internal state.
func (s *Status) Copy() *Status {
	if s == nil {
		return nil
	}

	copy := *s

	// Deep copy the GitSSHURLs map
	if s.GitSSHURLs != nil {
		copy.GitSSHURLs = make(map[string]string)
		for k, v := range s.GitSSHURLs {
			copy.GitSSHURLs[k] = v
		}
	}

	// Note: VMMetadata is shallow-copied, which is fine since it should be treated as immutable
	return &copy
}
