package vmm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/cloudinit"
	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

const (
	defaultMemoryMB = 2048
	defaultVCPUs    = 2
	defaultDiskSize = "20G"
	defaultNetwork  = "default"
)

// VMM manages libvirt virtual machines.
type VMM struct {
	conn    *libvirt.Connect
	domains map[string]*libvirt.Domain
	baseDir string // Optional base directory for VM temporary files
	// virtiofsds stores the virtiofsd processes started for each VM,
	// along with their cancellation functions.
	virtiofsds map[string][]struct {
		Cmd    *exec.Cmd
		Cancel context.CancelFunc
	}
}

// VMMOption is a function that modifies VMM configuration
type VMMOption func(*VMM)

// WithBaseDir returns an option that sets the base directory for VM temporary files
func WithBaseDir(baseDir string) VMMOption {
	return func(v *VMM) {
		v.baseDir = baseDir
	}
}

// NewVMM creates a new VMM instance and connects to libvirt.
// Optional options can be passed to configure the VMM.
func NewVMM(opts ...VMMOption) (*VMM, error) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	vmm := &VMM{
		conn:    conn,
		domains: make(map[string]*libvirt.Domain),
		baseDir: "",
		virtiofsds: make(map[string][]struct {
			Cmd    *exec.Cmd
			Cancel context.CancelFunc
		}),
	}

	// Apply options
	for _, opt := range opts {
		opt(vmm)
	}

	return vmm, nil
}

// Close closes the libvirt connection.
func (v *VMM) Close() error {
	if v.conn == nil {
		return nil
	}
	_, err := v.conn.Close()
	return err
}

type VMConfig struct {
	Name           string
	ImageQCOW2Path string
	DiskSize       string
	MemoryMB       uint
	VCPUs          uint
	Network        string
	UserData       cloudinit.UserData
	VirtioFS       []VirtioFSConfig // New field for virtiofs mounts
	TempDir        string           // Optional: directory for temporary VM files (disk overlay, cloud-init ISO). Defaults to os.TempDir() if empty
}

type VirtioFSConfig struct {
	Tag        string
	MountPoint string
}

func NewVMConfig(name, imagePath string, userData cloudinit.UserData) VMConfig {
	return VMConfig{
		Name:           name,
		ImageQCOW2Path: imagePath,
		DiskSize:       defaultDiskSize,
		MemoryMB:       defaultMemoryMB,
		VCPUs:          defaultVCPUs,
		Network:        defaultNetwork,
		UserData:       userData,
	}
}

// CreateVM creates and starts a new virtual machine.
// Returns metadata about the created VM including its IP address and domain XML.
func (v *VMM) CreateVM(cfg VMConfig) (*VMMetadata, error) {
	// Determine temp directory: cfg.TempDir > VMM.baseDir > os.TempDir()
	tempDir := cfg.TempDir
	if tempDir == "" && v.baseDir != "" {
		tempDir = v.baseDir
	}
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	userData, err := cfg.UserData.Render()
	if err != nil {
		return nil, err
	}

	cloudInitISOPath, err := generateCloudInitISO(cfg.Name, userData, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cloud-init ISO: %w", err)
	}
	defer os.Remove(cloudInitISOPath)

	// -- Create overlay vm image
	vmDiskPath := filepath.Join(tempDir, fmt.Sprintf("%s.qcow2", cfg.Name))
	qemuImgCmd := exec.Command(
		"qemu-img",
		"create",
		"-f",
		"qcow2",
		"-o",
		fmt.Sprintf("backing_file=%s,backing_fmt=qcow2", cfg.ImageQCOW2Path),
		vmDiskPath,
		cfg.DiskSize,
	)
	if output, err := qemuImgCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create VM disk: %w\nOutput: %s", err, output)
	}
	// Note: vmDiskPath is deleted in DestroyVM, not here
	// This allows the VM to keep using the disk while running

	var filesystems []libvirtxml.DomainFilesystem

	for _, fs := range cfg.VirtioFS {
		// libvirt will manage virtiofsd, so we don't start it manually here.
		// We only need to configure the DomainFilesystem.
		filesystems = append(filesystems, libvirtxml.DomainFilesystem{
			AccessMode: "passthrough", // As per user's documentation
			Driver: &libvirtxml.DomainFilesystemDriver{
				Type:  "virtiofs",
				Queue: 1024, // As per user's documentation
			},
			Target: &libvirtxml.DomainFilesystemTarget{
				Dir: fs.Tag, // This is the guest mount tag
			},
			Source: &libvirtxml.DomainFilesystemSource{
				Mount: &libvirtxml.DomainFilesystemSourceMount{
					Dir: fs.MountPoint, // This should be the host-side path
					// No Socket field here, libvirt will manage it implicitly
				},
			},
		})
	}
	// Remove virtiofsd processes map as libvirt will manage virtiofsd
	v.virtiofsds = make(map[string][]struct {
		Cmd    *exec.Cmd
		Cancel context.CancelFunc
	})

	domain := &libvirtxml.Domain{
		Type: "kvm",
		Name: cfg.Name,
		Memory: &libvirtxml.DomainMemory{
			Value: cfg.MemoryMB,
			Unit:  "MiB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: cfg.VCPUs,
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "pc-q35-8.0",
				Type:    "hvm",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{Dev: "hd"},
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "host-passthrough",
			// MigsFeature: "on", // This field is not directly available in libvirtxml.DomainCPU
		},
		Clock: &libvirtxml.DomainClock{
			Offset: "utc",
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "destroy",
		MemoryBacking: &libvirtxml.DomainMemoryBacking{
			MemorySource: &libvirtxml.DomainMemorySource{
				Type: "memfd",
			},
			MemoryAccess: &libvirtxml.DomainMemoryAccess{
				Mode: "shared",
			},
		},
		Devices: &libvirtxml.DomainDeviceList{
			Disks: []libvirtxml.DomainDisk{
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: vmDiskPath,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
				},
				{
					Device: "cdrom",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "raw",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: cloudInitISOPath,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "sdb",
						Bus: "sata",
					},
					ReadOnly: &libvirtxml.DomainDiskReadOnly{},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: cfg.Network,
						},
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
				},
			},
			Consoles: []libvirtxml.DomainConsole{
				{
					Target: &libvirtxml.DomainConsoleTarget{
						Type: "serial",
						Port: ptr(uint(0)),
					},
					Source: &libvirtxml.DomainChardevSource{
						Pty: &libvirtxml.DomainChardevSourcePty{},
					},
				},
			},
			Channels: []libvirtxml.DomainChannel{
				{
					Target: &libvirtxml.DomainChannelTarget{
						VirtIO: &libvirtxml.DomainChannelTargetVirtIO{
							Name: "org.qemu.guest_agent.0",
						},
					},
					Address: &libvirtxml.DomainAddress{
						VirtioSerial: &libvirtxml.DomainAddressVirtioSerial{
							Controller: ptr(uint(0)),
							Bus:        ptr(uint(0)),
							Port:       ptr(uint(1)),
						},
					},
				},
			},
			RNGs: []libvirtxml.DomainRNG{
				{
					Model: "virtio",
					Backend: &libvirtxml.DomainRNGBackend{
						Random: &libvirtxml.DomainRNGBackendRandom{
							Device: "/dev/urandom",
						},
					},
				},
			},
			Filesystems: filesystems, // Add filesystems here
		},
	}

	vmXML, err := domain.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal domain XML: %w", err)
	}

	dom, err := v.conn.DomainDefineXML(vmXML)
	if err != nil {
		return nil, fmt.Errorf("failed to define domain: %w", err)
	}

	if err := dom.Create(); err != nil {
		dom.Free()
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	v.domains[cfg.Name] = dom

	// Capture domain XML for recovery/debugging
	domXML, err := dom.GetXMLDesc(0)
	if err != nil {
		// Log error but don't fail - this is for debugging purposes
		fmt.Printf("Warning: failed to get domain XML for %s: %v\n", cfg.Name, err)
		domXML = ""
	}

	// Get the VM's IP address with retry logic
	ipAddress, err := v.GetDomainIP(context.Background(), cfg.Name, 60*time.Second)
	if err != nil {
		// Log but don't fail - IP might not be available immediately
		fmt.Printf("Warning: failed to get IP for VM %s: %v\n", cfg.Name, err)
		ipAddress = ""
	}

	// Track created files for audit and cleanup
	createdFiles := []string{vmDiskPath}
	if cloudInitISOPath != "" {
		createdFiles = append(createdFiles, cloudInitISOPath)
	}

	// Return metadata about the created VM
	return &VMMetadata{
		Name:         cfg.Name,
		IP:           ipAddress,
		DomainXML:    domXML,
		SSHPort:      22,
		MemoryMB:     cfg.MemoryMB,
		VCPUs:        cfg.VCPUs,
		CreatedFiles: createdFiles,
	}, nil
}

// DomainExists checks if a VM domain exists in libvirt.
// First checks the in-memory domains map for efficiency.
// If not found in memory, queries libvirt directly (critical for cleanup when using new VMM instances).
func (v *VMM) DomainExists(ctx context.Context, name string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// Check in-memory map first (optimization)
	dom, ok := v.domains[name]
	if ok && dom != nil {
		// Try to get the domain state to confirm it still exists
		_, _, err := dom.GetState()
		if err != nil {
			return false, nil // Domain doesn't exist
		}
		return true, nil
	}

	// Domain not in memory, query libvirt directly
	// This is critical for cleanup scenarios where a new VMM instance is created
	if v.conn == nil {
		return false, fmt.Errorf("libvirt connection is not initialized")
	}

	domain, err := v.conn.LookupDomainByName(name)
	if err != nil {
		// Domain not found in libvirt
		return false, nil
	}

	// Domain exists in libvirt, cache it in memory for future use
	if domain != nil {
		v.domains[name] = domain
		return true, nil
	}

	return false, nil
}

// GetDomainIP retrieves the IP address of a running VM
// Polls with backoff up to the specified timeout duration
func (v *VMM) GetDomainIP(ctx context.Context, name string, timeout time.Duration) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	dom, ok := v.domains[name]
	if !ok || dom == nil {
		return "", fmt.Errorf("VM %s not found or not running", name)
	}

	// Retry with exponential backoff up to timeout
	deadline := time.Now().Add(timeout)
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf("timed out waiting for VM %s IP address", name)
		}

		// Try to get IP address
		ifaces, err := dom.ListAllInterfaceAddresses(
			libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE,
		)
		if err == nil {
			for _, iface := range ifaces {
				for _, addr := range iface.Addrs {
					if addr.Type == libvirt.IP_ADDR_TYPE_IPV4 {
						return strings.Split(addr.Addr, "/")[0], nil
					}
				}
			}
		}

		// Wait before retrying with exponential backoff
		waitTime := backoff
		if backoff < maxBackoff {
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(waitTime):
			// Continue to next iteration
		}
	}
}

// GetDomainXML returns the full XML definition of a domain
func (v *VMM) GetDomainXML(ctx context.Context, name string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	dom, ok := v.domains[name]
	if !ok || dom == nil {
		return "", fmt.Errorf("VM %s not found", name)
	}

	xml, err := dom.GetXMLDesc(0)
	if err != nil {
		return "", fmt.Errorf("failed to get domain XML for %s: %w", name, err)
	}

	return xml, nil
}

// GetDomainByName gets a domain handle by name, checking memory first then querying libvirt
// This helper function supports cleanup scenarios where a new VMM instance is created
// Returns nil if domain does not exist (allows idempotent cleanup)
func (v *VMM) GetDomainByName(ctx context.Context, name string) (*libvirt.Domain, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check in-memory map first (optimization)
	if dom, ok := v.domains[name]; ok && dom != nil {
		return dom, nil
	}

	// Domain not in memory, query libvirt directly
	if v.conn == nil {
		return nil, fmt.Errorf("libvirt connection is not initialized")
	}

	domain, err := v.conn.LookupDomainByName(name)
	if err != nil {
		// Domain not found - return nil, not error (for idempotent cleanup)
		return nil, nil
	}

	// Cache domain in memory for future use
	if domain != nil {
		v.domains[name] = domain
	}

	return domain, nil
}

// DestroyVM destroys a virtual machine and deletes its storage unconditionally
// This stops the VM, undefines it in libvirt, and deletes its disk files
// Caller is responsible for deciding whether to call this
func (v *VMM) DestroyVM(ctx context.Context, vmName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get domain handle (checks memory first, then queries libvirt)
	// GetDomainByName returns nil if domain doesn't exist (for idempotent cleanup)
	dom, err := v.GetDomainByName(ctx, vmName)
	if err != nil {
		return err
	}

	// If domain doesn't exist, treat as success (idempotent cleanup)
	if dom == nil {
		fmt.Printf("VM %s not found in libvirt, skipping destroy\n", vmName)
		// Still try to delete disk files if they exist
		tempDir := v.baseDir
		if tempDir == "" {
			tempDir = os.TempDir()
		}
		vmDiskPath := filepath.Join(tempDir, fmt.Sprintf("%s.qcow2", vmName))
		cloudInitISOPath := filepath.Join(tempDir, fmt.Sprintf("%s-cloud-init.iso", vmName))
		os.Remove(vmDiskPath)
		os.Remove(cloudInitISOPath)
		return nil
	}

	state, _, err := dom.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state for %s: %w", vmName, err)
	}

	// Stop the VM if it's running
	if state == libvirt.DOMAIN_RUNNING {
		if err := dom.Destroy(); err != nil {
			return fmt.Errorf("failed to destroy domain %s: %w", vmName, err)
		}
	}

	// Undefine the domain from libvirt
	if err := dom.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine domain %s: %w", vmName, err)
	}

	// Determine temp directory: v.baseDir > os.TempDir()
	tempDir := v.baseDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Delete the VM's disk file
	vmDiskPath := filepath.Join(tempDir, fmt.Sprintf("%s.qcow2", vmName))
	if err := os.Remove(vmDiskPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete VM disk %s: %w", vmDiskPath, err)
	}

	// Delete the cloud-init ISO if it exists
	cloudInitISOPath := filepath.Join(tempDir, fmt.Sprintf("%s-cloud-init.iso", vmName))
	if err := os.Remove(cloudInitISOPath); err != nil && !os.IsNotExist(err) {
		// Log but don't fail - this is just cleanup
		fmt.Printf("Warning: failed to delete cloud-init ISO %s: %v\n", cloudInitISOPath, err)
	}

	dom.Free()
	delete(v.domains, vmName)
	return nil
}

func generateCloudInitISO(vmName, userData, tempDir string) (string, error) {
	metaData := fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n", vmName, vmName)

	isoPath := filepath.Join(tempDir, fmt.Sprintf("%s-cloud-init.iso", vmName))

	// Create a temporary directory for cloud-init config files
	cloudInitDir := filepath.Join(tempDir, fmt.Sprintf("%s-cloud-init-config", vmName))
	if err := os.MkdirAll(cloudInitDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cloud-init config directory: %w", err)
	}
	defer os.RemoveAll(cloudInitDir)

	userFile := filepath.Join(cloudInitDir, "user-data")
	if err := os.WriteFile(userFile, []byte(userData), 0o644); err != nil {
		return "", fmt.Errorf("failed to write user-data file: %w", err)
	}

	metaFile := filepath.Join(cloudInitDir, "meta-data")
	if err := os.WriteFile(metaFile, []byte(metaData), 0o644); err != nil {
		return "", fmt.Errorf("failed to write meta-data file: %w", err)
	}

	xorrisoCmd := exec.Command(
		"xorriso",
		"-as", "mkisofs",
		"-o", isoPath,
		"-V", "cidata",
		"-J", "-R",
		cloudInitDir,
	)
	if output, err := xorrisoCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf(
			"failed to create cloud-init ISO with xorriso: %w\nOutput: %s",
			err,
			output,
		)
	}
	return isoPath, nil
}

// GetVMIPAddress retrieves the IP address of a running VM.
func (v *VMM) GetVMIPAddress(vmName string) (string, error) {
	dom, ok := v.domains[vmName]
	if !ok || dom == nil {
		return "", fmt.Errorf("VM %s not found or not running", vmName)
	}

	// Retry for up to 60 seconds to get the VM's IP address
	timeout := time.After(60 * time.Second)
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timed out waiting for VM %s IP address", vmName)
		case <-tick.C:
			ifaces, err := dom.ListAllInterfaceAddresses(
				libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE,
			)
			if err != nil {
				fmt.Printf("Error listing interface addresses for %s: %v\n", vmName, err)
				continue
			}

			for _, iface := range ifaces {
				for _, addr := range iface.Addrs {
					if addr.Type == libvirt.IP_ADDR_TYPE_IPV4 {
						return strings.Split(addr.Addr, "/")[0], nil
					}
				}
			}
			fmt.Printf("VM %s IP address not found in libvirt interface addresses, retrying...\n", vmName)
		}
	}
}

// GetConsoleOutput retrieves the serial console output of a VM.
func (v *VMM) GetConsoleOutput(vmName string) (string, error) {
	dom, ok := v.domains[vmName]
	if !ok || dom == nil {
		return "", fmt.Errorf("VM %s not found or not running", vmName)
	}

	domainName, err := dom.GetName()
	if err != nil {
		return "", fmt.Errorf("failed to get domain name for %s: %w", vmName, err)
	}

	stream, err := v.conn.NewStream(0)
	if err != nil {
		return "", fmt.Errorf("failed to create new stream for %s: %w", vmName, err)
	}
	defer stream.Free()

	err = dom.OpenConsole("", stream, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open console for domain %s: %w", domainName, err)
	}

	var consoleOutput bytes.Buffer
	buffer := make([]byte, 4096)

	readTimeout := time.After(10 * time.Second)
	readDone := make(chan struct{})

	go func() {
		for {
			select {
			case <-readTimeout:
				close(readDone)
				return
			default:
				n, err := stream.Recv(buffer)
				if err != nil {
					if err == io.EOF {
						close(readDone)
						return
					}
					fmt.Printf("Error reading from console stream for %s: %v\n", vmName, err)
					close(readDone)
					return
				}
				if n > 0 {
					consoleOutput.Write(buffer[:n])
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	<-readDone
	return consoleOutput.String(), nil
}

// Helper function to get a pointer to a uint
func ptr[T any](v T) *T {
	return &v
}
