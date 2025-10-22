package e2e

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"libvirt.org/go/libvirt"
)

const (
	defaultMemoryMB = 2048
	defaultVCPUs    = 2
	defaultDiskSize = "20G"
	defaultNetwork  = "default"
)

type VMConfig struct {
	Name           string
	ImageQCOW2Path string
	DiskSize       string
	MemoryMB       uint
	VCPUs          uint
	Network        string
	UserData       string // Cloud-init user-data
	SSHKeyPath     string // Path to the SSH private key for connecting to the VM
}

func NewVMConfig(name, imagePath, sshKeyPath string) VMConfig {
	return VMConfig{
		Name:           name,
		ImageQCOW2Path: imagePath,
		DiskSize:       defaultDiskSize,
		MemoryMB:       defaultMemoryMB,
		VCPUs:          defaultVCPUs,
		Network:        defaultNetwork,
		SSHKeyPath:     sshKeyPath,
	}
}

func CreateVM(cfg VMConfig) (*libvirt.Connect, *libvirt.Domain, error) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	cloudInitISOPath, err := generateCloudInitISO(cfg.Name, cfg.UserData)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to generate cloud-init ISO: %w", err)
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
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create VM disk: %w\nOutput: %s", err, output)
	}
	defer os.Remove(vmDiskPath)

	vmXML := fmt.Sprintf(`
		<domain type='kvm'>
			<name>%s</name>
			<memory unit='MiB'>%d</memory>
			<vcpu>%d</vcpu>
			<os>
				<type arch='x86_64' machine='pc-q35-8.0'>hvm</type>
				<boot dev='hd'/>
			</os>
			<features>
				<acpi/>
				<apic/>
			</features>
			<cpu mode='host-passthrough' migs-feature='on'/>
			<clock offset='utc'/>
			<on_poweroff>destroy</on_poweroff>
			<on_reboot>restart</on_reboot>
			<on_crash>destroy</on_crash>
			<devices>
				<disk type='file' device='disk'>
					<driver name='qemu' type='qcow2'/>
					<source file='%s'/>
					<target dev='vda' bus='virtio'/>
				</disk>
				<disk type='file' device='cdrom'>
					<driver name='qemu' type='raw'/>
					<source file='%s'/>
					<target dev='sdb' bus='sata'/>
					<readonly/>
				</disk>
				<interface type='network'>
					<source network='%s'/>
					<model type='virtio'/>
				</interface>
				<console type='pty'>
					<target type='serial' port='0'/>
				</console>
				<channel type='unix'>
					<target type='virtio' name='org.qemu.guest_agent.0'/>
					<address type='virtio-serial' controller='0' bus='0' port='1'/>
				</channel>
				<rng model='virtio'>
					<backend model='random'>/dev/urandom</backend>
				</rng>
			</devices>
		</domain>
		`, cfg.Name, cfg.MemoryMB, cfg.VCPUs, vmDiskPath, cloudInitISOPath, cfg.Network)

	dom, err := conn.DomainDefineXML(vmXML)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to define domain: %w", err)
	}

	if err := dom.Create(); err != nil {
		dom.Free()
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create domain: %w", err)
	}

	return conn, dom, nil
}

func DestroyVM(conn *libvirt.Connect, dom *libvirt.Domain) error {
	if dom == nil {
		return nil // Already destroyed or not created
	}

	state, _, err := dom.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_RUNNING {
		if err := dom.Destroy(); err != nil {
			return fmt.Errorf("failed to destroy domain: %w", err)
		}
	}

	if err := dom.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine domain: %w", err)
	}

	dom.Free()
	conn.Close()
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
	if err := ioutil.WriteFile(userFile, []byte(userData), 0o644); err != nil {
		return "", fmt.Errorf("failed to write user-data file: %w", err)
	}

	metaFile := filepath.Join(cloudInitDir, "meta-data")
	if err := ioutil.WriteFile(metaFile, []byte(metaData), 0o644); err != nil {
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
		return "", fmt.Errorf("failed to create cloud-init ISO with xorriso: %w\nOutput: %s", err, output)
	}
	return isoPath, nil
}

func GetVMIPAddress(conn *libvirt.Connect, dom *libvirt.Domain) (string, error) {
	// Retry for up to 30 seconds to get the VM's IP address
	timeout := time.After(30 * time.Second)
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timed out waiting for VM IP address")
		case <-tick.C:
			ifaces, err := dom.ListAllInterfaceAddresses(
				libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE,
			)
			if err != nil {
				fmt.Printf("Error listing interface addresses: %v\n", err)
				continue
			}

			for _, iface := range ifaces {
				for _, addr := range iface.Addrs {
					if addr.Type == libvirt.IP_ADDR_TYPE_IPV4 {
						return strings.Split(addr.Addr, "/")[0], nil
					}
				}
			}
			fmt.Printf("VM IP address not found in libvirt interface addresses, retrying...\n")
		}
	}
}

func getConsoleOutput(conn *libvirt.Connect, dom *libvirt.Domain) (string, error) {
	domainName, err := dom.GetName()
	if err != nil {
		return "", fmt.Errorf("failed to get domain name: %w", err)
	}

	stream, err := conn.NewStream(0)
	if err != nil {
		return "", fmt.Errorf("failed to create new stream: %w", err)
	}
	defer stream.Free()

	// Open the console, passing the Stream object. Empty string for default console.
	// Flags can be 0 for default behavior.
	err = dom.OpenConsole("", stream, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open console for domain %s: %w", domainName, err)
	}

	var consoleOutput bytes.Buffer
	buffer := make([]byte, 4096)

	// Use a timeout for reading console output to prevent blocking indefinitely
	readTimeout := time.After(10 * time.Second) // Read for 10 seconds
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
					// Handle specific errors like EOF or stream closure
					if err == io.EOF {
						close(readDone)
						return
					}
					fmt.Printf("Error reading from console stream: %v\n", err)
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
