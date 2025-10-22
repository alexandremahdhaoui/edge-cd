package e2e

import (
	"fmt"
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

	vmDiskPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", cfg.Name))
	qemuImgCmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-b", cfg.ImageQCOW2Path, vmDiskPath, cfg.DiskSize)
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

	userFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-user-data", vmName))
	if err := ioutil.WriteFile(userFile, []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data file: %w", err)
	}
	defer os.Remove(userFile)

	metaFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-meta-data", vmName))
	if err := ioutil.WriteFile(metaFile, []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data file: %w", err)
	}
	defer os.Remove(metaFile)

	isoPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-cloud-init.iso", vmName))
	genisoimageCmd := exec.Command("genisoimage", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", userFile, metaFile)
	if output, err := genisoimageCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create cloud-init ISO: %w\nOutput: %s", err, output)
	}

	return isoPath, nil
}

func GetVMIPAddress(conn *libvirt.Connect, dom *libvirt.Domain) (string, error) {
	time.Sleep(30 * time.Second)

	net, err := conn.LookupNetworkByUUIDString("default") // Assuming 'default' network
	if err != nil {
		return "", fmt.Errorf("failed to lookup default network: %w", err)
	}
	defer net.Free()

	domName, err := dom.GetName()
	if err != nil {
		return "", fmt.Errorf("failed to get domain name: %w", err)
	}

	virshCmd := exec.Command("virsh", "net-dhcp-leases", "default")
	output, err := virshCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get DHCP leases: %w\nOutput: %s", err, output)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, domName) && strings.Contains(line, "ipv4") {
			fields := strings.Fields(line)
			if len(fields) > 4 {
				ipField := fields[3] // e.g., 192.168.122.10/24
				return strings.Split(ipField, "/")[0], nil
			}
		}
	}

	return "", fmt.Errorf("VM IP address not found")
}
