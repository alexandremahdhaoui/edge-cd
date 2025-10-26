package vmm

// VMMetadata holds information about a virtual machine
type VMMetadata struct {
	Name          string   // VM domain name in libvirt (e.g., "e2e-target-abc123")
	IP            string   // IP address assigned to VM (e.g., "192.168.1.100")
	DomainXML     string   // Complete libvirt domain XML (for recovery/debugging)
	SSHPort       int      // SSH port if non-standard
	MemoryMB      uint     // Memory allocated to VM
	VCPUs         uint     // Number of virtual CPUs
	CreatedFiles  []string // List of created files (disk, ISO, etc.) for audit and cleanup
}
