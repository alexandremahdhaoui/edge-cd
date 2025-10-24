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
	// virtiofsds stores the virtiofsd processes started for each VM,
	// along with their cancellation functions.
	virtiofsds map[string][]struct {
		Cmd    *exec.Cmd
		Cancel context.CancelFunc
	}
}

// NewVMM creates a new VMM instance and connects to libvirt.
func NewVMM() (*VMM, error) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	return &VMM{
		conn:    conn,
		domains: make(map[string]*libvirt.Domain),
		virtiofsds: make(map[string][]struct {
			Cmd    *exec.Cmd
			Cancel context.CancelFunc
		}),
	}, nil
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
func (v *VMM) CreateVM(cfg VMConfig) (*libvirt.Domain, error) {
	userData, err := cfg.UserData.Render()
	if err != nil {
		return nil, err
	}

	cloudInitISOPath, err := generateCloudInitISO(cfg.Name, userData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cloud-init ISO: %w", err)
	}
	defer os.Remove(cloudInitISOPath)

	// -- Create overlay vm image
	vmDiskPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", cfg.Name))
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
	defer os.Remove(vmDiskPath)

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
	return dom, nil
}

// DestroyVM destroys a virtual machine.
func (v *VMM) DestroyVM(vmName string) error {
	dom, ok := v.domains[vmName]
	if !ok || dom == nil {
		return nil // Already destroyed or not created
	}

	state, _, err := dom.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state for %s: %w", vmName, err)
	}

	if state == libvirt.DOMAIN_RUNNING {
		if err := dom.Destroy(); err != nil {
			return fmt.Errorf("failed to destroy domain %s: %w", vmName, err)
		}
	}

	if err := dom.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine domain %s: %w", vmName, err)
	}

	dom.Free()
	delete(v.domains, vmName)
	return nil
}

func generateCloudInitISO(vmName, userData string) (string, error) {
	metaData := fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n", vmName, vmName)

	isoPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-cloud-init.iso", vmName))

	// Create a temporary directory for cloud-init config files
	cloudInitDir := filepath.Join(os.TempDir(), fmt.Sprintf("%s-cloud-init-config", vmName))
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
