# libvirtxml.Domain

The `Domain` struct in `libvirt.org/go/libvirtxml` represents the top-level configuration for a virtual machine domain. It is used to define, create, and manage virtual machines using the libvirt API. This document details the `Domain` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Domain XML format](https://libvirt.org/formatdomain.html).

```go
// Domain represents the top-level configuration for a virtual machine domain.
type Domain struct {
	XMLName         xml.Name // XMLName is the XML element name, typically "domain".
	Type            string   // Corresponds to the `type` attribute of the `<domain>` root element. Specifies the hypervisor used for running the domain (e.g., "kvm", "qemu", "lxc").
	ID              *int     // Corresponds to the `id` attribute of the `<domain>` root element. A unique integer identifier for the running guest machine. Inactive machines have no `id` value.
	Name            string   // Corresponds to the `<name>` element. Provides a short name for the virtual machine. Must be unique within the host.
	UUID            string   // Corresponds to the `<uuid>` element. Provides a globally unique identifier for the virtual machine (RFC 4122 compliant). If omitted, a random UUID is generated.
	HWUUID          string   // Corresponds to the `<hwuuid>` element. An optional alternative UUID for identifying the virtual machine, affecting all devices that expose the UUID to the guest. (QEMU/KVM only, since 11.7.0)
	GenID           *DomainGenID // Corresponds to the `<genid>` element. Used to add a Virtual Machine Generation ID (GUID) to notify the guest OS of re-execution events.
	Title           string   // Corresponds to the `<title>` element. A short description of the domain.
	Description     string   // Corresponds to the `<description>` element. A human-readable description of the virtual machine.
	Metadata        *DomainMetadata // Corresponds to the `<metadata>` element. Allows applications to store custom XML metadata.
	MaximumMemory   *DomainMaxMemory // Corresponds to the `<maxMemory>` element. The run-time maximum memory allocation for the guest, allowing for hot-plugging memory up to this limit.
	Memory          *DomainMemory // Corresponds to the `<memory>` element. The initial memory allocation for the guest at boot time.
	CurrentMemory   *DomainCurrentMemory // Corresponds to the `<currentMemory>` element. The actual memory allocated for the guest, which can be less than `MaximumMemory` to allow for ballooning.
	BlockIOTune     *DomainBlockIOTune // Corresponds to the `<blkiotune>` element. Provides details regarding Block I/O tunable parameters for the domain.
	MemoryTune      *DomainMemoryTune // Corresponds to the `<memtune>` element. Provides details regarding memory tunable parameters for the domain.
	MemoryBacking   *DomainMemoryBacking // Corresponds to the `<memoryBacking>` element. Influences how virtual memory pages are backed by host pages.
	VCPU            *DomainVCPU // Corresponds to the `<vcpu>` element. Defines the maximum number of virtual CPUs allocated for the guest OS.
	VCPUs           *DomainVCPUs // Corresponds to the `<vcpus>` element. Allows control over the state of individual vCPUs.
	IOThreads       uint // Corresponds to the `<iothreads>` element. Defines the number of IOThreads to be assigned to the domain for supported disk devices.
	IOThreadIDs     *DomainIOThreadIDs // Corresponds to the `<iothreadids>` element. Provides the capability to specifically define the IOThread IDs for the domain.
	DefaultIOThread *DomainDefaultIOThread // Corresponds to the `<defaultiothread>` element. Represents the default event loop within the hypervisor for I/O requests not assigned to a specific IOThread.
	CPUTune         *DomainCPUTune // Corresponds to the `<cputune>` element. Provides details regarding CPU tunable parameters for the domain.
	NUMATune        *DomainNUMATune // Corresponds to the `<numatune>` element. Provides details on how to tune the performance of a NUMA host.
	Resource        *DomainResource // Corresponds to the `<resource>` element. Groups configuration related to resource partitioning.
	SysInfo         []DomainSysInfo // Corresponds to the `<sysinfo>` element. Allows control over system information presented to the guest.
	Bootloader      string // Corresponds to the `<bootloader>` element. Provides a fully qualified path to the bootloader executable in the host OS for paravirtualized guests.
	BootloaderArgs  string // Corresponds to the `<bootloader_args>` element. Allows command line arguments to be passed to the bootloader.
	OS              *DomainOS // Corresponds to the `<os>` element. Specifies the operating system booting configuration.
	IDMap           *DomainIDMap // Corresponds to the `<idmap>` element. Used to enable user namespace mapping for container-based virtualization.
	ThrottleGroups  *DomainThrottleGroups // Corresponds to the `<throttlegroups>` element. Allows creating multiple named throttle groups for disk I/O.
	Features        *DomainFeatureList // Corresponds to the `<features>` element. Allows toggling various hypervisor features.
	CPU             *DomainCPU // Corresponds to the `<cpu>` element. Main container for describing guest CPU requirements.
	Clock           *DomainClock // Corresponds to the `<clock>` element. Controls how the guest clock is synchronized to the host clock and timer settings.
	OnPoweroff      string // Corresponds to the `<on_poweroff>` element. Specifies the action to take when the guest requests a poweroff.
	OnReboot        string // Corresponds to the `<on_reboot>` element. Specifies the action to take when the guest requests a reboot.
	OnCrash         string // Corresponds to the `<on_crash>` element. Specifies the action to take when the guest crashes.
	PM              *DomainPM // Corresponds to the `<pm>` element. Forcibly enables or disables BIOS advertisements to the guest OS for power management.
	Perf            *DomainPerf // Corresponds to the `<perf>` element. Allows monitoring of performance events for the virtual machine.
	Devices         *DomainDeviceList // Corresponds to the `<devices>` element. Main container for describing all devices provided to the guest domain.
	SecLabel        []DomainSecLabel // Corresponds to the `<seclabel>` element. Allows control over the operation of security drivers for the domain.
	KeyWrap         *DomainKeyWrap // Corresponds to the `<keywrap>` element. Specifies whether the guest will be allowed to perform S390 cryptographic key management operations.
	LaunchSecurity  *DomainLaunchSecurity // Corresponds to the `<launchSecurity>` element. Provides guest owner input for creating encrypted VMs.
	QEMUCommandline      *DomainQEMUCommandline      // Corresponds to the `qemu:commandline` element. Allows passing arbitrary command-line arguments and environment variables to the QEMU process.
	QEMUCapabilities     *DomainQEMUCapabilities     // Corresponds to the `qemu:capabilities` element. Allows adding or deleting specific QEMU capabilities.
	QEMUOverride         *DomainQEMUOverride         // Corresponds to the `qemu:override` element. Allows overriding properties of QEMU devices.
	QEMUDeprecation      *DomainQEMUDeprecation      // Corresponds to the `qemu:deprecation` element. Configures QEMU deprecation behavior.
	LXCNamespace         *DomainLXCNamespace         // Corresponds to the `lxc:namespace` element. Configures LXC namespace sharing.
	BHyveCommandline     *DomainBHyveCommandline     // Corresponds to the `bhyve:commandline` element. Allows passing arbitrary command-line arguments and environment variables to the BHyve process.
	VMWareDataCenterPath *DomainVMWareDataCenterPath // Corresponds to the `vmware:datacenterpath` element. Specifies the VMware datacenter path.
	XenCommandline       *DomainXenCommandline       // Corresponds to the `xen:commandline` element. Allows passing arbitrary command-line arguments to the Xen process.
}
```

## Nested Structs

```go
// DomainGenID represents the `<genid>` element, used to add a Virtual Machine Generation ID (GUID).
type DomainGenID struct {
	Value string  // The 128-bit cryptographically random integer value (GUID).
}

// DomainMetadata represents the `<metadata>` element, allowing applications to store custom XML metadata.
type DomainMetadata struct {
	XML string  // Contains arbitrary XML content for custom metadata, using custom namespaces.
}

// DomainMaxMemory represents the `<maxMemory>` element, defining the run-time maximum memory allocation for the guest.
type DomainMaxMemory struct {
	Value uint             // The maximum memory allocation.
	Unit  string  (e.g., "KiB", "MiB", "GiB").
	Slots uint    // The number of slots available for adding memory.
}

// DomainMemory represents the `<memory>` element, defining the initial memory allocation for the guest at boot time.
type DomainMemory struct {
	Value    uint             // The initial memory allocation.
	Unit     string .
	DumpCore string  // Controls whether guest memory should be included in the coredump ("on", "off").
}

// DomainCurrentMemory represents the `<currentMemory>` element, defining the actual memory allocated for the guest.
type DomainCurrentMemory struct {
	Value uint             // The actual memory allocated.
	Unit  string .
}

// DomainBlockIOTune represents the `<blkiotune>` element, providing details regarding Block I/O tunable parameters for the domain.
type DomainBlockIOTune struct {
	Weight uint                       // Overall I/O weight of the guest.
	Device []DomainBlockIOTuneDevice            // List of devices with specific I/O tuning.
}

// DomainBlockIOTuneDevice represents the `<device>` sub-element of `<blkiotune>`, specifying I/O tuning for a specific block device.
type DomainBlockIOTuneDevice struct {
	Path          string                     // Absolute path of the host block device.
	Weight        uint           // Relative weight of the device.
	ReadIopsSec   uint    // Read I/O operations per second limit.
	WriteIopsSec  uint    // Write I/O operations per second limit.
	ReadBytesSec  uint    // Read throughput limit in bytes per second.
	WriteBytesSec uint    // Write throughput limit in bytes per second.
}

// DomainMemoryTune represents the `<memtune>` element, providing details regarding memory tunable parameters for the domain.
type DomainMemoryTune struct {
	HardLimit     *DomainMemoryTuneLimit      // Maximum memory the guest can use.
	SoftLimit     *DomainMemoryTuneLimit      // Memory limit to enforce during memory contention.
	MinGuarantee  *DomainMemoryTuneLimit   // Guaranteed minimum memory allocation.
	SwapHardLimit *DomainMemoryTuneLimit  // Maximum memory plus swap the guest can use.
}

// DomainMemoryTuneLimit represents memory limit values within `<memtune>`.
type DomainMemoryTuneLimit struct {
	Value uint64           // The memory limit value.
	Unit  string .
}

// DomainMemoryBacking represents the `<memoryBacking>` element, influencing how virtual memory pages are backed by host pages.
type DomainMemoryBacking struct {
	MemoryHugePages    *DomainMemoryHugepages        // Configuration for hugepages.
	MemoryNosharepages *DomainMemoryNosharepages  // Disables shared pages (KSM).
	MemoryLocked       *DomainMemoryLocked              // Locks memory pages in host's memory.
	MemorySource       *DomainMemorySource              // Specifies the source of memory backing.
	MemoryAccess       *DomainMemoryAccess              // Specifies if memory is "shared" or "private".
	MemoryAllocation   *DomainMemoryAllocation      // Specifies when to allocate memory.
	MemoryDiscard      *DomainMemoryDiscard            // Discards memory content on guest shutdown.
}

// DomainMemoryHugepages represents the `<hugepages>` element, containing a list of hugepage configurations.
type DomainMemoryHugepages struct {
	Hugepages []DomainMemoryHugepage  // List of hugepage configurations.
}

// DomainMemoryHugepage represents a `<page>` sub-element of `<hugepages>`, specifying hugepage size and NUMA nodes.
type DomainMemoryHugepage struct {
	Size    uint             // Size of the hugepage.
	Unit    string .
	Nodeset string  // NUMA nodes to tie hugepages to.
}

// DomainMemoryNosharepages represents the `<nosharepages>` element, disabling shared pages (KSM).
type DomainMemoryNosharepages struct {
}

// DomainMemoryLocked represents the `<locked>` element, locking memory pages in host's memory.
type DomainMemoryLocked struct {
}

// DomainMemorySource represents the `<source>` sub-element of `<memoryBacking>`, specifying the source of memory backing.
type DomainMemorySource struct {
	Type string  // Type of memory backing ("file", "anonymous", "memfd").
}

// DomainMemoryAccess represents the `<access>` sub-element of `<memoryBacking>`, specifying memory access mode.
type DomainMemoryAccess struct {
	Mode string  // Memory access mode ("shared", "private").
}
// DomainMemoryAllocation represents the `<allocation>` sub-element of `<memoryBacking>`, specifying memory allocation timing.
type DomainMemoryAllocation struct {
	Mode    string  // Allocation mode ("immediate", "ondemand").
	Threads uint    // Number of threads for memory allocation.
}

// DomainMemoryDiscard represents the `<discard>` element, discarding memory content on guest shutdown.
type DomainMemoryDiscard struct {
}

// DomainVCPU represents the `<vcpu>` element, defining the maximum number of virtual CPUs allocated for the guest OS.
type DomainVCPU struct {
	Placement string  // CPU placement mode ("static", "auto").
	CPUSet    string     // Comma-separated list of physical CPU numbers.
	Current   uint      // Number of currently enabled vCPUs.
	Value     uint                   // Maximum number of virtual CPUs.
}

// DomainVCPUs represents the `<vcpus>` element, allowing control over the state of individual vCPUs.
type DomainVCPUs struct {
	VCPU []DomainVCPUsVCPU  // List of individual vCPU configurations.
}

// DomainVCPUsVCPU represents a `<vcpu>` sub-element of `<vcpus>`, configuring an individual vCPU.
type DomainVCPUsVCPU struct {
	Id           *uint                       // vCPU identifier.
	Enabled      string       // Whether the vCPU is enabled ("yes", "no").
	Hotpluggable string  // Whether the vCPU can be hotplugged ("yes", "no").
	Order        *uint                    // Order to add online vCPUs.
}

// DomainIOThreadIDs represents the `<iothreadids>` element, defining specific IOThread IDs for the domain.
type DomainIOThreadIDs struct {
	IOThreads []DomainIOThread  // List of individual IOThread configurations.
}

// DomainIOThread represents an `<iothread>` sub-element of `<iothreadids>`, configuring an individual IOThread.
type DomainIOThread struct {
	ID      uint                                  // IOThread identifier.
	PoolMin *uint                    // Lower boundary for worker threads.
	PoolMax *uint                    // Upper boundary for worker threads.
	Poll    *DomainIOThreadPoll                      // Polling settings for the IOThread.
}

// DomainIOThreadPoll represents the `<poll>` sub-element of `<iothread>`, configuring polling settings.
type DomainIOThreadPoll struct {
	Max    *uint     // Maximum polling time in nanoseconds.
	Grow   *uint    // Steps for increasing polling interval.
	Shrink *uint  // Steps for decreasing polling interval.
}

// DomainDefaultIOThread represents the `<defaultiothread>` element, configuring the default event loop for I/O requests.
type DomainDefaultIOThread struct {
	PoolMin *uint  // Lower boundary for worker threads of the default event loop.
	PoolMax *uint  // Upper boundary for worker threads of the default event loop.
}

// DomainCPUTune represents the `<cputune>` element, providing details regarding CPU tunable parameters for the domain.
type DomainCPUTune struct {
	Shares         *DomainCPUTuneShares                   // Proportional weighted share for the domain.
	Period         *DomainCPUTunePeriod                   // Enforcement interval for vCPUs.
	Quota          *DomainCPUTuneQuota                     // Maximum allowed bandwidth for vCPUs.
	GlobalPeriod   *DomainCPUTunePeriod            // Enforcement interval for the whole domain.
	GlobalQuota    *DomainCPUTuneQuota              // Maximum allowed bandwidth for the whole domain.
	EmulatorPeriod *DomainCPUTunePeriod          // Enforcement interval for emulator threads.
	EmulatorQuota  *DomainCPUTuneQuota            // Maximum allowed bandwidth for emulator threads.
	IOThreadPeriod *DomainCPUTunePeriod          // Enforcement interval for IOThreads.
	IOThreadQuota  *DomainCPUTuneQuota            // Maximum allowed bandwidth for IOThreads.
	VCPUPin        []DomainCPUTuneVCPUPin                // Specifies host physical CPUs for domain vCPUs.
	EmulatorPin    *DomainCPUTuneEmulatorPin         // Specifies host physical CPUs for emulator threads.
	IOThreadPin    []DomainCPUTuneIOThreadPin        // Specifies host physical CPUs for IOThreads.
	VCPUSched      []DomainCPUTuneVCPUSched            // Scheduler type for vCPUs.
	EmulatorSched  *DomainCPUTuneEmulatorSched     // Scheduler type for emulator threads.
	IOThreadSched  []DomainCPUTuneIOThreadSched    // Scheduler type for IOThreads.
	CacheTune      []DomainCPUCacheTune                // Controls allocations for CPU caches.
	MemoryTune     []DomainCPUMemoryTune              // Controls allocations for memory bandwidth.
}

// DomainCPUTuneShares represents the `<shares>` sub-element of `<cputune>`, specifying proportional weighted share.
type DomainCPUTuneShares struct {
	Value uint  // The share value.
}

// DomainCPUTunePeriod represents period values within `<cputune>`, specifying enforcement intervals.
type DomainCPUTunePeriod struct {
	Value uint64  // The period value in microseconds.
}

// DomainCPUTuneQuota represents quota values within `<cputune>`, specifying maximum allowed bandwidth.
type DomainCPUTuneQuota struct {
	Value int64  // The quota value in microseconds.
}

// DomainCPUTuneVCPUPin represents a `<vcpupin>` sub-element of `<cputune>`, pinning a vCPU to host physical CPUs.
type DomainCPUTuneVCPUPin struct {
	VCPU   uint      // vCPU ID.
	CPUSet string  // Physical CPUs to pin to.
}

// DomainCPUTuneEmulatorPin represents the `<emulatorpin>` sub-element of `<cputune>`, pinning emulator threads to host physical CPUs.
type DomainCPUTuneEmulatorPin struct {
	CPUSet string  // Physical CPUs to pin emulator to.
}

// DomainCPUTuneIOThreadPin represents an `<iothreadpin>` sub-element of `<cputune>`, pinning an IOThread to host physical CPUs.
type DomainCPUTuneIOThreadPin struct {
	IOThread uint    // IOThread ID.
	CPUSet   string    // Physical CPUs to pin to.
}

// DomainCPUTuneVCPUSched represents a `<vcpusched>` sub-element of `<cputune>`, specifying scheduler type for vCPUs.
type DomainCPUTuneVCPUSched struct {
	VCPUs     string                 // vCPU IDs this setting applies to.
	Scheduler string   // Scheduler type ("batch", "idle", "fifo", "rr").
	Priority  *int                // Priority for real-time schedulers.
}

// DomainCPUTuneEmulatorSched represents the `<emulatorsched>` sub-element of `<cputune>`, specifying scheduler type for emulator threads.
type DomainCPUTuneEmulatorSched struct {
	Scheduler string  // Scheduler type.
	Priority  *int               // Priority.
}

// DomainCPUTuneIOThreadSched represents an `<iothreadsched>` sub-element of `<cputune>`, specifying scheduler type for IOThreads.
type DomainCPUTuneIOThreadSched struct {
	IOThreads string            // IOThread IDs this setting applies to.
	Scheduler string  // Scheduler type.
	Priority  *int               // Priority.
}

// DomainCPUCacheTune represents a `<cachetune>` sub-element of `<cputune>`, controlling allocations for CPU caches.
type DomainCPUCacheTune struct {
	VCPUs   string                       // vCPUs this allocation applies to.
	ID      string                          // Unique identifier for the cache allocation (output only).
	Cache   []DomainCPUCacheTuneCache                   // Controls allocation of CPU cache.
	Monitor []DomainCPUCacheTuneMonitor               // Creates cache monitor(s).
}

// DomainCPUCacheTuneCache represents a `<cache>` sub-element of `<cachetune>`, configuring CPU cache allocation.
type DomainCPUCacheTuneCache struct {
	ID    uint       // Host cache ID.
	Level uint    // Host cache level.
	Type  string   // Type of allocation ("code", "data", "both").
	Size  uint     // Size of the region to allocate.
	Unit  string .
}

// DomainCPUCacheTuneMonitor represents a `<monitor>` sub-element of `<cachetune>`, creating cache monitors.
type DomainCPUCacheTuneMonitor struct {
	Level uint    // Host cache level the monitor belongs to.
	VCPUs string  // vCPU list the monitor applies to.
}

// DomainCPUMemoryTune represents a `<memorytune>` sub-element of `<cputune>`, controlling allocations for memory bandwidth.
type DomainCPUMemoryTune struct {
	VCPUs   string                       // vCPUs this allocation applies to.
	Nodes   []DomainCPUMemoryTuneNode          // Controls allocation of memory bandwidth per node.
	Monitor []DomainCPUMemoryTuneMonitor     // Creates memory bandwidth monitor(s).
}

// DomainCPUMemoryTuneNode represents a `<node>` sub-element of `<memorytune>`, configuring memory bandwidth allocation per node.
type DomainCPUMemoryTuneNode struct {
	ID        uint         // Host node ID.
	Bandwidth uint  // Memory bandwidth to allocate from this node (e.g., in percent).
}

// DomainCPUMemoryTuneMonitor represents a `<monitor>` sub-element of `<memorytune>`, creating memory bandwidth monitors.
type DomainCPUMemoryTuneMonitor struct {
	Level uint    // Host cache level the monitor belongs to.
	VCPUs string  // vCPU list the monitor applies to.
}

// DomainNUMATune represents the `<numatune>` element, tuning NUMA policy for the domain process.
type DomainNUMATune struct {
	Memory   *DomainNUMATuneMemory      // Specifies how to allocate memory for the domain process on a NUMA host.
	MemNodes []DomainNUMATuneMemNode  // Specifies memory allocation policies per guest NUMA node.
}

// DomainNUMATuneMemory represents the `<memory>` sub-element of `<numatune>`, configuring memory allocation for the domain process.
type DomainNUMATuneMemory struct {
	Mode      string       // Allocation mode ("interleave", "strict", "preferred", "restrictive").
	Nodeset   string    // NUMA nodes to allocate from.
	Placement string  // Memory placement mode ("static", "auto").
}

// DomainNUMATuneMemNode represents a `<memnode>` sub-element of `<numatune>`, specifying memory allocation policies per guest NUMA node.
type DomainNUMATuneMemNode struct {
	CellID  uint     // Guest NUMA node ID.
	Mode    string     // Allocation mode.
	Nodeset string  // NUMA nodes to allocate from.
}

// DomainResource represents the `<resource>` element, grouping configuration related to resource partitioning.
type DomainResource struct {
	Partition    string                          // Absolute path of the resource partition.
	FibreChannel *DomainResourceFibreChannel  // Fibre Channel VMID configuration.
}

// DomainResourceFibreChannel represents the `<fibrechannel>` sub-element of `<resource>`, configuring Fibre Channel VMID.
type DomainResourceFibreChannel struct {
	AppID string  // Application ID for Fibre Channel VMID.
}

// DomainSysInfo represents the `<sysinfo>` element, allowing control over system information presented to the guest.
type DomainSysInfo struct {
	SMBIOS *DomainSysInfoSMBIOS  // SMBIOS system information.
	FWCfg  *DomainSysInfoFWCfg   // Firmware configuration.
}

// DomainSysInfoSMBIOS represents the `<smbios>` sub-element of `<sysinfo>`, configuring SMBIOS system information.
type DomainSysInfoSMBIOS struct {
	BIOS       *DomainSysInfoBIOS              // BIOS information (Block 0).
	System     *DomainSysInfoSystem          // System information (Block 1).
	BaseBoard  []DomainSysInfoBaseBoard   // Base board information (Block 2).
	Chassis    *DomainSysInfoChassis        // Chassis information (Block 3).
	Processor  []DomainSysInfoProcessor   // Processor information.
	Memory     []DomainSysInfoMemory         // Memory information.
	OEMStrings *DomainSysInfoOEMStrings  // OEM strings (Block 11).
}

// DomainSysInfoBIOS represents the `<bios>` sub-element of `<smbios>`, configuring BIOS information.
type DomainSysInfoBIOS struct {
	Entry []DomainSysInfoEntry  // List of BIOS entries.
}

// DomainSysInfoSystem represents the `<system>` sub-element of `<smbios>`, configuring system information.
type DomainSysInfoSystem struct {
	Entry []DomainSysInfoEntry  // List of system entries.
}

// DomainSysInfoBaseBoard represents the `<baseBoard>` sub-element of `<smbios>`, configuring base board information.
type DomainSysInfoBaseBoard struct {
	Entry []DomainSysInfoEntry  // List of base board entries.
}

// DomainSysInfoChassis represents the `<chassis>` sub-element of `<smbios>`, configuring chassis information.
type DomainSysInfoChassis struct {
	Entry []DomainSysInfoEntry  // List of chassis entries.
}

// DomainSysInfoProcessor represents the `<processor>` sub-element of `<smbios>`, configuring processor information.
type DomainSysInfoProcessor struct {
	Entry []DomainSysInfoEntry  // List of processor entries.
}

// DomainSysInfoMemory represents the `<memory>` sub-element of `<smbios>`, configuring memory information.
type DomainSysInfoMemory struct {
	Entry []DomainSysInfoEntry  // List of memory entries.
}

// DomainSysInfoOEMStrings represents the `<oemStrings>` sub-element of `<smbios>`, configuring OEM strings.
type DomainSysInfoOEMStrings struct {
	Entry []string  // List of OEM string entries.
}

// DomainSysInfoEntry represents an `<entry>` sub-element within SMBIOS blocks, configuring a specific SMBIOS field.
type DomainSysInfoEntry struct {
	Name  string             // Name of the SMBIOS field.
	File  string   // Path to a file containing the value.
	Value string             // The value of the SMBIOS field.
}

// DomainSysInfoFWCfg represents the `<fwcfg>` sub-element of `<sysinfo>`, configuring firmware.
type DomainSysInfoFWCfg struct {
	Entry []DomainSysInfoEntry  // List of firmware configuration entries.
}

// DomainOS represents the `<os>` element, specifying the operating system booting configuration.
type DomainOS struct {
	Type         *DomainOSType                  // Type of operating system to be booted.
	Firmware     string                 // Firmware attribute for auto-selection ("bios", "efi").
	FirmwareInfo *DomainOSFirmwareInfo      // Firmware information and features.
	Init         string                 // Path to the init binary for container boot.
	InitArgs     []string                    // Arguments for the init binary.
	InitEnv      []DomainOSInitEnv           // Environment variables for the init binary.
	InitDir      string                 // Custom work directory for the init binary.
	InitUser     string                 // User to run the init command as.
	InitGroup    string                 // Group to run the init command as.
	Loader       *DomainLoader                // Firmware blob used to assist domain creation.
	NVRam        *DomainNVRam                  // Non-volatile memory for UEFI variables.
	Kernel       string                 // Path to the kernel image for direct kernel boot.
	Initrd       string                 // Path to the ramdisk image.
	Cmdline      string                 // Arguments passed to the kernel at boot time.
	Shim         string                 // Path to an initial UEFI bootloader.
	DTB          string                 // Path to the device tree binary image.
	ACPI         *DomainACPI                    // ACPI table configuration.
	BootDevices  []DomainBootDevice             // List of boot devices.
	BootMenu     *DomainBootMenu            // Interactive boot menu prompt settings.
	BIOS         *DomainBIOS                    // BIOS settings.
	SMBios       *DomainSMBios                // SMBIOS mode.
}

// DomainOSType represents the `<type>` sub-element of `<os>`, specifying the type of operating system to be booted.
type DomainOSType struct {
	Arch    string     // CPU architecture.
	Machine string  // Machine type.
	Type    string               // OS type ("hvm", "linux", "exe").
}

// DomainOSFirmwareInfo represents the `<firmware>` sub-element of `<os>`, configuring firmware information and features.
type DomainOSFirmwareInfo struct {
	Features []DomainOSFirmwareFeature  // List of firmware features.
}

// DomainOSFirmwareFeature represents a `<feature>` sub-element of `<firmwareInfo>`, configuring a specific firmware feature.
type DomainOSFirmwareFeature struct {
	Enabled string  // Whether the feature is enabled ("yes", "no").
	Name    string     // Name of the feature ("enrolled-keys", "secure-boot").
}

// DomainOSInitEnv represents an `<initenv>` sub-element of `<os>`, configuring environment variables for the init binary.
type DomainOSInitEnv struct {
	Name  string  // Environment variable name.
	Value string  // Environment variable value.
}

// DomainLoader represents the `<loader>` sub-element of `<os>`, configuring a firmware blob for domain creation.
type DomainLoader struct {
	Path      string           // Absolute path to the firmware blob.
	Readonly  string  // Whether the image is read-only ("yes", "no").
	Secure    string  // Whether the firmware supports Secure Boot ("yes", "no").
	Stateless string  // Controls NVRAM creation ("yes", "no").
	Type      string     // Where in guest memory the file should be mapped ("rom", "pflash").
	Format    string  // Format of the firmware build ("raw", "qcow2").
}

// DomainNVRam represents the `<nvram>` sub-element of `<os>`, configuring non-volatile memory for UEFI variables.
type DomainNVRam struct {
	NVRam          string                      // Absolute path to the NVRAM file.
	Source         *DomainDiskSource              // Source for network-backed NVRAM.
	Template       string             // Path to the master NVRAM store file.
	Format         string             // Format of the NVRAM image.
	TemplateFormat string             // Format for the template file.
}

// DomainACPI represents the `<acpi>` sub-element of `<os>`, configuring ACPI table information.
type DomainACPI struct {
	Tables []DomainACPITable  // List of ACPI tables.
}

// DomainACPITable represents a `<table>` sub-element of `<acpi>`, configuring a specific ACPI table.
type DomainACPITable struct {
	Type string  // Type of ACPI table ("raw", "rawset", "slic", "msdm").
	Path string  // Fully-qualified path to the ACPI table file.
}

// DomainBootDevice represents a `<boot>` sub-element of `<os>`, specifying a boot device.
type DomainBootDevice struct {
	Dev      string               // Boot device ("fd", "hd", "cdrom", "network").
	LoadParm string  // 8-character string for S390 guests.
}

// DomainBootMenu represents the `<bootmenu>` sub-element of `<os>`, configuring interactive boot menu settings.
type DomainBootMenu struct {
	Enable  string   // Whether to enable the boot menu ("yes", "no").
	Timeout string  // Timeout in milliseconds.
}

// DomainBIOS represents the `<bios>` sub-element of `<os>`, configuring BIOS settings.
type DomainBIOS struct {
	UseSerial     string      // Enables Serial Graphics Adapter ("yes", "no").
	RebootTimeout *int              // Timeout for guest reboot on boot failure.
}

// DomainSMBios represents the `<smbios>` sub-element of `<os>`, configuring SMBIOS information population.
type DomainSMBios struct {
	Mode string  // How to populate SMBIOS information ("emulate", "host", "sysinfo").
}

// DomainIDMap represents the `<idmap>` element, enabling user namespace mapping for container-based virtualization.
type DomainIDMap struct {
	UIDs []DomainIDMapRange  // User ID mapping ranges.
	GIDs []DomainIDMapRange  // Group ID mapping ranges.
}

// DomainIDMapRange represents a `<uid>` or `<gid>` sub-element of `<idmap>`, defining an ID mapping range.
type DomainIDMapRange struct {
	Start  uint   // First user/group ID in container.
	Target uint  // Target user/group ID in host.
	Count  uint   // Number of IDs to map.
}

// DomainThrottleGroups represents the `<throttlegroups>` element, allowing creation of multiple named throttle groups for disk I/O.
type DomainThrottleGroups struct {
	ThrottleGroups []ThrottleGroup  // List of named throttle groups.
}

// ThrottleGroup embeds DomainDiskIOTune, representing a named throttle group.
type ThrottleGroup DomainDiskIOTune

// DomainFeatureList represents the `<features>` element, allowing toggling various hypervisor features.
type DomainFeatureList struct {
	PAE           *DomainFeature                         // Physical Address Extension.
	ACPI          *DomainFeature                        // ACPI support.
	APIC          *DomainFeatureAPIC                    // APIC support.
	HAP           *DomainFeatureState                    // Hardware Assisted Paging.
	Viridian      *DomainFeature                    // Viridian hypervisor extensions.
	PrivNet       *DomainFeature                     // Private network namespace.
	HyperV        *DomainFeatureHyperV                // Hyper-V enlightenments.
	KVM           *DomainFeatureKVM                      // KVM hypervisor features.
	Xen           *DomainFeatureXen                      // Xen hypervisor features.
	PVSpinlock    *DomainFeatureState             // Paravirtual spinlocks.
	PMU           *DomainFeatureState                    // Performance Monitoring Unit.
	VMPort        *DomainFeatureState                 // VMware IO port emulation.
	GIC           *DomainFeatureGIC                      // General Interrupt Controller.
	SMM           *DomainFeatureSMM                      // System Management Mode.
	IOAPIC        *DomainFeatureIOAPIC                // I/O APIC tuning.
	HPT           *DomainFeatureHPT                      // Hash Page Table configuration.
	HTM           *DomainFeatureState                    // Hardware Transactional Memory.
	NestedHV      *DomainFeatureState              // Nested HV availability.
	Capabilities  *DomainFeatureCapabilities    // Linux capabilities.
	VMCoreInfo    *DomainFeatureState             // QEMU vmcoreinfo device.
	MSRS          *DomainFeatureMSRS                    // Model Specific Registers.
	CCFAssist     *DomainFeatureState             // Count Cache Flush Assist.
	CFPC          *DomainFeatureCFPC                    // Cache Flush on Privilege Change.
	SBBC          *DomainFeatureSBBC                    // Speculation Barrier Bounds Checking.
	IBS           *DomainFeatureIBS                      // Indirect Branch Speculation.
	TCG           *DomainFeatureTCG                      // TCG accelerator features.
	AsyncTeardown *DomainFeatureAsyncTeardown  // QEMU asynchronous teardown.
	RAS           *DomainFeatureState                    // Report host memory errors.
	PS2           *DomainFeatureState                    // PS/2 controller emulation.
	AIA           *DomainFeatureAIA                      // Advanced Interrupt Architecture for RISC-V.
}

// DomainFeature is a base struct for simple hypervisor features.
type DomainFeature struct {
}

// DomainFeatureAPIC represents the `<apic>` sub-element of `<features>`, configuring APIC support.
type DomainFeatureAPIC struct {
	EOI string  // Toggles EOI availability ("on", "off").
}

// DomainFeatureState is a base struct for hypervisor features with a state attribute.
type DomainFeatureState struct {
	State string  // Feature state ("on", "off").
}

// DomainFeatureHyperV represents the `<hyperv>` sub-element of `<features>`, configuring Hyper-V enlightenments.
type DomainFeatureHyperV struct {
	DomainFeature
	Mode            string                                   // Hyper-V mode ("custom", "passthrough", "host-model").
	Relaxed         *DomainFeatureState                                  // Relax constraints on timers.
	VAPIC           *DomainFeatureState                                    // Enable virtual APIC.
	Spinlocks       *DomainFeatureHyperVSpinlocks                       // Enable spinlock support.
	VPIndex         *DomainFeatureState                                  // Virtual processor index.
	Runtime         *DomainFeatureState                                  // Processor time spent on running guest code.
	Synic           *DomainFeatureState                                    // Enable Synthetic Interrupt Controller (SynIC).
	STimer          *DomainFeatureHyperVSTimer                            // Enable SynIC timers.
	Reset           *DomainFeatureState                                    // Enable hypervisor reset.
	VendorId        *DomainFeatureHyperVVendorId                       // Set hypervisor vendor ID.
	Frequencies     *DomainFeatureState                              // Expose frequency MSRs.
	ReEnlightenment *DomainFeatureState                          // Enable re-enlightenment notification on migration.
	TLBFlush        *DomainFeatureHyperVTLBFlush                        // Enable PV TLB flush support.
	IPI             *DomainFeatureState                                      // Enable PV IPI support.
	EVMCS           *DomainFeatureState                                    // Enable Enlightened VMCS.
	AVIC            *DomainFeatureState                                     // Enable use Hyper-V SynIC with hardware APICv/AVIC.
	EMSRBitmap      *DomainFeatureState                              // Avoid unnecessary updates to L2 MSR Bitmap.
	XMMInput        *DomainFeatureState                                // Enable XMM Fast Hypercall Input.
}

// DomainFeatureHyperVSpinlocks configures spinlock support for Hyper-V.
type DomainFeatureHyperVSpinlocks struct {
	DomainFeatureState
	Retries uint  // Number of failed acquisition attempts before notifying hypervisor.
}

// DomainFeatureHyperVSTimer configures SynIC timers for Hyper-V.
type DomainFeatureHyperVSTimer struct {
	DomainFeatureState
	Direct *DomainFeatureState  // Direct Mode support for SynIC timers.
}

// DomainFeatureHyperVVendorId configures the hypervisor vendor ID for Hyper-V.
type DomainFeatureHyperVVendorId struct {
	DomainFeatureState
	Value string  // Vendor ID string.
}

// DomainFeatureHyperVTLBFlush configures PV TLB flush support for Hyper-V.
type DomainFeatureHyperVTLBFlush struct {
	DomainFeatureState
	Direct   *DomainFeatureState    // Direct TLB flush support.
	Extended *DomainFeatureState  // Extended TLB flush support.
}

// DomainFeatureKVM represents the `<kvm>` sub-element of `<features>`, configuring KVM hypervisor features.
type DomainFeatureKVM struct {
	Hidden        *DomainFeatureState                 // Hide KVM hypervisor from discovery.
	HintDedicated *DomainFeatureState         // Enable optimizations for dedicated vCPUs.
	PollControl   *DomainFeatureState           // Decrease IO completion latency.
	PVIPI         *DomainFeatureState                 // Paravirtualized send IPIs.
	DirtyRing     *DomainFeatureKVMDirtyRing      // Enable dirty ring feature.
}

// DomainFeatureKVMDirtyRing configures the dirty ring feature for KVM.
type DomainFeatureKVMDirtyRing struct {
	DomainFeatureState
	Size uint  // Size of the dirty ring.
}

// DomainFeatureXen represents the `<xen>` sub-element of `<features>`, configuring Xen hypervisor features.
type DomainFeatureXen struct {
	E820Host    *DomainFeatureXenE820Host        // Expose host e820 to guest.
	Passthrough *DomainFeatureXenPassthrough  // Enable IOMMU mappings for PCI passthrough.
}

// DomainFeatureXenE820Host configures host e820 exposure for Xen.
type DomainFeatureXenE820Host struct {
	State string  // State of e820 host exposure ("on", "off").
}

// DomainFeatureXenPassthrough configures IOMMU passthrough for Xen.
type DomainFeatureXenPassthrough struct {
	State string  // State of IOMMU passthrough ("on", "off").
	Mode  string   // Passthrough mode ("sync_pt", "share_pt").
}

// DomainFeatureGIC represents the `<gic>` sub-element of `<features>`, configuring General Interrupt Controller.
type DomainFeatureGIC struct {
	Version string  // GIC version ("2", "3", "host").
}

// DomainFeatureSMM represents the `<smm>` sub-element of `<features>`, configuring System Management Mode.
type DomainFeatureSMM struct {
	State string                 // System Management Mode state ("on", "off").
	TSeg  *DomainFeatureSMMTSeg                  // Amount of memory dedicated to SMM's extended TSEG.
}

// DomainFeatureSMMTSeg configures the extended TSEG size for SMM.
type DomainFeatureSMMTSeg struct {
	Unit  string .
	Value uint              // TSEG size.
}

// DomainFeatureIOAPIC represents the `<ioapic>` sub-element of `<features>`, configuring I/O APIC tuning.
type DomainFeatureIOAPIC struct {
	Driver string  // I/O APIC driver ("kvm", "qemu").
}

// DomainFeatureHPT represents the `<hpt>` sub-element of `<features>`, configuring Hash Page Table.
type DomainFeatureHPT struct {
	Resizing    string                        // HPT resizing ("enabled", "disabled", "required").
	MaxPageSize *DomainFeatureHPTPageSize                 // Limits usable page size.
}

// DomainFeatureHPTPageSize configures the maximum usable page size for HPT guests.
type DomainFeatureHPTPageSize struct {
	Unit  string .
	Value string            // Page size value.
}

// DomainFeatureCapabilities represents the `<capabilities>` sub-element of `<features>`, configuring Linux capabilities.
type DomainFeatureCapabilities struct {
	Policy         string                    // Policy for Linux capabilities.
	AuditControl   *DomainFeatureCapability    // Audit control capability.
	AuditWrite     *DomainFeatureCapability      // Audit write capability.
	BlockSuspend   *DomainFeatureCapability    // Block suspend capability.
	Chown          *DomainFeatureCapability            // Chown capability.
	DACOverride    *DomainFeatureCapability     // DAC override capability.
	DACReadSearch  *DomainFeatureCapability  // DAC read search capability.
	FOwner         *DomainFeatureCapability           // Fowner capability.
	FSetID         *DomainFeatureCapability           // Fsetid capability.
	IPCLock        *DomainFeatureCapability         // IPC lock capability.
	IPCOwner       *DomainFeatureCapability        // IPC owner capability.
	Kill           *DomainFeatureCapability             // Kill capability.
	Lease          *DomainFeatureCapability            // Lease capability.
	LinuxImmutable *DomainFeatureCapability  // Linux immutable capability.
	MACAdmin       *DomainFeatureCapability        // MAC admin capability.
	MACOverride    *DomainFeatureCapability     // MAC override capability.
	MkNod          *DomainFeatureCapability            // MkNod capability.
	NetAdmin       *DomainFeatureCapability        // Net admin capability.
	NetBindService *DomainFeatureCapability  // Net bind service capability.
	NetBroadcast   *DomainFeatureCapability    // Net broadcast capability.
	NetRaw         *DomainFeatureCapability          // Net raw capability.
	SetGID         *DomainFeatureCapability           // Set GID capability.
	SetFCap        *DomainFeatureCapability          // Set FCap capability.
	SetPCap        *DomainFeatureCapability          // Set PCap capability.
	SetUID         *DomainFeatureCapability           // Set UID capability.
	SysAdmin       *DomainFeatureCapability        // Sys admin capability.
	SysBoot        *DomainFeatureCapability         // Sys boot capability.
	SysChRoot      *DomainFeatureCapability       // Sys chroot capability.
	SysModule      *DomainFeatureCapability       // Sys module capability.
	SysNice        *DomainFeatureCapability         // Sys nice capability.
	SysPAcct       *DomainFeatureCapability        // Sys pacct capability.
	SysPTrace      *DomainFeatureCapability       // Sys ptrace capability.
	SysRawIO       *DomainFeatureCapability        // Sys raw IO capability.
	SysResource    *DomainFeatureCapability     // Sys resource capability.
	SysTime        *DomainFeatureCapability         // Sys time capability.
	SysTTYCnofig   *DomainFeatureCapability   // Sys TTY config capability.
	SysLog         *DomainFeatureCapability           // Syslog capability.
	WakeAlarm      *DomainFeatureCapability       // Wake alarm capability.
}

// DomainFeatureCapability represents an individual Linux capability.
type DomainFeatureCapability struct {
	State string  // State of the capability.
}

// DomainFeatureMSRS represents the `<msrs>` sub-element of `<features>`, configuring Model Specific Registers.
type DomainFeatureMSRS struct {
	Unknown string  // Policy for unknown MSRs ("ignore", "fault").
}

// DomainFeatureCFPC represents the `<cfpc>` sub-element of `<features>`, configuring Cache Flush on Privilege Change.
type DomainFeatureCFPC struct {
	Value string  // CFPC availability ("broken", "workaround", "fixed").
}

// DomainFeatureSBBC represents the `<sbbc>` sub-element of `<features>`, configuring Speculation Barrier Bounds Checking.
type DomainFeatureSBBC struct {
	Value string  // SBBC availability ("broken", "workaround", "fixed").
}

// DomainFeatureIBS represents the `<ibs>` sub-element of `<features>`, configuring Indirect Branch Speculation.
type DomainFeatureIBS struct {
	Value string  // IBS availability ("broken", "workaround", "fixed-ibs", "fixed-ccd", "fixed-na").
}

// DomainFeatureTCG represents the `<tcg>` sub-element of `<features>`, configuring TCG accelerator features.
type DomainFeatureTCG struct {
	TBCache *DomainFeatureTCGTBCache  // Translation block cache settings.
}

// DomainFeatureTCGTBCache configures the translation block cache size for TCG.
type DomainFeatureTCGTBCache struct {
	Unit string .
	Size uint              // Size of the translation block cache.
}

// DomainFeatureAsyncTeardown represents the `<async-teardown>` sub-element of `<features>`, configuring QEMU asynchronous teardown.
type DomainFeatureAsyncTeardown struct {
	Enabled string  // Whether asynchronous teardown is enabled ("yes", "no").
}

// DomainFeatureAIA represents the `<aia>` sub-element of `<features>`, configuring Advanced Interrupt Architecture for RISC-V.
type DomainFeatureAIA struct {
	Value string  // AIA configuration for RISC-V ("aplic", "aplic-imsic", "none").
}

// DomainCPU represents the `<cpu>` element, configuring guest CPU requirements.
type DomainCPU struct {
	XMLName            xml.Name              
	Match              string                            // How strictly the virtual CPU matches requirements ("minimum", "exact", "strict").
	Mode               string                             // CPU configuration mode ("custom", "host-model", "host-passthrough", "maximum").
	Check              string                            // Specific way of checking if virtual CPU matches specification ("none", "partial", "full").
	Migratable         string                       // Whether features blocking migration should be removed ("on", "off").
	DeprecatedFeatures string                 // Toggles deprecated CPU model features for S390.
	Model              *DomainCPUModel                                  // Guest CPU model.
	Vendor             string                                // Guest CPU vendor.
	Topology           *DomainCPUTopology                            // Requested topology of virtual CPU.
	Cache              *DomainCPUCache                                  // Virtual CPU cache settings.
	MaxPhysAddr        *DomainCPUMaxPhysAddr                      // Virtual CPU address size in bits.
	Features           []DomainCPUFeature                             // Fine-tune features provided by the CPU model.
	Numa               *DomainNuma                                       // Guest NUMA topology.
}

// DomainCPUModel represents the `<model>` sub-element of `<cpu>`, configuring the guest CPU model.
type DomainCPUModel struct {
	Fallback string  // Forbids fallback to a closest model ("allow", "forbid").
	Value    string                // CPU model name.
	VendorID string  // Vendor ID seen by the guest.
}

// DomainCPUTopology represents the `<topology>` sub-element of `<cpu>`, specifying the requested topology of virtual CPU.
type DomainCPUTopology struct {
	Sockets  int   // Number of CPU sockets.
	Dies     int      // Number of dies per socket.
	Clusters int  // Number of clusters per die.
	Cores    int     // Number of cores per cluster.
	Threads  int   // Number of threads per core.
}

// DomainCPUCache represents the `<cache>` sub-element of `<cpu>`, configuring virtual CPU cache settings.
type DomainCPUCache struct {
	Level *uint   // Cache level.
	Mode  string             // Cache mode ("emulate", "passthrough", "disable").
}

// DomainCPUMaxPhysAddr represents the `<maxphysaddr>` sub-element of `<cpu>`, configuring virtual CPU address size.
type DomainCPUMaxPhysAddr struct {
	Mode  string             // How address size is presented ("passthrough", "emulate").
	Bits  uint     // Virtual CPU address size in bits.
	Limit uint    // Restrict maximum value of address bits for passthrough mode.
}

// DomainCPUFeature represents a `<feature>` sub-element of `<cpu>`, fine-tuning features provided by the CPU model.
type DomainCPUFeature struct {
	Policy string  // Feature policy ("force", "require", "optional", "disable", "forbid").
	Name   string    // Feature name.
}

// DomainNuma represents the `<numa>` sub-element of `<cpu>`, configuring guest NUMA topology.
type DomainNuma struct {
	Cell          []DomainCell                       // List of NUMA cells/nodes.
	Interconnects *DomainNUMAInterconnects  // Describes distances and bandwidth between NUMA cells.
}

// DomainCell represents a `<cell>` sub-element of `<numa>`, configuring a NUMA cell or node.
type DomainCell struct {
	ID        *uint                                // NUMA cell ID.
	CPUs      string                   // CPUs part of the node.
	Memory    uint                             // Node memory.
	Unit      string               .
	MemAccess string                // Memory access mode ("shared", "private").
	Discard   string                // Discard feature for the NUMA node ("yes", "no").
	Distances *DomainCellDistances               // Distances between NUMA cells.
	Caches    []DomainCellCache                      // Memory side cache for memory proximity domains.
}

// DomainCellDistances represents the `<distances>` sub-element of `<cell>`, describing distances between NUMA cells.
type DomainCellDistances struct {
	Siblings []DomainCellSibling  // List of sibling NUMA cells and their distances.
}

// DomainCellSibling represents a `<sibling>` sub-element of `<distances>`, specifying distance to a sibling NUMA cell.
type DomainCellSibling struct {
	ID    uint     // Sibling NUMA cell ID.
	Value uint  // Distance value.
}

// DomainCellCache represents a `<cache>` sub-element of `<cell>`, describing memory side cache for memory proximity domains.
type DomainCellCache struct {
	Level         uint                         // Level of the cache.
	Associativity string               // Cache associativity ("none", "direct", "full").
	Policy        string                      // Cache write associativity ("none", "writeback", "writethrough").
	Size          DomainCellCacheSize                // Cache size.
	Line          DomainCellCacheLine                // Cache line size.
}

// DomainCellCacheSize represents the `<size>` sub-element of `<cache>`, specifying cache size.
type DomainCellCacheSize struct {
	Value string  // Cache size value.
	Unit  string .
}

// DomainCellCacheLine represents the `<line>` sub-element of `<cache>`, specifying cache line size.
type DomainCellCacheLine struct {
	Value string  // Cache line size value.
	Unit  string .
}

// DomainNUMAInterconnects represents the `<interconnects>` sub-element of `<numa>`, describing distances and bandwidth between NUMA cells.
type DomainNUMAInterconnects struct {
	Latencies  []DomainNUMAInterconnectLatency      // List of latency descriptions.
	Bandwidths []DomainNUMAInterconnectBandwidth  // List of bandwidth descriptions.
}

// DomainNUMAInterconnectLatency represents a `<latency>` sub-element of `<interconnects>`, describing latency between NUMA nodes.
type DomainNUMAInterconnectLatency struct {
	Initiator uint           // Source NUMA node.
	Target    uint              // Target NUMA node.
	Cache     uint     // Target NUMA node's cache level.
	Type      string              // Type of access ("access", "read", "write").
	Value     uint               // Delay in nanoseconds.
}

// DomainNUMAInterconnectBandwidth represents a `<bandwidth>` sub-element of `<interconnects>`, describing bandwidth between NUMA nodes.
type DomainNUMAInterconnectBandwidth struct {
	Initiator uint           // Source NUMA node.
	Target    uint              // Target NUMA node.
	Cache     uint     // Target NUMA node's cache level.
	Type      string              // Type of access ("access", "read", "write").
	Value     uint               // Bandwidth value.
	Unit      string .
}

// DomainClock represents the `<clock>` element, controlling how the guest clock is synchronized to the host clock and timer settings.
type DomainClock struct {
	Offset     string             // How guest clock is synchronized ("utc", "localtime", "timezone", "variable", "absolute").
	Basis      string              // Basis for variable offset ("utc", "localtime").
	Adjustment string         // Delta relative to UTC/localtime.
	TimeZone   string           // Timezone for synchronization.
	Start      uint                // Epoch timestamp for absolute offset.
	Timer      []DomainTimer                      // List of timer configurations.
}

// DomainTimer represents a `<timer>` sub-element of `<clock>`, configuring a specific timer.
type DomainTimer struct {
	Name       string                            // Timer name ("platform", "hpet", "kvmclock", "pit", "rtc", "tsc", "hypervclock", "armvtimer").
	Track      string                 // What the timer tracks ("boot", "guest", "wall", "realtime").
	TickPolicy string               // Behavior when QEMU misses a tick deadline ("delay", "catchup", "merge", "discard").
	CatchUp    *DomainTimerCatchUp                 // Catchup settings for "catchup" policy.
	Frequency  uint64               // Frequency at which "tsc" runs.
	Mode       string                  // How "tsc" timer is managed ("auto", "native", "emulate", "paravirt", "smpsafe").
	Present    string               // Whether timer is available to guest ("yes", "no").
}

// DomainTimerCatchUp configures catchup settings for the "catchup" tick policy.
type DomainTimerCatchUp struct {
	Threshold uint  // Threshold for catchup.
	Slew      uint       // Slew rate for catchup.
	Limit     uint      // Limit for catchup.
}

// DomainPM represents the `<pm>` element, configuring BIOS advertisements for power management.
type DomainPM struct {
	SuspendToMem  *DomainPMPolicy   // BIOS support for S3 (suspend-to-mem).
	SuspendToDisk *DomainPMPolicy  // BIOS support for S4 (suspend-to-disk).
}

// DomainPMPolicy represents a power management policy within `<pm>`.
type DomainPMPolicy struct {
	Enabled string  // Whether the policy is enabled ("yes", "no").
}

// DomainPerf represents the `<perf>` element, allowing monitoring of performance events for the virtual machine.
type DomainPerf struct {
	Events []DomainPerfEvent  // List of performance monitoring events.
}

// DomainPerfEvent represents an `<event>` sub-element of `<perf>`, configuring a specific performance event.
type DomainPerfEvent struct {
	Name    string     // Event name.
	Enabled string  // Whether the event is enabled ("yes", "no").
}

// DomainDeviceList represents the `<devices>` element, the main container for describing all devices provided to the guest domain.
type DomainDeviceList struct {
	Emulator     string                   // Path to the device model emulator binary.
	Disks        []DomainDisk                           // List of disk devices.
	Controllers  []DomainController               // List of controller devices.
	Leases       []DomainLease                         // List of device leases.
	Filesystems  []DomainFilesystem               // List of filesystem devices.
	Interfaces   []DomainInterface                 // List of network interfaces.
	Smartcards   []DomainSmartcard                 // List of smartcard devices.
	Serials      []DomainSerial                       // List of serial devices.
	Parallels    []DomainParallel                   // List of parallel devices.
	Consoles     []DomainConsole                     // List of console devices.
	Channels     []DomainChannel                     // List of channel devices.
	Inputs       []DomainInput                         // List of input devices.
	TPMs         []DomainTPM                             // List of TPM devices.
	Graphics     []DomainGraphic                    // List of graphical framebuffers.
	Sounds       []DomainSound                         // List of sound devices.
	Audios       []DomainAudio                         // List of audio backends.
	Videos       []DomainVideo                         // List of video devices.
	Hostdevs     []DomainHostdev                     // List of host device assignments.
	RedirDevs    []DomainRedirDev                   // List of redirected devices.
	RedirFilters []DomainRedirFilter             // List of redirection filters.
	Hubs         []DomainHub                             // List of hub devices.
	Watchdogs    []DomainWatchdog                   // List of watchdog devices.
	MemBalloon   *DomainMemBalloon                // Memory balloon device.
	RNGs         []DomainRNG                             // List of random number generator devices.
	NVRAM        *DomainNVRAM                          // NVRAM device.
	Panics       []DomainPanic                         // List of panic devices.
	Shmems       []DomainShmem                         // List of shared memory devices.
	Memorydevs   []DomainMemorydev                    // List of memory devices.
	IOMMU        *DomainIOMMU                          // IOMMU device.
	VSock        *DomainVSock                          // Vsock device.
	Crypto       []DomainCrypto                       // List of crypto devices.
	PStore       *DomainPStore                        // Pstore device.
}

// DomainDisk represents a `<disk>` element, configuring disk devices.
type DomainDisk struct {
	XMLName         xml.Name                
	Device          string                            // How the disk is exposed to the guest OS ("floppy", "disk", "cdrom", "lun").
	RawIO           string                             // Whether the disk needs rawio capability ("yes", "no").
	SGIO            string                              // Whether unprivileged SG_IO commands are filtered ("filtered", "unfiltered").
	Snapshot        string                          // Default behavior during disk snapshots ("internal", "external", "no", "manual").
	Model           string                             // Emulated device model of the disk.
	Driver          *DomainDiskDriver                                // Hypervisor driver details.
	Auth            *DomainDiskAuth                                    // Authentication credentials.
	Source          *DomainDiskSource                                // Underlying source for the disk.
	BackingStore    *DomainDiskBackingStore                    // Backing store for copy-on-write.
	BackendDomain   *DomainBackendDomain                      // Backend domain hosting the disk.
	Geometry        *DomainDiskGeometry                            // Overrides geometry settings.
	BlockIO         *DomainDiskBlockIO                              // Overrides block device properties.
	Mirror          *DomainDiskMirror                                // Long-running block job operation.
	Target          *DomainDiskTarget                                // Bus/device under which disk is exposed to guest.
	IOTune          *DomainDiskIOTune                                // Per-device I/O tuning.
	ThrottleFilters *ThrottleFilters                        // Additional per-device throttle chain.
	ReadOnly        *DomainDiskReadOnly                            // Device cannot be modified by guest.
	Shareable       *DomainDiskShareable                          // Device is shared between domains.
	Transient       *DomainDiskTransient                          // Changes reverted on guest exit.
	Serial          string                                 // Serial number of virtual hard drive.
	WWN             string                                    // WWN of virtual hard disk or CD-ROM.
	Vendor          string                                 // Vendor of virtual hard disk or CD-ROM.
	Product         string                                // Product of virtual hard disk or CD-ROM.
	Encryption      *DomainDiskEncryption                        // Specifies how the volume is encrypted.
	Boot            *DomainDeviceBoot                                  // Specifies that the disk is bootable.
	ACPI            *DomainDeviceACPI                                  // ACPI device configuration.
	Alias           *DomainAlias                                      // Identifier for the device.
	Address         *DomainAddress                                  // Device address on the virtual bus.
}

// DomainDiskDriver represents the `<driver>` sub-element of `<disk>`, specifying hypervisor driver details for a disk.
type DomainDiskDriver struct {
	Name           string                               // Primary backend driver name.
	Type           string                               // Sub-type of the backend driver.
	Cache          string                              // Cache mechanism ("default", "none", "writethrough", "writeback", "directsync", "unsafe").
	ErrorPolicy    string                       // Behavior on disk read/write error ("stop", "report", "ignore", "enospace").
	RErrorPolicy   string                      // Behavior for read errors only.
	IO             string                                 // Specific policies on I/O ("threads", "native", "io_uring").
	IOEventFD      string                          // Enables/disables I/O asynchronous handling ("on", "off").
	EventIDX       string                          // Controls device event processing ("on", "off").
	CopyOnRead     string                       // Whether to copy read backing file into image file ("on", "off").
	Discard        string                            // Whether discard requests are ignored or passed ("unmap", "ignore").
	DiscardNoUnref string                    // Handles guest discard commands inside qcow2 image.
	IOThread       *uint                                      // Assigns disk to an IOThread.
	IOThreads      *DomainDiskIOThreads                           // Specifies multiple IOThreads.
	DetectZeros    string                      // Whether to detect zero write requests ("off", "on", "unmap").
	Queues         *uint                                        // Number of virt queues.
	QueueSize      *uint                                    // Size of each virt queue.
	IOMMU          string                              // Enables emulated IOMMU.
	ATS            string                                // Address Translation Service support.
	Packed         string                             // Whether to use packed virtqueues.
	PagePerVQ      string                        // Layout of notification capabilities.
	MetadataCache  *DomainDiskMetadataCache                  // Format specific caching of storage image metadata.
}

// DomainDiskIOThreads represents the `<iothreads>` sub-element of `<driver>`, specifying multiple IOThreads for a disk.
type DomainDiskIOThreads struct {
	IOThread []DomainDiskIOThread  // List of IOThread configurations.
}

// DomainDiskIOThread represents an `<iothread>` sub-element of `<iothreads>`, configuring an individual IOThread for a disk.
type DomainDiskIOThread struct {
	ID     uint                            // IOThread identifier.
	Queues []DomainDiskIOThreadQueue       // List of queue mappings.
}

// DomainDiskIOThreadQueue represents a `<queue>` sub-element of `<iothread>`, mapping an IOThread to a virt queue.
type DomainDiskIOThreadQueue struct {
	ID uint  // Queue ID.
}

// DomainDiskMetadataCache represents the `<metadata_cache>` sub-element of `<driver>`, configuring metadata caching.
type DomainDiskMetadataCache struct {
	MaxSize *DomainDiskMetadataCacheSize  // Maximum size of the metadata cache.
}

// DomainDiskMetadataCacheSize represents the `<max_size>` sub-element of `<metadata_cache>`, specifying metadata cache size.
type DomainDiskMetadataCacheSize struct {
	Unit  string .
	Value int                  // Size value.
}

// DomainDiskAuth represents the `<auth>` sub-element of `<disk>`, providing authentication credentials.
type DomainDiskAuth struct {
	Username string             // Username for authentication.
	Secret   *DomainDiskSecret                   // Reference to a libvirt secret object.
}

// DomainDiskSecret represents the `<secret>` sub-element of `<auth>`, referencing a libvirt secret object.
type DomainDiskSecret struct {
	Type  string  // Secret type ("ceph", "iscsi").
	Usage string  // Usage key of the secret object.
	UUID  string   // UUID of the secret object.
}

// DomainDiskSource represents the `<source>` sub-element of `<disk>`, specifying the underlying source for the disk.
type DomainDiskSource struct {
	File          *DomainDiskSourceFile                            // File-backed disk source.
	Block         *DomainDiskSourceBlock                           // Block device-backed disk source.
	Dir           *DomainDiskSourceDir                             // Directory-backed disk source.
	Network       *DomainDiskSourceNetwork                         // Network-backed disk source.
	Volume        *DomainDiskSourceVolume                          // Storage volume-backed disk source.
	NVME          *DomainDiskSourceNVME                            // NVMe disk source.
	VHostUser     *DomainDiskSourceVHostUser                       // vhost-user disk source.
	VHostVDPA     *DomainDiskSourceVHostVDPA                       // vhost-vdpa disk source.
	StartupPolicy string                      // Policy if source file is not accessible ("mandatory", "requisite", "optional").
	Index         uint                          // Index for referring to a specific part of the disk chain.
	Encryption    *DomainDiskEncryption                   // Encryption for storage sources.
	Reservations  *DomainDiskReservations               // Enables persistent reservations for SCSI disks.
	Slices        *DomainDiskSlices                           // Configures offset and size of image format.
	SSL           *DomainDiskSourceSSL                           // SSL transport parameters for HTTPS/FTPS.
	Cookies       *DomainDiskCookies                         // Cookies for HTTP/HTTPS.
	Readahead     *DomainDiskSourceReadahead               // Size of readahead buffer.
	Timeout       *DomainDiskSourceTimeout                   // Connection timeout.
	DataStore     *DomainDiskDataStore                     // External data store.
}

// DomainDiskSourceFile represents a `<file>` sub-element of `<source>`, specifying a file-backed disk source.
type DomainDiskSourceFile struct {
	File     string                  // Fully-qualified path to the file.
	FDGroup  string                  // Access disk via file descriptors.
	SecLabel []DomainDeviceSecLabel             // Security label override.
}

// DomainDiskSourceBlock represents a `<block>` sub-element of `<source>`, specifying a block device-backed disk source.
type DomainDiskSourceBlock struct {
	Dev      string                  // Fully-qualified path to the host device.
	SecLabel []DomainDeviceSecLabel            // Security label override.
}

// DomainDiskSourceDir represents a `<dir>` sub-element of `<source>`, specifying a directory-backed disk source.
type DomainDiskSourceDir struct {
	Dir string  // Fully-qualified path to the directory.
}

// DomainDiskSourceNetwork represents a `<network>` sub-element of `<source>`, specifying a network-backed disk source.
type DomainDiskSourceNetwork struct {
	Protocol    string                               // Protocol to access image ("nbd", "iscsi", "rbd", "sheepdog", "gluster", "vxhs", "nfs", "http", "https", "ftp", "ftps", "tftp", "ssh").
	Name        string                                   // Volume/image name.
	Query       string                                  // Query string for HTTP/HTTPS.
	TLS         string                                    // TLS transport for NBD/VxHS ("yes", "no").
	TLSHostname string                            // Overrides expected hostname for TLS certificate verification.
	Hosts       []DomainDiskSourceHost                                  // List of hosts to connect.
	Identity    *DomainDiskSourceNetworkIdentity                    // User/group configuration for NFS, SSH authentication.
	KnownHosts  *DomainDiskSourceNetworkKnownHosts                  // Path to file for remote host verification (SSH).
	Initiator   *DomainDiskSourceNetworkInitiator                   // Initiator IQN for iSCSI.
	Snapshot    *DomainDiskSourceNetworkSnapshot                    // Internal snapshot name.
	Config      *DomainDiskSourceNetworkConfig                        // Path to configuration file.
	Reconnect   *DomainDiskSourceNetworkReconnect                   // Reconnect timeout.
	Auth        *DomainDiskAuth                                         // Authentication credentials.
}

// DomainDiskSourceHost represents a `<host>` sub-element of `<network>`, specifying a host to connect for network-backed disks.
type DomainDiskSourceHost struct {
	Transport string  // Transport type ("tcp", "rdma", "unix").
	Name      string       // Hostname or IP address.
	Port      string       // Port number.
	Socket    string     // Path to AF_UNIX socket.
}

// DomainDiskSourceNetworkIdentity represents the `<identity>` sub-element of `<network>`, configuring user/group for NFS or SSH authentication.
type DomainDiskSourceNetworkIdentity struct {
	User      string       // User for NFS.
	Group     string      // Group for NFS.
	UserName  string   // Username for SSH.
	Keyfile   string    // Path to SSH key file.
	AgentSock string  // Path to SSH agent socket.
}

// DomainDiskSourceNetworkKnownHosts represents the `<knownHosts>` sub-element of `<network>`, configuring path to known_hosts file for SSH.
type DomainDiskSourceNetworkKnownHosts struct {
	Path string  // Path to the known_hosts file.
}

// DomainDiskSourceNetworkInitiator represents the `<initiator>` sub-element of `<network>`, configuring initiator IQN for iSCSI.
type DomainDiskSourceNetworkInitiator struct {
	IQN *DomainDiskSourceNetworkIQN  // Initiator IQN.
}

// DomainDiskSourceNetworkIQN represents the `<iqn>` sub-element of `<initiator>`, specifying initiator IQN name.
type DomainDiskSourceNetworkIQN struct {
	Name string  // Initiator IQN name.
}

// DomainDiskSourceNetworkSnapshot represents the `<snapshot>` sub-element of `<network>`, specifying an internal snapshot name.
type DomainDiskSourceNetworkSnapshot struct {
	Name string  // Internal snapshot name.
}

// DomainDiskSourceNetworkConfig represents the `<config>` sub-element of `<network>`, specifying a configuration file for network storage.
type DomainDiskSourceNetworkConfig struct {
	File string  // Path to configuration file.
}

// DomainDiskSourceNetworkReconnect represents the `<reconnect>` sub-element of `<network>`, configuring reconnect timeout.
type DomainDiskSourceNetworkReconnect struct {
	Delay string  // Reconnect delay in seconds.
}

// DomainDiskSourceVolume represents a `<volume>` sub-element of `<source>`, specifying a storage volume-backed disk source.
type DomainDiskSourceVolume struct {
	Pool     string                      // Name of the storage pool.
	Volume   string                    // Name of the storage volume.
	Mode     string                      // How to represent the LUN ("direct", "host").
	SecLabel []DomainDeviceSecLabel                 // Security label override.
}

// DomainDiskSourceNVME represents an `<nvme>` sub-element of `<source>`, specifying an NVMe disk source.
type DomainDiskSourceNVME struct {
	PCI *DomainDiskSourceNVMEPCI // PCI address of the host NVMe controller.
}

// DomainDiskSourceNVMEPCI represents the `<pci>` sub-element of `<nvme>`, specifying the PCI address of the host NVMe controller.
type DomainDiskSourceNVMEPCI struct {
	Managed   string             // Detach NVMe controller automatically ("yes", "no").
	Namespace uint64             // Namespace ID.
	Address   *DomainAddressPCI                 // PCI address of the NVMe controller.
}

// DomainDiskSourceVHostUser embeds DomainChardevSource, representing a vhost-user disk source.
type DomainDiskSourceVHostUser DomainChardevSource

// DomainDiskSourceVHostVDPA represents a `<vhostvdpa>` sub-element of `<source>`, specifying a vhost-vdpa disk source.
type DomainDiskSourceVHostVDPA struct {
	Dev string  // Path to the vhost-vdpa character device.
}

// DomainDiskEncryption represents the `<encryption>` sub-element of `<source>`, configuring encryption for storage sources.
type DomainDiskEncryption struct {
	Format  string             // Encryption format ("luks", "luks2", "luks-any").
	Engine  string             // Component handling encryption ("qemu", "librbd").
	Secrets []DomainDiskSecret                 // List of secrets for encryption.
}

// DomainDiskReservations represents the `<reservations>` sub-element of `<source>`, enabling persistent reservations for SCSI disks.
type DomainDiskReservations struct {
	Enabled string                         // Enables persistent reservations ("yes", "no").
	Managed string                         // Libvirt manages resources ("yes", "no").
	Source  *DomainDiskReservationsSource                  // Source for unmanaged reservations.
}

// DomainDiskReservationsSource embeds DomainChardevSource, specifying the source for unmanaged reservations.
type DomainDiskReservationsSource DomainChardevSource

// DomainDiskSlices represents the `<slices>` sub-element of `<source>`, configuring offset and size of image format.
type DomainDiskSlices struct {
	Slices []DomainDiskSlice  // List of slice configurations.
}

// DomainDiskSlice represents a `<slice>` sub-element of `<slices>`, configuring a specific slice.
type DomainDiskSlice struct {
	Type   string    // Type of slice ("storage").
	Offset uint    // Offset in bytes.
	Size   uint      // Size in bytes.
}

// DomainDiskSourceSSL represents the `<ssl>` sub-element of `<source>`, configuring SSL transport parameters.
type DomainDiskSourceSSL struct {
	Verify string  // SSL certificate validation ("yes", "no").
}

// DomainDiskCookies represents the `<cookies>` sub-element of `<source>`, configuring cookies for HTTP/HTTPS.
type DomainDiskCookies struct {
	Cookies []DomainDiskCookie  // List of cookies.
}

// DomainDiskCookie represents a `<cookie>` sub-element of `<cookies>`, configuring a specific cookie.
type DomainDiskCookie struct {
	Name  string  // Cookie name.
	Value string  // Cookie value.
}

// DomainDiskSourceReadahead represents the `<readahead>` sub-element of `<source>`, configuring readahead buffer size.
type DomainDiskSourceReadahead struct {
	Size string  // Size of the readahead buffer in bytes.
}

// DomainDiskSourceTimeout represents the `<timeout>` sub-element of `<source>`, configuring connection timeout.
type DomainDiskSourceTimeout struct {
	Seconds string  // Connection timeout in seconds.
}

// DomainDiskDataStore represents the `<dataStore>` sub-element of `<source>`, configuring an external data store.
type DomainDiskDataStore struct {
	Format *DomainDiskFormat  // Format of the data store.
	Source *DomainDiskSource  // Location of the data store.
}

// DomainDiskBackingStore represents the `<backingStore>` sub-element of `<disk>`, configuring a backing store for copy-on-write.
type DomainDiskBackingStore struct {
	Index        uint                     // Index for referring to a specific part of the disk chain (output only).
	Format       *DomainDiskFormat                      // Internal format of the backing store.
	Source       *DomainDiskSource                      // Location of the backing store.
	BackingStore *DomainDiskBackingStore          // Nested backing store for chained images.
}

// DomainDiskFormat represents the `<format>` sub-element of `<backingStore>`, specifying the internal format of the backing store.
type DomainDiskFormat struct {
	Type          string                             // Internal format type (e.g., "raw", "qcow2").
	MetadataCache *DomainDiskMetadataCache      // Metadata cache settings.
}

// DomainBackendDomain represents the `<backenddomain>` sub-element of `<disk>`, specifying a backend domain hosting the disk.
type DomainBackendDomain struct {
	Name string  // Name of the backend domain.
}

// DomainDiskGeometry represents the `<geometry>` sub-element of `<disk>`, overriding geometry settings.
type DomainDiskGeometry struct {
	Cylinders uint             // Number of cylinders.
	Headers   uint            // Number of heads.
	Sectors   uint             // Number of sectors per track.
	Trans     string  // BIOS-Translation-Modus ("none", "lba", "auto").
}

// DomainDiskBlockIO represents the `<blockio>` sub-element of `<disk>`, overriding block device properties.
type DomainDiskBlockIO struct {
	LogicalBlockSize   uint     // Logical block size reported to guest.
	PhysicalBlockSize  uint    // Physical block size reported to guest.
	DiscardGranularity *uint             // Smallest amount of data that can be discarded.
}

// DomainDiskMirror represents the `<mirror>` sub-element of `<disk>`, configuring a long-running block job operation.
type DomainDiskMirror struct {
	Job          string                            // API that started the operation ("copy", "active-commit").
	Ready        string                          // Tracks progress of the job ("yes", "abort", "pivot").
	Format       *DomainDiskFormat                             // File format of the mirror.
	Source       *DomainDiskSource                             // Location of the mirror.
	BackingStore *DomainDiskBackingStore                 // Backing store for the mirror.
}

// DomainDiskTarget represents the `<target>` sub-element of `<disk>`, configuring the bus/device under which the disk is exposed to the guest.
type DomainDiskTarget struct {
	Dev          string           // Logical device name.
	Bus          string           // Type of disk device to emulate.
	Tray         string          // Tray status of removable disks ("open", "closed").
	Removable    string     // Removable flag for USB/SCSI disks ("on", "off").
	RotationRate uint    // Rotation rate of the storage.
}

// DomainDiskIOTune represents the `<iotune>` sub-element of `<disk>`, providing per-device I/O tuning.
type DomainDiskIOTune struct {
	TotalBytesSec          uint64           // Total throughput limit in bytes/sec.
	ReadBytesSec           uint64            // Read throughput limit in bytes/sec.
	WriteBytesSec          uint64           // Write throughput limit in bytes/sec.
	TotalIopsSec           uint64            // Total I/O operations per second.
	ReadIopsSec            uint64             // Read I/O operations per second.
	WriteIopsSec           uint64            // Write I/O operations per second.
	TotalBytesSecMax       uint64       // Maximum total throughput limit.
	ReadBytesSecMax        uint64        // Maximum read throughput limit.
	WriteBytesSecMax       uint64       // Maximum write throughput limit.
	TotalIopsSecMax        uint64        // Maximum total I/O operations per second.
	ReadIopsSecMax         uint64         // Maximum read I/O operations per second.
	WriteIopsSecMax        uint64        // Maximum write I/O operations per second.
	TotalBytesSecMaxLength uint64  // Max duration for total_bytes_sec_max burst.
	ReadBytesSecMaxLength  uint64   // Max duration for read_bytes_sec_max burst.
	WriteBytesSecMaxLength uint64  // Max duration for write_bytes_sec_max burst.
	TotalIopsSecMaxLength  uint64   // Max duration for total_iops_sec_max burst.
	ReadIopsSecMaxLength   uint64    // Max duration for read_iops_sec_max burst.
	WriteIopsSecMaxLength  uint64   // Max duration for write_iops_sec_max burst.
	SizeIopsSec            uint64             // Size of I/O operations per second.
	GroupName              string                // Group name for sharing I/O throttling quota.
}

// ThrottleFilters represents the `<throttlefilters>` sub-element of `<disk>`, providing additional per-device throttle chains.
type ThrottleFilters struct {
	ThrottleFilter []ThrottleFilter  // List of throttle filter references.
}

// ThrottleFilter represents a `<throttlefilter>` sub-element of `<throttlefilters>`, referencing a defined throttle group.
type ThrottleFilter struct {
	Group string  // Name of the throttle group to reference.
}

// DomainDiskReadOnly represents the `<readonly>` sub-element of `<disk>`, indicating the device cannot be modified by the guest.
type DomainDiskReadOnly struct {
}

// DomainDiskShareable represents the `<shareable>` sub-element of `<disk>`, indicating the device is shared between domains.
type DomainDiskShareable struct {
}

// DomainDiskTransient represents the `<transient>` sub-element of `<disk>`, indicating changes are reverted on guest exit.
type DomainDiskTransient struct {
	ShareBacking string  // Whether backing image is shared ("yes", "no").
}

// DomainDeviceBoot represents a `<boot>` sub-element, specifying that the device is bootable.
type DomainDeviceBoot struct {
	Dev      string               // Boot device.
	LoadParm string  // Load parameter for S390.
}

// DomainDeviceACPI represents the `<acpi>` sub-element for device configuration.
type DomainDeviceACPI struct {
	Index uint  // ACPI index.
}

// DomainAlias represents the `<alias>` sub-element, providing a user-defined identifier for a device.
type DomainAlias struct {
	Name string  // Identifier for the device.
}

// DomainAddress represents the `<address>` sub-element, specifying the device's placement on the virtual bus.
type DomainAddress struct {
	PCI          *DomainAddressPCI          // PCI address.
	Drive        *DomainAddressDrive        // Drive address.
	VirtioSerial *DomainAddressVirtioSerial // Virtio-serial address.
	CCID         *DomainAddressCCID         // CCID address.
	USB          *DomainAddressUSB          // USB address.
	SpaprVIO     *DomainAddressSpaprVIO     // SPAPR-VIO address.
	VirtioS390   *DomainAddressVirtioS390   // Virtio-S390 address.
	CCW          *DomainAddressCCW          // CCW address.
	VirtioMMIO   *DomainAddressVirtioMMIO   // Virtio-MMIO address.
	ISA          *DomainAddressISA          // ISA address.
	DIMM         *DomainAddressDIMM         // DIMM address.
	Unassigned   *DomainAddressUnassigned   // Unassigned address.
}

// DomainAddressPCI represents a PCI address.
type DomainAddressPCI struct {
	Domain        *uint                      // PCI domain.
	Bus           *uint                         // PCI bus.
	Slot          *uint                        // PCI slot.
	Function      *uint                    // PCI function.
	MultiFunction string              // Multifunction bit ("on", "off").
	ZPCI          *DomainAddressZPCI                // ZPCI attributes for S390.
}

// DomainAddressZPCI represents ZPCI attributes for S390 PCI devices.
type DomainAddressZPCI struct {
	UID *uint  // User-defined Identifier.
	FID *uint  // Function Identifier.
}

// DomainAddressDrive represents a drive address.
type DomainAddressDrive struct {
	Controller *uint  // Controller number.
	Bus        *uint         // Bus number.
	Target     *uint      // Target number.
	Unit       *uint        // Unit number.
}

// DomainAddressVirtioSerial represents a virtio-serial address.
type DomainAddressVirtioSerial struct {
	Controller *uint  // Controller number.
	Bus        *uint         // Bus number.
	Port       *uint        // Slot within the bus.
}

// DomainAddressCCID represents a CCID address for smart-cards.
type DomainAddressCCID struct {
	Controller *uint  // Controller number.
	Slot       *uint        // Slot within the bus.
}

// DomainAddressUSB represents a USB address.
type DomainAddressUSB struct {
	Bus    *uint            // USB bus number.
	Port   string  // Port on the USB bus.
	Device *uint         // Device number.
}

// DomainAddressSpaprVIO represents a SPAPR-VIO bus address for PowerPC pseries guests.
type DomainAddressSpaprVIO struct {
	Reg *uint64  // Hex value address of the starting register.
}

// DomainAddressVirtioS390 represents a virtio-S390 address.
type DomainAddressVirtioS390 struct {
}

// DomainAddressCCW represents a CCW bus address for S390 guests.
type DomainAddressCCW struct {
	CSSID *uint  // Channel subsystem identifier.
	SSID  *uint   // Subchannel-set identifier.
	DevNo *uint  // Device number.
}

// DomainAddressVirtioMMIO represents a virtio-mmio address.
type DomainAddressVirtioMMIO struct {
}

// DomainAddressISA represents an ISA address.
type DomainAddressISA struct {
	IOBase *uint  // I/O base address.
	IRQ    *uint     // IRQ number.
}

// DomainAddressDIMM represents a DIMM address.
type DomainAddressDIMM struct {
	Slot *uint    // DIMM slot number.
	Base *uint64  // Base address.
}

// DomainAddressUnassigned represents an unassigned device address.
type DomainAddressUnassigned struct {
}

// DomainController represents a `<controller>` element, configuring device bus controllers.
type DomainController struct {
	XMLName      xml.Name                      
	Type         string                                    // Controller type ("ide", "fdc", "scsi", "sata", "usb", "ccid", "virtio-serial", "pci", "xenbus", "nvme").
	Index        *uint                                    // Order of the bus controller.
	Model        string                         // Controller model.
	Driver       *DomainControllerDriver                      // Driver specific options.
	PCI          *DomainControllerPCI                              // PCI controller specific settings.
	USB          *DomainControllerUSB                              // USB controller specific settings.
	VirtIOSerial *DomainControllerVirtIOSerial                     // Virtio-serial controller specific settings.
	XenBus       *DomainControllerXenBus                           // Xenbus controller specific settings.
	NVME         *DomainControllerNVME                             // NVMe controller specific settings.
	ACPI         *DomainDeviceACPI                              // ACPI device configuration.
	Alias        *DomainAlias                                  // Identifier for the device.
	Address      *DomainAddress                              // Device address on the virtual bus.
}

// DomainControllerDriver represents the `<driver>` sub-element of `<controller>`, specifying driver options for a controller.
type DomainControllerDriver struct {
	Queues     *uint                                           // Number of queues for the controller.
	CmdPerLUN  *uint                                      // Max commands queued per LUN.
	MaxSectors *uint                                      // Max data transferred in a single command.
	IOEventFD  string                               // I/O asynchronous handling ("on", "off").
	IOThread   uint                                 // Assigns controller to an IOThread.
	IOMMU      string                                 // Enables emulated IOMMU.
	ATS        string                                   // Address Translation Service support.
	Packed     string                                // Whether to use packed virtqueues.
	PagePerVQ  string                               // Layout of notification capabilities.
	IOThreads  *DomainControllerDriverIOThreads               // Specifies multiple IOThreads.
}

// DomainControllerDriverIOThreads represents the `<iothreads>` sub-element of `<driver>`, specifying multiple IOThreads for a controller.
type DomainControllerDriverIOThreads struct {
	IOThread []DomainControllerDriverIOThread  // List of IOThread configurations.
}

// DomainControllerDriverIOThread represents an `<iothread>` sub-element of `<iothreads>`, configuring an individual IOThread for a controller.
type DomainControllerDriverIOThread struct {
	ID     uint                                      // IOThread identifier.
	Queues []DomainControllerDriverIOThreadQueue       // List of queue mappings.
}

// DomainControllerDriverIOThreadQueue represents a `<queue>` sub-element of `<iothread>`, mapping an IOThread to a virt queue.
type DomainControllerDriverIOThreadQueue struct {
	ID uint  // Queue ID.
}

// DomainControllerPCI represents the `<pci>` sub-element of `<controller>`, configuring PCI controller specific settings.
type DomainControllerPCI struct {
	Model  *DomainControllerPCIModel     // Name of the specific device QEMU is emulating.
	Target *DomainControllerPCITarget   // Configurable items visible to guest OS.
	Hole64 *DomainControllerPCIHole64  // Size of the 64-bit PCI hole.
}

// DomainControllerPCIModel represents the `<model>` sub-element of `<pci>`, specifying the device model name.
type DomainControllerPCIModel struct {
	Name string  // Device model name.
}

// DomainControllerPCITarget represents the `<target>` sub-element of `<pci>`, configuring PCI target properties.
type DomainControllerPCITarget struct {
	ChassisNr  *uint   // Chassis number.
	Chassis    *uint   // Chassis configuration value.
	Port       *uint   // Port configuration value.
	BusNr      *uint   // Bus number of the new bus.
	Index      *uint   // Order they will show up in the guest.
	NUMANode   *uint   // NUMA node reported to the guest OS.
	Hotplug    string  // Disables hotplug/unplug of devices ("on", "off").
	MemReserve *uint64 // Memory to reserve for PCI devices.
}

// DomainControllerPCIHole64 represents the `<pcihole64>` sub-element of `<pci>`, configuring the size of the 64-bit PCI hole.
type DomainControllerPCIHole64 struct {
	Size uint64           // Size of the 64-bit PCI hole.
	Unit string .
}

// DomainControllerUSB represents the `<usb>` sub-element of `<controller>`, configuring USB controller specific settings.
type DomainControllerUSB struct {
	Port   *uint                       // Number of devices that can be connected.
	Master *DomainControllerUSBMaster      // Master controller for companion controllers.
}

// DomainControllerUSBMaster represents the `<master>` sub-element of `<usb>`, configuring the master controller for companion controllers.
type DomainControllerUSBMaster struct {
	StartPort uint  // Starting port for companion controller.
}

// DomainControllerVirtIOSerial represents the `<virtio-serial>` sub-element of `<controller>`, configuring virtio-serial controller specific settings.
type DomainControllerVirtIOSerial struct {
	Ports   *uint    // Number of devices that can be connected.
	Vectors *uint  // Number of vectors.
}

// DomainControllerXenBus represents the `<xenbus>` sub-element of `<controller>`, configuring Xenbus controller specific settings.
type DomainControllerXenBus struct {
	MaxGrantFrames   uint    // Maximum number of grant frames.
	MaxEventChannels uint  // Maximum number of event channels.
}

// DomainControllerNVME represents the `<nvme>` sub-element of `<controller>`, configuring NVMe controller specific settings.
type DomainControllerNVME struct {
	Serial string  // Serial number of the NVMe controller.
}

// DomainLease represents the `<lease>` element, configuring device leases for a VM.
type DomainLease struct {
	Lockspace string             // Identifier for the lockspace.
	Key       string                   // Identifier for the lease.
	Target    *DomainLeaseTarget     // File associated with the lockspace.
}

// DomainLeaseTarget represents the `<target>` sub-element of `<lease>`, configuring the file associated with the lockspace.
type DomainLeaseTarget struct {
	Path   string           // Fully qualified path of the file.
	Offset uint64  // Offset where the lease is stored.
}

// DomainFilesystem represents a `<filesystem>` element, configuring a directory on the host accessible from the guest.
type DomainFilesystem struct {
	XMLName        xml.Name                        
	AccessMode     string                           // Security mode for accessing the source ("passthrough", "mapped", "squash").
	Model          string                                // Virtio device model.
	MultiDevs      string                            // How to deal with multiple devices in an export ("default", "remap", "forbid", "warn").
	FMode          string                                // Creation mode for files.
	DMode          string                                // Creation mode for directories.
	Driver         *DomainFilesystemDriver                             // Hypervisor driver details.
	Binary         *DomainFilesystemBinary                             // Options for virtiofsd.
	IDMap          *DomainFilesystemIDMap                               // Maps IDs in the user namespace.
	Source         *DomainFilesystemSource                             // Resource on the host.
	Target         *DomainFilesystemTarget                             // Where source can be accessed in guest.
	ReadOnly       *DomainFilesystemReadOnly                         // Exports filesystem as read-only.
	SpaceHardLimit *DomainFilesystemSpaceHardLimit           // Maximum space available.
	SpaceSoftLimit *DomainFilesystemSpaceSoftLimit           // Memory limit to enforce during memory contention.
	Boot           *DomainDeviceBoot                                     // Specifies that the filesystem is bootable.
	ACPI           *DomainDeviceACPI                                     // ACPI device configuration.
	Alias          *DomainAlias                                         // Identifier for the device.
	Address        *DomainAddress                                     // Device address on the virtual bus.
}

// DomainFilesystemDriver represents the `<driver>` sub-element of `<filesystem>`, specifying driver details for a filesystem.
type DomainFilesystemDriver struct {
	Type      string       // Backend driver name ("path", "handle", "virtiofs", "loop", "ploop", "mtp").
	Format    string     // Format type ("raw", "fat").
	Name      string       // Driver name.
	WRPolicy  string   // Write policy ("immediate").
	IOMMU     string      // Enables emulated IOMMU.
	ATS       string        // Address Translation Service support.
	Packed    string     // Whether to use packed virtqueues.
	PagePerVQ string  // Layout of notification capabilities.
	Queue     uint        // Queue size for virtiofs.
}

// DomainFilesystemBinary represents the `<binary>` sub-element of `<filesystem>`, configuring options for virtiofsd.
type DomainFilesystemBinary struct {
	Path       string                              // Path to the virtiofsd daemon.
	XAttr      string                             // Enables extended attributes ("on", "off").
	Cache      *DomainFilesystemBinaryCache                        // Caching settings.
	Sandbox    *DomainFilesystemBinarySandbox                    // Sandboxing method.
	Lock       *DomainFilesystemBinaryLock                          // Locking settings.
	ThreadPool *DomainFilesystemBinaryThreadPool             // Thread pool settings.
	OpenFiles  *DomainFilesystemBinaryOpenFiles                // Maximum number of file descriptors.
}

// DomainFilesystemBinaryCache represents the `<cache>` sub-element of `<binary>`, configuring caching settings for virtiofsd.
type DomainFilesystemBinaryCache struct {
	Mode string  // Cache mode ("none", "always").
}

// DomainFilesystemBinarySandbox represents the `<sandbox>` sub-element of `<binary>`, configuring sandboxing method for virtiofsd.
type DomainFilesystemBinarySandbox struct {
	Mode string  // Sandbox mode ("namespace", "chroot").
}

// DomainFilesystemBinaryLock represents the `<lock>` sub-element of `<binary>`, configuring locking settings for virtiofsd.
type DomainFilesystemBinaryLock struct {
	POSIX string  // POSIX locking ("on", "off").
	Flock string  // Flock locking ("on", "off").
}

// DomainFilesystemBinaryThreadPool represents the `<thread_pool>` sub-element of `<binary>`, configuring thread pool settings for virtiofsd.
type DomainFilesystemBinaryThreadPool struct {
	Size uint  // Maximum thread pool size.
}

// DomainFilesystemBinaryOpenFiles represents the `<openfiles>` sub-element of `<binary>`, configuring maximum number of file descriptors for virtiofsd.
type DomainFilesystemBinaryOpenFiles struct {
	Max uint  // Maximum number of file descriptors.
}

// DomainFilesystemIDMap represents the `<idmap>` sub-element of `<filesystem>`, mapping IDs in the user namespace.
type DomainFilesystemIDMap struct {
	UID []DomainFilesystemIDMapEntry  // User ID mapping entries.
	GID []DomainFilesystemIDMapEntry  // Group ID mapping entries.
}

// DomainFilesystemIDMapEntry represents a `<uid>` or `<gid>` sub-element of `<idmap>`, defining an ID mapping entry.
type DomainFilesystemIDMapEntry struct {
	Start  uint   // First ID in container.
	Target uint  // Target ID in host.
	Count  uint   // Number of IDs to map.
}

// DomainFilesystemSource represents the `<source>` sub-element of `<filesystem>`, specifying the resource on the host.
type DomainFilesystemSource struct {
	Mount    *DomainFilesystemSourceMount     // Host directory to mount.
	Block    *DomainFilesystemSourceBlock     // Host block device.
	File     *DomainFilesystemSourceFile      // Host file as an image.
	Template *DomainFilesystemSourceTemplate  // OpenVZ filesystem template.
	RAM      *DomainFilesystemSourceRAM       // In-memory filesystem.
	Bind     *DomainFilesystemSourceBind      // Directory binding.
	Volume   *DomainFilesystemSourceVolume    // Storage volume.
}

// DomainFilesystemSourceMount represents a `<mount>` sub-element of `<source>`, specifying a host directory to mount.
type DomainFilesystemSourceMount struct {
	Dir    string     // Host directory path.
	Socket string  // Unix socket path for virtiofs.
}

// DomainFilesystemSourceBlock represents a `<block>` sub-element of `<source>`, specifying a host block device.
type DomainFilesystemSourceBlock struct {
	Dev string  // Host block device path.
}

// DomainFilesystemSourceFile represents a `<file>` sub-element of `<source>`, specifying a host file as an image.
type DomainFilesystemSourceFile struct {
	File string  // Host file path.
}

// DomainFilesystemSourceTemplate represents a `<template>` sub-element of `<source>`, specifying an OpenVZ filesystem template.
type DomainFilesystemSourceTemplate struct {
	Name string  // Template name.
}

// DomainFilesystemSourceRAM represents a `<ram>` sub-element of `<source>`, specifying an in-memory filesystem.
type DomainFilesystemSourceRAM struct {
	Usage uint             // Memory usage limit.
	Units string .
}

// DomainFilesystemSourceBind represents a `<bind>` sub-element of `<source>`, specifying a directory binding.
type DomainFilesystemSourceBind struct {
	Dir string  // Directory to bind.
}

// DomainFilesystemSourceVolume represents a `<volume>` sub-element of `<source>`, specifying a storage volume.
type DomainFilesystemSourceVolume struct {
	Pool   string    // Storage pool name.
	Volume string  // Storage volume name.
}

// DomainFilesystemTarget represents the `<target>` sub-element of `<filesystem>`, specifying where the source can be accessed in the guest.
type DomainFilesystemTarget struct {
	Dir string  // Mount tag or guest directory.
}

// DomainFilesystemReadOnly represents the `<readonly>` sub-element of `<filesystem>`, exporting the filesystem as read-only.
type DomainFilesystemReadOnly struct {
}

// DomainFilesystemSpaceHardLimit represents the `<space_hard_limit>` sub-element of `<filesystem>`, specifying maximum space available.
type DomainFilesystemSpaceHardLimit struct {
	Value uint             // Maximum space available.
	Unit  string .
}

// DomainFilesystemSpaceSoftLimit represents the `<space_soft_limit>` sub-element of `<filesystem>`, specifying memory limit during contention.
type DomainFilesystemSpaceSoftLimit struct {
	Value uint             // Maximum space available during contention.
	Unit  string .
}

// DomainInterface represents an `<interface>` element, configuring network interfaces.
type DomainInterface struct {
	XMLName             xml.Name                           
	Managed             string                                          // Whether libvirt manages the PCI device ("yes", "no").
	TrustGuestRXFilters string                              // Host trusts guest RX filter reports ("yes", "no").
	MAC                 *DomainInterfaceMAC                                                // MAC address of the interface.
	Source              *DomainInterfaceSource                                          // Source of the network connection.
	Boot                *DomainDeviceBoot                                                 // Specifies that the interface is bootable.
	VLan                *DomainInterfaceVLan                                              // VLAN tagging configuration.
	VirtualPort         *DomainInterfaceVirtualPort                                // Configuration for virtual port.
	IP                  []DomainInterfaceIP                                                 // IP addresses for the network device in the guest.
	Route               []DomainInterfaceRoute                                           // IP routes to add in the guest.
	PortForward         []DomainInterfaceSourcePortForward                         // Forwards incoming traffic to guest.
	Script              *DomainInterfaceScript                                          // Path to shell script to run after creating/opening tap device.
	DownScript          *DomainInterfaceScript                                      // Path to shell script to run after detaching/closing tap device.
	BackendDomain       *DomainBackendDomain                                     // Backend domain for the interface.
	Target              *DomainInterfaceTarget                                          // Target device name.
	Guest               *DomainInterfaceGuest                                            // Guest-side device name.
	Model               *DomainInterfaceModel                                            // Model of emulated network interface card.
	Driver              *DomainInterfaceDriver                                          // Driver-specific options.
	Backend             *DomainInterfaceBackend                                        // Network backend-specific options.
	FilterRef           *DomainInterfaceFilterRef                                    // Reference to an nwfilter profile.
	Tune                *DomainInterfaceTune                                              // Network backend tuning.
	Teaming             *DomainInterfaceTeaming                                        // Connects two interfaces as a team/bond device.
	Link                *DomainInterfaceLink                                              // State of the virtual network link.
	MTU                 *DomainInterfaceMTU                                                // MTU of the virtual network link.
	Bandwidth           *DomainInterfaceBandwidth                                    // Quality of service settings.
	PortOptions         *DomainInterfacePortOptions                                       // Isolates network traffic.
	Coalesce            *DomainInterfaceCoalesce                                      // Coalesce settings.
	ROM                 *DomainROM                                                         // PCI Network device's ROM configuration.
	ACPI                *DomainDeviceACPI                                                 // ACPI device configuration.
	Alias               *DomainAlias                                                     // Identifier for the device.
	Address             *DomainAddress                                                 // Device address on the virtual bus.
}

// DomainInterfaceMAC represents the `<mac>` sub-element of `<interface>`, configuring the MAC address.
type DomainInterfaceMAC struct {
	Address string           // MAC address.
	Type    string    // MAC address type ("static").
	Check   string   // Whether to check MAC address.
}

// DomainInterfaceSource represents the `<source>` sub-element of `<interface>`, specifying the source of the network connection.
type DomainInterfaceSource struct {
	User      *DomainInterfaceSourceUser       // Userspace connection using SLIRP.
	Ethernet  *DomainInterfaceSourceEthernet   // Generic ethernet connection.
	VHostUser *DomainInterfaceSourceVHostUser  // vhost-user connection.
	Server    *DomainInterfaceSourceServer     // TCP server tunnel.
	Client    *DomainInterfaceSourceClient     // TCP client tunnel.
	MCast     *DomainInterfaceSourceMCast      // Multicast tunnel.
	Network   *DomainInterfaceSourceNetwork    // Virtual network connection.
	Bridge    *DomainInterfaceSourceBridge     // Bridge to LAN.
	Internal  *DomainInterfaceSourceInternal   // Internal network.
	Direct    *DomainInterfaceSourceDirect     // Direct attachment to physical interface.
	Hostdev   *DomainInterfaceSourceHostdev    // PCI Passthrough.
	UDP       *DomainInterfaceSourceUDP        // UDP unicast tunnel.
	VDPA      *DomainInterfaceSourceVDPA       // vDPA devices.
	Null      *DomainInterfaceSourceNull       // Unconnected network interface.
	VDS       *DomainInterfaceSourceVDS        // VMWare Distributed Switch.
}

// DomainInterfaceSourceUser represents the `<user>` sub-element of `<source>`, configuring a userspace connection using SLIRP.
type DomainInterfaceSourceUser struct {
	Dev string  // Device name.
}

// DomainInterfaceSourceEthernet represents the `<ethernet>` sub-element of `<source>`, configuring a generic ethernet connection.
type DomainInterfaceSourceEthernet struct {
	IP    []DomainInterfaceIP        // IP addresses for the host side.
	Route []DomainInterfaceRoute  // IP routes for the host side.
}

// DomainInterfaceSourceVHostUser embeds DomainChardevSource, representing a vhost-user connection.
type DomainInterfaceSourceVHostUser struct {
	Chardev *DomainChardevSource  // Character device for control plane.
	Dev     string                // Device name.
}

// DomainInterfaceSourceServer represents the `<server>` sub-element of `<source>`, configuring a TCP server tunnel.
type DomainInterfaceSourceServer struct {
	Address string                       // Server address.
	Port    uint                            // Server port.
	Local   *DomainInterfaceSourceLocal                   // Local address and port.
}

// DomainInterfaceSourceLocal represents the `<local>` sub-element of network sources, configuring local address and port.
type DomainInterfaceSourceLocal struct {
	Address string  // Local address.
	Port    uint       // Local port.
}

// DomainInterfaceSourceClient represents the `<client>` sub-element of `<source>`, configuring a TCP client tunnel.
type DomainInterfaceSourceClient struct {
	Address string                       // Server address.
	Port    uint                            // Server port.
	Local   *DomainInterfaceSourceLocal                   // Local address and port.
}

// DomainInterfaceSourceMCast represents the `<mcast>` sub-element of `<source>`, configuring a multicast tunnel.
type DomainInterfaceSourceMCast struct {
	Address string                       // Multicast address.
	Port    uint                            // Multicast port.
	Local   *DomainInterfaceSourceLocal                   // Local address and port.
}

// DomainInterfaceSourceNetwork represents the `<network>` sub-element of `<source>`, configuring a virtual network connection.
type DomainInterfaceSourceNetwork struct {
	Network   string    // Name of the virtual network.
	PortGroup string  // Name of the portgroup.
	Bridge    string     // Bridge name.
	PortID    string     // UUID of associated virNetworkPortPtr object.
}

// DomainInterfaceSourceBridge represents the `<bridge>` sub-element of `<source>`, configuring a bridge to LAN.
type DomainInterfaceSourceBridge struct {
	Bridge string  // Name of the host bridge device.
}

// DomainInterfaceSourceInternal represents the `<internal>` sub-element of `<source>`, configuring an internal network.
type DomainInterfaceSourceInternal struct {
	Name string  // Internal network name.
}

// DomainInterfaceSourceDirect represents the `<direct>` sub-element of `<source>`, configuring direct attachment to a physical interface.
type DomainInterfaceSourceDirect struct {
	Dev  string  // Physical interface name.
	Mode string  // Operation mode of macvtap device ("vepa", "bridge", "private", "passthrough").
}

// DomainInterfaceSourceHostdev represents the `<hostdev>` sub-element of `<source>`, configuring PCI Passthrough.
type DomainInterfaceSourceHostdev struct {
	PCI *DomainHostdevSubsysPCISource  // PCI host device source.
	USB *DomainHostdevSubsysUSBSource  // USB host device source.
}

// DomainInterfaceSourceUDP represents the `<udp>` sub-element of `<source>`, configuring a UDP unicast tunnel.
type DomainInterfaceSourceUDP struct {
	Address string                       // Endpoint address.
	Port    uint                            // Port number.
	Local   *DomainInterfaceSourceLocal                   // Local address and port.
}

// DomainInterfaceSourceVDPA represents the `<vdpa>` sub-element of `<source>`, configuring vDPA devices.
type DomainInterfaceSourceVDPA struct {
	Device string  // vDPA character device path.
}

// DomainInterfaceSourceNull represents the `<null>` sub-element of `<source>`, configuring an unconnected network interface.
type DomainInterfaceSourceNull struct {
}

// DomainInterfaceSourceVDS represents the `<vds>` sub-element of `<source>`, configuring a VMWare Distributed Switch connection.
type DomainInterfaceSourceVDS struct {
	SwitchID     string           // VMWare Distributed Switch ID.
	PortID       int      // Port ID.
	PortGroupID  string  // Port group ID.
	ConnectionID int     // Connection ID.
}

// DomainInterfaceIP represents an `<ip>` sub-element of `<interface>`, configuring IP addresses for the network device in the guest.
type DomainInterfaceIP struct {
	Address string           // IP address.
	Family  string  // Address family ("ipv4", "ipv6").
	Prefix  uint    // Number of 1 bits in the netmask.
	Peer    string    // IP address of the other end of a point-to-point device.
}

// DomainInterfaceRoute represents a `<route>` sub-element of `<interface>`, configuring IP routes to add in the guest.
type DomainInterfaceRoute struct {
	Family  string  // Address family.
	Address string           // Network address.
	Netmask string  // Network mask.
	Prefix  uint    // Prefix length.
	Gateway string           // Gateway address.
	Metric  uint    // Route metric.
}

// DomainInterfaceSourcePortForward represents a `<portForward>` sub-element of `<interface>`, forwarding incoming traffic to the guest.
type DomainInterfaceSourcePortForward struct {
	Proto   string                                     // Protocol ("tcp", "udp").
	Address string                                   // Original address.
	Dev     string                                       // Host interface to use.
	Ranges  []DomainInterfaceSourcePortForwardRange         // List of port ranges.
}

// DomainInterfaceSourcePortForwardRange represents a `<range>` sub-element of `<portForward>`, specifying a port range.
type DomainInterfaceSourcePortForwardRange struct {
	Start   uint             // Start port.
	End     uint     // End port.
	To      uint      // Port offset.
	Exclude string  // Exclude this range ("yes").
}

// DomainInterfaceScript represents a `<script>` or `<downscript>` sub-element of `<interface>`, specifying a shell script to run.
type DomainInterfaceScript struct {
	Path string  // Path to the shell script.
}

// DomainInterfaceTarget represents the `<target>` sub-element of `<interface>`, configuring the target device name.
type DomainInterfaceTarget struct {
	Dev     string               // Logical device name.
	Managed string  // Whether libvirt manages the device ("yes", "no").
}

// DomainInterfaceGuest represents the `<guest>` sub-element of `<interface>`, configuring the guest-side device name.
type DomainInterfaceGuest struct {
	Dev    string     // Name of the device on the guest side.
	Actual string  // Actual device name (output only).
}

// DomainInterfaceModel represents the `<model>` sub-element of `<interface>`, configuring the model of emulated network interface card.
type DomainInterfaceModel struct {
	Type string  // Model of emulated network interface card.
}

// DomainInterfaceDriver represents the `<driver>` sub-element of `<interface>`, specifying driver-specific options for an interface.
type DomainInterfaceDriver struct {
	Name          string                                 // Backend driver type ("qemu", "vhost", "vfio", "kvm").
	TXMode        string                               // How to handle transmission of packets ("iothread", "timer").
	IOEventFD     string                            // I/O asynchronous handling ("on", "off").
	EventIDX      string                            // Device event processing ("on", "off").
	Queues        uint                                 // Number of queues.
	RXQueueSize   uint                          // Size of virtio ring for RX.
	TXQueueSize   uint                          // Size of virtio ring for TX.
	IOMMU         string                                // Enables emulated IOMMU.
	ATS           string                                  // Address Translation Service support.
	Packed        string                               // Whether to use packed virtqueues.
	PagePerVQ     string                          // Layout of notification capabilities.
	RSS           string                                  // Enables in-qemu/ebpf RSS.
	RSSHashReport string                        // Enables in-qemu RSS hash report.
	Host          *DomainInterfaceDriverHost                            // Host offloading options.
	Guest         *DomainInterfaceDriverGuest                         // Guest offloading options.
}

// DomainInterfaceDriverHost represents the `<host>` sub-element of `<driver>`, configuring host offloading options.
type DomainInterfaceDriverHost struct {
	CSum     string      // Checksum offload.
	GSO      string       // Generic Segmentation Offload.
	TSO4     string      // TCP Segmentation Offload (IPv4).
	TSO6     string      // TCP Segmentation Offload (IPv6).
	ECN      string       // Explicit Congestion Notification.
	UFO      string       // UDP Fragmentation Offload.
	MrgRXBuf string  // Mergeable RX buffers.
}

// DomainInterfaceDriverGuest represents the `<guest>` sub-element of `<driver>`, configuring guest offloading options.
type DomainInterfaceDriverGuest struct {
	CSum string  // Checksum offload.
	TSO4 string  // TCP Segmentation Offload (IPv4).
	TSO6 string  // TCP Segmentation Offload (IPv6).
	ECN  string   // Explicit Congestion Notification.
	UFO  string   // UDP Fragmentation Offload.
}

// DomainInterfaceBackend represents the `<backend>` sub-element of `<interface>`, configuring network backend-specific options.
type DomainInterfaceBackend struct {
	Type    string     // Backend type ("passt").
	Tap     string      // Tun/tap device path.
	VHost   string    // Vhost device path.
	LogFile string  // Log file for passt process.
}

// DomainInterfaceFilterRef represents the `<filterref>` sub-element of `<interface>`, referencing an nwfilter profile.
type DomainInterfaceFilterRef struct {
	Filter     string                            // Name of the nwfilter to use.
	Parameters []DomainInterfaceFilterParam  // Parameters for the nwfilter.
}

// DomainInterfaceFilterParam represents a `<parameter>` sub-element of `<filterref>`, configuring a parameter for the nwfilter.
type DomainInterfaceFilterParam struct {
	Name  string   // Parameter name.
	Value string  // Parameter value.
}

// DomainInterfaceTune represents the `<tune>` sub-element of `<interface>`, configuring network backend tuning.
type DomainInterfaceTune struct {
	SndBuf uint  // Size of send buffer in host.
}

// DomainInterfaceTeaming represents the `<teaming>` sub-element of `<interface>`, connecting two interfaces as a team/bond device.
type DomainInterfaceTeaming struct {
	Type       string           // Teaming type ("persistent", "transient").
	Persistent string  // Alias name of the persistent device.
}

// DomainInterfaceLink represents the `<link>` sub-element of `<interface>`, configuring the state of the virtual network link.
type DomainInterfaceLink struct {
	State string  // State of the virtual network link ("up", "down").
}

// DomainInterfaceMTU represents the `<mtu>` sub-element of `<interface>`, configuring the MTU of the virtual network link.
type DomainInterfaceMTU struct {
	Size uint  // MTU size.
}

// DomainInterfaceBandwidth represents the `<bandwidth>` sub-element of `<interface>`, configuring quality of service settings.
type DomainInterfaceBandwidth struct {
	Inbound  *DomainInterfaceBandwidthParams   // Inbound traffic shaping.
	Outbound *DomainInterfaceBandwidthParams  // Outbound traffic shaping.
}

// DomainInterfaceBandwidthParams represents inbound/outbound traffic shaping parameters.
type DomainInterfaceBandwidthParams struct {
	Average *int  // Desired average bit rate.
	Peak    *int     // Maximum rate.
	Burst   *int    // Amount of kibibytes transmitted in a single burst.
	Floor   *int    // Guaranteed minimal throughput.
}

// DomainInterfacePortOptions represents the `<port>` sub-element of `<interface>`, configuring network traffic isolation.
type DomainInterfacePortOptions struct {
	Isolated string  // Isolates network traffic ("yes", "no").
}

// DomainInterfaceCoalesce represents the `<coalesce>` sub-element of `<interface>`, configuring coalesce settings.
type DomainInterfaceCoalesce struct {
	RX *DomainInterfaceCoalesceRX  // Receive coalesce settings.
}

// DomainInterfaceCoalesceRX represents the `<rx>` sub-element of `<coalesce>`, configuring receive coalesce settings.
type DomainInterfaceCoalesceRX struct {
	Frames *DomainInterfaceCoalesceRXFrames  // Frames coalesce settings.
}

// DomainInterfaceCoalesceRXFrames represents the `<frames>` sub-element of `<rx>`, configuring maximum number of packets received before interrupt.
type DomainInterfaceCoalesceRXFrames struct {
	Max *uint  // Maximum number of packets received before interrupt.
}

// DomainROM represents the `<rom>` sub-element of `<interface>`, configuring PCI Network device's ROM.
type DomainROM struct {
	Bar     string       // Whether ROM will be visible in guest's memory map ("on", "off").
	File    *string               // Path to binary file for ROM BIOS.
	Enabled string   // Disables PCI ROM loading ("yes", "no").
}

// DomainInterfaceVLan represents the `<vlan>` sub-element of `<interface>`, configuring VLAN tagging.
type DomainInterfaceVLan struct {
	Trunk string                    // VLAN trunking ("yes").
	Tags  []DomainInterfaceVLanTag                   // List of VLAN tags.
}

// DomainInterfaceVLanTag represents a `<tag>` sub-element of `<vlan>`, configuring a specific VLAN tag.
type DomainInterfaceVLanTag struct {
	ID         uint             // VLAN ID.
	NativeMode string  // Native VLAN mode ("tagged", "untagged").
}

// DomainInterfaceVirtualPort represents the `<virtualport>` sub-element of `<interface>`, configuring virtual port parameters.
type DomainInterfaceVirtualPort struct {
	Params *DomainInterfaceVirtualPortParams  // Virtual port parameters.
}

// DomainInterfaceVirtualPortParams represents the `<parameters>` sub-element of `<virtualport>`, configuring virtual port parameters.
type DomainInterfaceVirtualPortParams struct {
	Any          *DomainInterfaceVirtualPortParamsAny           // Generic virtual port parameters.
	VEPA8021QBG  *DomainInterfaceVirtualPortParamsVEPA8021QBG   // 802.1Qbg parameters.
	VNTag8011QBH *DomainInterfaceVirtualPortParamsVNTag8021QBH  // 802.1Qbh parameters.
	OpenVSwitch  *DomainInterfaceVirtualPortParamsOpenVSwitch   // Open vSwitch parameters.
	MidoNet      *DomainInterfaceVirtualPortParamsMidoNet       // Midonet parameters.
}

// DomainInterfaceVirtualPortParamsAny represents generic virtual port parameters.
type DomainInterfaceVirtualPortParamsAny struct {
	ManagerID     *uint            // VSI Manager ID.
	TypeID        *uint               // VSI Type ID.
	TypeIDVersion *uint        // VSI Type Version.
	InstanceID    string  // VSI Instance ID Identifier.
	ProfileID     string   // Port profile ID.
	InterfaceID   string  // Interface ID.
}

// DomainInterfaceVirtualPortParamsVEPA8021QBG represents 802.1Qbg virtual port parameters.
type DomainInterfaceVirtualPortParamsVEPA8021QBG struct {
	ManagerID     *uint            // VSI Manager ID.
	TypeID        *uint               // VSI Type ID.
	TypeIDVersion *uint        // VSI Type Version.
	InstanceID    string  // VSI Instance ID Identifier.
}

// DomainInterfaceVirtualPortParamsVNTag8021QBH represents 802.1Qbh virtual port parameters.
type DomainInterfaceVirtualPortParamsVNTag8021QBH struct {
	ProfileID string  // Port profile ID.
}

// DomainInterfaceVirtualPortParamsOpenVSwitch represents Open vSwitch virtual port parameters.
type DomainInterfaceVirtualPortParamsOpenVSwitch struct {
	InterfaceID string  // Interface ID.
	ProfileID   string    // Port profile ID.
}

// DomainInterfaceVirtualPortParamsMidoNet represents Midonet virtual port parameters.
type DomainInterfaceVirtualPortParamsMidoNet struct {
	InterfaceID string  // Interface ID.
}

// DomainSmartcard represents a `<smartcard>` element, configuring a virtual smartcard device.
type DomainSmartcard struct {
	XMLName     xml.Name                
	Passthrough *DomainChardevSource                 // Character device for tunneling requests.
	Protocol    *DomainChardevProtocol             // Protocol for the character device.
	Host        *DomainSmartcardHost                      // Host smartcard access.
	HostCerts   []DomainSmartcardHostCert         // NSS certificate names.
	Database    string                   // Path to alternate NSS database directory.
	ACPI        *DomainDeviceACPI                      // ACPI device configuration.
	Alias       *DomainAlias                          // Identifier for the device.
	Address     *DomainAddress                      // Device address on the virtual bus.
}

// DomainSmartcardHost represents the `<host>` sub-element of `<smartcard>`, configuring host smartcard access.
type DomainSmartcardHost struct {
}

// DomainSmartcardHostCert represents a `<certificate>` sub-element of `<smartcard>`, specifying an NSS certificate name.
type DomainSmartcardHostCert struct {
	File string  // NSS certificate name.
}

// DomainSerial represents a `<serial>` element, configuring serial devices.
type DomainSerial struct {
	XMLName  xml.Name               
	Source   *DomainChardevSource      // Host interface for the character device.
	Protocol *DomainChardevProtocol  // Protocol for the character device.
	Target   *DomainSerialTarget       // Guest interface for the character device.
	Log      *DomainChardevLog            // Log file for the character device.
	ACPI     *DomainDeviceACPI           // ACPI device configuration.
	Alias    *DomainAlias               // Identifier for the device.
	Address  *DomainAddress           // Device address on the virtual bus.
}

// DomainSerialTarget represents the `<target>` sub-element of `<serial>`, configuring the guest interface for a serial device.
type DomainSerialTarget struct {
	Type  string                    // Target type.
	Port  *uint                               // Port number.
	Model *DomainSerialTargetModel                // Target model.
}

// DomainSerialTargetModel represents the `<model>` sub-element of `<target>`, specifying the target model.
type DomainSerialTargetModel struct {
	Name string  // Model name.
}

// DomainParallel represents a `<parallel>` element, configuring parallel devices.
type DomainParallel struct {
	XMLName  xml.Name               
	Source   *DomainChardevSource      // Host interface for the character device.
	Protocol *DomainChardevProtocol  // Protocol for the character device.
	Target   *DomainParallelTarget     // Guest interface for the character device.
	Log      *DomainChardevLog            // Log file for the character device.
	ACPI     *DomainDeviceACPI           // ACPI device configuration.
	Alias    *DomainAlias               // Identifier for the device.
	Address  *DomainAddress           // Device address on the virtual bus.
}

// DomainParallelTarget represents the `<target>` sub-element of `<parallel>`, configuring the guest interface for a parallel device.
type DomainParallelTarget struct {
	Type string  // Target type.
	Port *uint             // Port number.
}

// DomainConsole represents a `<console>` element, configuring interactive serial consoles.
type DomainConsole struct {
	XMLName  xml.Name               
	TTY      string                  // TTY path (for compatibility).
	Source   *DomainChardevSource                // Host interface for the character device.
	Protocol *DomainChardevProtocol            // Protocol for the character device.
	Target   *DomainConsoleTarget                // Guest interface for the character device.
	Log      *DomainChardevLog                      // Log file for the character device.
	ACPI     *DomainDeviceACPI                     // ACPI device configuration.
	Alias    *DomainAlias                         // Identifier for the device.
	Address  *DomainAddress                     // Device address on the virtual bus.
}

// DomainConsoleTarget represents the `<target>` sub-element of `<console>`, configuring the guest interface for a console device.
type DomainConsoleTarget struct {
	Type string  // Target type.
	Port *uint             // Port number.
}

// DomainChannel represents a `<channel>` element, configuring private communication channels between host and guest.
type DomainChannel struct {
	XMLName  xml.Name               
	Source   *DomainChardevSource      // Host interface for the character device.
	Protocol *DomainChardevProtocol  // Protocol for the character device.
	Target   *DomainChannelTarget      // Guest interface for the character device.
	Log      *DomainChardevLog            // Log file for the character device.
	ACPI     *DomainDeviceACPI           // ACPI device configuration.
	Alias    *DomainAlias               // Identifier for the device.
	Address  *DomainAddress           // Device address on the virtual bus.
}

// DomainChannelTarget represents the `<target>` sub-element of `<channel>`, configuring the guest interface for a channel device.
type DomainChannelTarget struct {
	VirtIO   *DomainChannelTargetVirtIO    // Paravirtualized virtio channel.
	Xen      *DomainChannelTargetXen       // Paravirtualized Xen channel.
	GuestFWD *DomainChannelTargetGuestFWD  // TCP traffic forwarded to channel device.
}

// DomainChannelTargetVirtIO represents a virtio channel target.
type DomainChannelTargetVirtIO struct {
	Name  string   // Name for the virtio channel.
	State string  // Whether guest agent is connected ("connected", "disconnected").
}

// DomainChannelTargetXen represents a Xen channel target.
type DomainChannelTargetXen struct {
	Name  string   // Name for the Xen channel.
	State string  // Whether guest agent is connected.
}

// DomainChannelTargetGuestFWD represents a guestfwd channel target.
type DomainChannelTargetGuestFWD struct {
	Address string  // IP address for forwarding.
	Port    string     // Port for forwarding.
}

// DomainChardevLog represents the `<log>` sub-element of character devices, configuring a log file.
type DomainChardevLog struct {
	File   string           // Path to the log file.
	Append string  // Whether to append to the file ("on", "off").
}

// DomainChardevProtocol represents the `<protocol>` sub-element of character devices, configuring the protocol.
type DomainChardevProtocol struct {
	Type string  // Protocol type ("raw", "telnet", "tls").
}

// DomainChardevSource represents the `<source>` sub-element of character devices, specifying the host interface.
type DomainChardevSource struct {
	Null        *DomainChardevSourceNull         // Null device.
	VC          *DomainChardevSourceVC           // Virtual console.
	Pty         *DomainChardevSourcePty          // Pseudo TTY.
	Dev         *DomainChardevSourceDev          // Host device proxy.
	File        *DomainChardevSourceFile         // Device logfile.
	Pipe        *DomainChardevSourcePipe         // Named pipe.
	StdIO       *DomainChardevSourceStdIO        // Domain logfile.
	UDP         *DomainChardevSourceUDP          // UDP network console.
	TCP         *DomainChardevSourceTCP          // TCP client/server.
	UNIX        *DomainChardevSourceUNIX         // UNIX domain socket client/server.
	SpiceVMC    *DomainChardevSourceSpiceVMC     // Spice VMC channel.
	SpicePort   *DomainChardevSourceSpicePort    // Spice channel.
	NMDM        *DomainChardevSourceNMDM         // Nmdm device.
	QEMUVDAgent *DomainChardevSourceQEMUVDAgent  // QEMU vdagent channel.
	DBus        *DomainChardevSourceDBus         // D-Bus channel.
}

// DomainChardevSourceNull represents a null device source.
type DomainChardevSourceNull struct {
}

// DomainChardevSourceVC represents a virtual console source.
type DomainChardevSourceVC struct {
}

// DomainChardevSourcePty represents a pseudo TTY source.
type DomainChardevSourcePty struct {
	Path     string                  // Path to the pseudo TTY.
	SecLabel []DomainDeviceSecLabel   // Security label override.
}

// DomainChardevSourceDev represents a host device proxy source.
type DomainChardevSourceDev struct {
	Path     string                  // Path to the physical character device.
	SecLabel []DomainDeviceSecLabel   // Security label override.
}

// DomainChardevSourceFile represents a device logfile source.
type DomainChardevSourceFile struct {
	Path     string                           // Path to the log file.
	Append   string                  // Whether to append to the file ("on", "off").
	SecLabel []DomainDeviceSecLabel             // Security label override.
}

// DomainChardevSourcePipe represents a named pipe source.
type DomainChardevSourcePipe struct {
	Path     string                  // Path to the named pipe.
	SecLabel []DomainDeviceSecLabel   // Security label override.
}

// DomainChardevSourceStdIO represents a domain logfile source.
type DomainChardevSourceStdIO struct {
}

// DomainChardevSourceUDP represents a UDP network console source.
type DomainChardevSourceUDP struct {
	BindHost       string  // Host to bind to.
	BindService    string  // Service to bind to.
	ConnectHost    string  // Host to connect to.
	ConnectService string  // Service to connect to.
}

// DomainChardevSourceTCP represents a TCP client/server source.
type DomainChardevSourceTCP struct {
	Mode      string                       // Connection mode ("connect", "bind").
	Host      string                       // Hostname or IP address.
	Service   string                       // Service name or port number.
	TLS       string                       // Whether to use TLS ("yes", "no").
	Reconnect *DomainChardevSourceReconnect          // Reconnect timeout.
}

// DomainChardevSourceReconnect configures reconnect timeout for TCP/UNIX character devices.
type DomainChardevSourceReconnect struct {
	Enabled string  // Whether reconnect is enabled ("yes", "no").
	Timeout *uint   // Timeout in seconds.
}

// DomainChardevSourceUNIX represents a UNIX domain socket client/server source.
type DomainChardevSourceUNIX struct {
	Mode      string                       // Connection mode ("bind", "connect").
	Path      string                       // Path to the UNIX domain socket.
	Reconnect *DomainChardevSourceReconnect          // Reconnect timeout.
	SecLabel  []DomainDeviceSecLabel                // Security label override.
}

// DomainChardevSourceSpiceVMC represents a Spice VMC channel source.
type DomainChardevSourceSpiceVMC struct {
}

// DomainChardevSourceSpicePort represents a Spice channel source.
type DomainChardevSourceSpicePort struct {
	Channel string  // Spice channel name.
}

// DomainChardevSourceNMDM represents an Nmdm device source.
type DomainChardevSourceNMDM struct {
	Master string  // Path to the master device.
	Slave  string   // Path to the slave device.
}

// DomainChardevSourceQEMUVDAgent represents a QEMU vdagent channel source.
type DomainChardevSourceQEMUVDAgent struct {
	Mouse     *DomainChardevSourceQEMUVDAgentMouse          // Mouse mode.
	ClipBoard *DomainChardevSourceQEMUVDAgentClipBoard  // Copy & Paste functionality.
}

// DomainChardevSourceQEMUVDAgentMouse configures mouse mode for QEMU vdagent.
type DomainChardevSourceQEMUVDAgentMouse struct {
	Mode string  // Mouse mode ("server", "client").
}

// DomainChardevSourceQEMUVDAgentClipBoard configures Copy & Paste functionality for QEMU vdagent.
type DomainChardevSourceQEMUVDAgentClipBoard struct {
	CopyPaste string  // Copy & Paste functionality ("yes", "no").
}

// DomainChardevSourceDBus represents a D-Bus channel source.
type DomainChardevSourceDBus struct {
	Channel string  // D-Bus channel name.
}

// DomainInput represents an `<input>` element, configuring input devices.
type DomainInput struct {
	XMLName xml.Name           
	Type    string                         // Input device type ("mouse", "tablet", "keyboard", "passthrough", "evdev").
	Bus     string                // Bus type ("xen", "ps2", "usb", "virtio").
	Model   string              // Virtio device model.
	Driver  *DomainInputDriver                // Driver specific options.
	Source  *DomainInputSource                // Source for passthrough/evdev.
	ACPI    *DomainDeviceACPI                   // ACPI device configuration.
	Alias   *DomainAlias                       // Identifier for the device.
	Address *DomainAddress                   // Device address on the virtual bus.
}

// DomainInputDriver represents the `<driver>` sub-element of `<input>`, specifying driver-specific options for an input device.
type DomainInputDriver struct {
	IOMMU     string      // Enables emulated IOMMU.
	ATS       string        // Address Translation Service support.
	Packed    string     // Whether to use packed virtqueues.
	PagePerVQ string  // Layout of notification capabilities.
}

// DomainInputSource represents the `<source>` sub-element of `<input>`, specifying the source for passthrough/evdev.
type DomainInputSource struct {
	Passthrough *DomainInputSourcePassthrough  // Passthrough event device.
	EVDev       *DomainInputSourceEVDev        // Event device.
}

// DomainInputSourcePassthrough represents a passthrough event device source.
type DomainInputSourcePassthrough struct {
	EVDev string  // Path to the event device.
}

// DomainInputSourceEVDev represents an event device source.
type DomainInputSourceEVDev struct {
	Dev        string               // Path to the event device.
	Grab       string    // Grabs all input devices ("all").
	GrabToggle string  // Grab key combination.
	Repeat     string  // Enables/disables auto-repeat events ("on", "off").
}

// DomainTPM represents a `<tpm>` element, configuring a TPM device.
type DomainTPM struct {
	XMLName xml.Name        
	Model   string           // Device model ("tpm-tis", "tpm-crb", "tpm-spapr", "spapr-tpm-proxy").
	Backend *DomainTPMBackend             // Type of TPM device.
	ACPI    *DomainDeviceACPI                // ACPI device configuration.
	Alias   *DomainAlias                  // Identifier for the device.
	Address *DomainAddress              // Device address on the virtual bus.
}

// DomainTPMBackend represents the `<backend>` sub-element of `<tpm>`, specifying the type of TPM device.
type DomainTPMBackend struct {
	Passthrough *DomainTPMBackendPassthrough  // Passthrough to host TPM.
	Emulator    *DomainTPMBackendEmulator     // TPM emulator.
	External    *DomainTPMBackendExternal     // External TPM emulator.
}

// DomainTPMBackendPassthrough represents a passthrough TPM backend.
type DomainTPMBackendPassthrough struct {
	Device *DomainTPMBackendDevice  // Host TPM device.
}

// DomainTPMBackendDevice represents the `<device>` sub-element of `<passthrough>`, specifying the path to the TPM device.
type DomainTPMBackendDevice struct {
	Path string  // Path to the TPM device.
}

// DomainTPMBackendEmulator represents a TPM emulator backend.
type DomainTPMBackendEmulator struct {
	Version         string                    // TPM version ("1.2", "2.0").
	Encryption      *DomainTPMBackendEncryption              // Encrypts TPM emulator state.
	PersistentState string                    // Whether TPM state is kept ("yes", "no").
	Debug           uint                        // Enables logging in emulator backend.
	ActivePCRBanks  *DomainTPMBackendPCRBanks        // PCR banks to activate.
	Source          *DomainTPMBackendSource                   // Location of TPM state storage.
	Profile         *DomainTPMBackendProfile                 // TPM profile settings.
}

// DomainTPMBackendEncryption represents the `<encryption>` sub-element of `<emulator>`, encrypting TPM emulator state.
type DomainTPMBackendEncryption struct {
	Secret string  // Reference to a secret object.
}

// DomainTPMBackendPCRBanks represents the `<active_pcr_banks>` sub-element of `<emulator>`, configuring PCR banks to activate.
type DomainTPMBackendPCRBanks struct {
	SHA1   *DomainTPMBackendPCRBank    // SHA1 PCR bank.
	SHA256 *DomainTPMBackendPCRBank  // SHA256 PCR bank.
	SHA384 *DomainTPMBackendPCRBank  // SHA384 PCR bank.
	SHA512 *DomainTPMBackendPCRBank  // SHA512 PCR bank.
}

// DomainTPMBackendPCRBank represents an individual PCR bank.
type DomainTPMBackendPCRBank struct {
}

// DomainTPMBackendSource represents the `<source>` sub-element of `<emulator>`, specifying the location of TPM state storage.
type DomainTPMBackendSource struct {
	File *DomainTPMBackendSourceFile  // File for TPM state storage.
	Dir  *DomainTPMBackendSourceDir   // Directory for TPM state storage.
}

// DomainTPMBackendSourceFile represents a file for TPM state storage.
type DomainTPMBackendSourceFile struct {
	Path string  // Path to the file.
}

// DomainTPMBackendSourceDir represents a directory for TPM state storage.
type DomainTPMBackendSourceDir struct {
	Path string  // Path to the directory.
}

// DomainTPMBackendProfile represents the `<profile>` sub-element of `<emulator>`, configuring TPM profile settings.
type DomainTPMBackendProfile struct {
	Source         string          // Name of the file where profile is stored.
	RemoveDisabled string  // Removes disabled algorithms ("check", "fips-host").
	Name           string            // Profile name.
}

// DomainTPMBackendExternal represents an external TPM emulator backend.
type DomainTPMBackendExternal struct {
	Source *DomainTPMBackendExternalSource  // Socket of the externally started TPM emulator.
}

// DomainTPMBackendExternalSource embeds DomainChardevSource, representing the source for an external TPM emulator.
type DomainTPMBackendExternalSource DomainChardevSource

// DomainGraphic represents a `<graphics>` element, configuring graphical framebuffers.
type DomainGraphic struct {
	XMLName     xml.Name                  
	SDL         *DomainGraphicSDL          // SDL graphics.
	VNC         *DomainGraphicVNC          // VNC server.
	RDP         *DomainGraphicRDP          // RDP server.
	Desktop     *DomainGraphicDesktop      // VirtualBox desktop.
	Spice       *DomainGraphicSpice        // SPICE server.
	EGLHeadless *DomainGraphicEGLHeadless  // OpenGL accelerated display.
	DBus        *DomainGraphicDBus         // D-Bus display.
	Audio       *DomainGraphicAudio        // Host audio backend mapping.
}

// DomainGraphicSDL represents the `<sdl>` sub-element of `<graphics>`, configuring SDL graphics.
type DomainGraphicSDL struct {
	Display    string                   // Display to use.
	XAuth      string                     // Authentication identifier.
	FullScreen string                // Fullscreen mode ("yes", "no").
	GL         *DomainGraphicsSDLGL                         // OpenGL support.
}

// DomainGraphicsSDLGL represents the `<gl>` sub-element of `<sdl>`, configuring OpenGL support.
type DomainGraphicsSDLGL struct {
	Enable string  // Enables OpenGL support ("yes", "no").
}

// DomainGraphicVNC represents the `<vnc>` sub-element of `<graphics>`, configuring a VNC server.
type DomainGraphicVNC struct {
	Socket        string                       // Unix domain socket path.
	Port          int                            // TCP port number.
	AutoPort      string                     // Auto-allocation of TCP port ("yes").
	WebSocket     int                       // Port to listen on for WebSocket.
	Keymap        string                       // Keymap to use.
	SharePolicy   string                   // Display sharing policy.
	Passwd        string                       // VNC password.
	PasswdValidTo string                   // Password validity timestamp.
	Connected     string                    // Control connected client during password changes.
	PowerControl  string                   // Enables VM power control features ("yes", "no").
	Listen        string                       // Listen address.
	Listeners     []DomainGraphicListener                     // List of listen configurations.
}

// DomainGraphicListener represents a `<listen>` sub-element of graphics devices, configuring listen settings.
type DomainGraphicListener struct {
	Address *DomainGraphicListenerAddress  // Listen on a specific IP address/hostname.
	Network *DomainGraphicListenerNetwork  // Listen on an existing network.
	Socket  *DomainGraphicListenerSocket   // Listen on a Unix socket.
}

// DomainGraphicListenerAddress represents an `<address>` sub-element of `<listen>`, configuring listen on a specific IP address/hostname.
type DomainGraphicListenerAddress struct {
	Address string  // IP address or hostname.
}

// DomainGraphicListenerNetwork represents a `<network>` sub-element of `<listen>`, configuring listen on an existing network.
type DomainGraphicListenerNetwork struct {
	Address string  // IP address.
	Network string  // Name of the network.
}

// DomainGraphicListenerSocket represents a `<socket>` sub-element of `<listen>`, configuring listen on a Unix socket.
type DomainGraphicListenerSocket struct {
	Socket string  // Path to Unix socket.
}

// DomainGraphicRDP represents the `<rdp>` sub-element of `<graphics>`, configuring an RDP server.
type DomainGraphicRDP struct {
	Port        int                            // TCP port number.
	AutoPort    string                     // Auto-allocation of TCP port ("yes").
	ReplaceUser string                   // Whether to replace existing connection ("yes", "no").
	MultiUser   string                    // Multiple simultaneous connections ("yes", "no").
	Username    string                     // RDP username.
	Passwd      string                       // RDP password.
	Listen      string                       // Listen address.
	Listeners   []DomainGraphicListener                     // List of listen configurations.
}

// DomainGraphicDesktop represents the `<desktop>` sub-element of `<graphics>`, configuring a VirtualBox desktop.
type DomainGraphicDesktop struct {
	Display    string     // Display to use.
	FullScreen string  // Fullscreen mode ("yes", "no").
}

// DomainGraphicSpice represents the `<spice>` sub-element of `<graphics>`, configuring a SPICE server.
type DomainGraphicSpice struct {
	Port          int                                  // TCP port number.
	TLSPort       int                               // Secure port number.
	AutoPort      string                           // Auto-allocation of port numbers ("yes").
	Listen        string                             // Listen address.
	Keymap        string                             // Keymap to use.
	DefaultMode   string                         // Default channel security policy.
	Passwd        string                             // SPICE password.
	PasswdValidTo string                         // Password validity timestamp.
	Connected     string                          // Control connected client during password changes.
	Listeners     []DomainGraphicListener                           // List of listen configurations.
	Channel       []DomainGraphicSpiceChannel                      // Restricts what channels can be run on each port.
	Image         *DomainGraphicSpiceImage                           // Image compression settings.
	JPEG          *DomainGraphicSpiceJPEG                             // JPEG compression settings.
	ZLib          *DomainGraphicSpiceZLib                             // ZLib compression settings.
	Playback      *DomainGraphicSpicePlayback                     // Audio stream compression.
	Streaming     *DomainGraphicSpiceStreaming                   // Streaming mode.
	Mouse         *DomainGraphicSpiceMouse                           // Mouse mode.
	ClipBoard     *DomainGraphicSpiceClipBoard                   // Copy & Paste functionality.
	FileTransfer  *DomainGraphicSpiceFileTransfer               // File transfer functionality.
	GL            *DomainGraphicSpiceGL                                 // OpenGL accelerated server-side rendering.
}

// DomainGraphicSpiceChannel represents a `<channel>` sub-element of `<spice>`, restricting channels.
type DomainGraphicSpiceChannel struct {
	Name string  // Channel name.
	Mode string  // Channel security policy ("secure", "insecure", "any").
}

// DomainGraphicSpiceImage represents the `<image>` sub-element of `<spice>`, configuring image compression.
type DomainGraphicSpiceImage struct {
	Compression string  // Image compression ("auto_glz", "auto_lz", "quic", "glz", "lz", "off").
}

// DomainGraphicSpiceJPEG represents the `<jpeg>` sub-element of `<spice>`, configuring JPEG compression.
type DomainGraphicSpiceJPEG struct {
	Compression string  // JPEG compression ("auto", "never", "always").
}

// DomainGraphicSpiceZLib represents the `<zlib>` sub-element of `<spice>`, configuring ZLib compression.
type DomainGraphicSpiceZLib struct {
	Compression string  // ZLib compression ("auto", "never", "always").
}

// DomainGraphicSpicePlayback represents the `<playback>` sub-element of `<spice>`, configuring audio stream compression.
type DomainGraphicSpicePlayback struct {
	Compression string  // Audio stream compression ("on", "off").
}

// DomainGraphicSpiceStreaming represents the `<streaming>` sub-element of `<spice>`, configuring streaming mode.
type DomainGraphicSpiceStreaming struct {
	Mode string  // Streaming mode ("filter", "all", "off").
}

// DomainGraphicSpiceMouse represents the `<mouse>` sub-element of `<spice>`, configuring mouse mode.
type DomainGraphicSpiceMouse struct {
	Mode string  // Mouse mode ("server", "client").
}

// DomainGraphicSpiceClipBoard represents the `<clipboard>` sub-element of `<spice>`, configuring Copy & Paste functionality.
type DomainGraphicSpiceClipBoard struct {
	CopyPaste string  // Copy & Paste functionality ("yes", "no").
}

// DomainGraphicSpiceFileTransfer represents the `<filetransfer>` sub-element of `<spice>`, configuring file transfer functionality.
type DomainGraphicSpiceFileTransfer struct {
	Enable string  // File transfer functionality ("yes", "no").
}

// DomainGraphicSpiceGL represents the `<gl>` sub-element of `<spice>`, configuring OpenGL accelerated server-side rendering.
type DomainGraphicSpiceGL struct {
	Enable     string  // Enables OpenGL support ("yes", "no").
	RenderNode string  // Path to DRM render node.
}

// DomainGraphicEGLHeadless represents the `<egl-headless>` sub-element of `<graphics>`, configuring an OpenGL accelerated display.
type DomainGraphicEGLHeadless struct {
	GL *DomainGraphicEGLHeadlessGL  // OpenGL rendering settings.
}

// DomainGraphicEGLHeadlessGL represents the `<gl>` sub-element of `<egl-headless>`, configuring OpenGL rendering settings.
type DomainGraphicEGLHeadlessGL struct {
	RenderNode string  // Path to host's DRI device.
}

// DomainGraphicDBus represents the `<dbus>` sub-element of `<graphics>`, configuring a D-Bus display.
type DomainGraphicDBus struct {
	Address string               // D-Bus address.
	P2P     string                   // Enables peer-to-peer connections ("yes", "no").
	GL      *DomainGraphicDBusGL                      // OpenGL rendering settings.
}

// DomainGraphicDBusGL represents the `<gl>` sub-element of `<dbus>`, configuring OpenGL rendering settings.
type DomainGraphicDBusGL struct {
	Enable     string  // Enables OpenGL support ("yes", "no").
	RenderNode string  // Path to DRM render node.
}

// DomainGraphicAudio represents the `<audio>` sub-element of `<graphics>`, configuring host audio backend mapping.
type DomainGraphicAudio struct {
	ID uint  // ID of the audio device.
}

// DomainSound represents a `<sound>` element, configuring virtual sound cards.
type DomainSound struct {
	XMLName      xml.Name           
	Model        string                         // Emulated sound device model.
	MultiChannel string              // Multi-channel mode for USB sound device ("yes", "no").
	Streams      uint                // Number of PCM streams for virtio sound device.
	Codec        []DomainSoundCodec                  // Audio codecs.
	Audio        *DomainSoundAudio                   // Host audio backend mapping.
	ACPI         *DomainDeviceACPI                    // ACPI device configuration.
	Alias        *DomainAlias                        // Identifier for the device.
	Driver       *DomainSoundDriver                 // Driver specific options.
	Address      *DomainAddress                    // Device address on the virtual bus.
}

// DomainSoundCodec represents a `<codec>` sub-element of `<sound>`, configuring audio codecs.
type DomainSoundCodec struct {
	Type string  // Codec type ("duplex", "micro", "output").
}

// DomainSoundAudio represents the `<audio>` sub-element of `<sound>`, configuring host audio backend mapping.
type DomainSoundAudio struct {
	ID uint  // ID of the audio device.
}

// DomainSoundDriver represents the `<driver>` sub-element of `<sound>`, specifying driver-specific options for a sound device.
type DomainSoundDriver struct {
	IOMMU     string      // Enables emulated IOMMU.
	ATS       string        // Address Translation Service support.
	Packed    string     // Whether to use packed virtqueues.
	PagePerVQ string  // Layout of notification capabilities.
}

// DomainAudio represents an `<audio>` element, configuring virtual audio devices.
type DomainAudio struct {
	XMLName     xml.Name                  
	ID          int                                   // Integer ID of the audio device.
	TimerPeriod uint                       // Timer period in microseconds.
	None        *DomainAudioNone                            // None audio backend.
	ALSA        *DomainAudioALSA                            // ALSA audio backend.
	CoreAudio   *DomainAudioCoreAudio                       // CoreAudio audio backend.
	Jack        *DomainAudioJack                            // Jack audio backend.
	OSS         *DomainAudioOSS                             // OSS audio backend.
	PulseAudio  *DomainAudioPulseAudio                      // PulseAudio audio backend.
	SDL         *DomainAudioSDL                             // SDL audio backend.
	SPICE       *DomainAudioSPICE                           // SPICE audio backend.
	File        *DomainAudioFile                            // File audio backend.
	DBus        *DomainAudioDBus                            // D-Bus audio backend.
	PipeWire    *DomainAudioPipeWire                        // PipeWire audio backend.
}

// DomainAudioNone represents the `<none>` sub-element of `<audio>`, configuring a dummy audio backend.
type DomainAudioNone struct {
	Input  *DomainAudioNoneChannel   // Input channel settings.
	Output *DomainAudioNoneChannel  // Output channel settings.
}

// DomainAudioNoneChannel embeds DomainAudioChannel, representing an audio channel for the "none" backend.
type DomainAudioNoneChannel struct {
	DomainAudioChannel
}

// DomainAudioALSA represents the `<alsa>` sub-element of `<audio>`, configuring an ALSA audio backend.
type DomainAudioALSA struct {
	Input  *DomainAudioALSAChannel   // Input channel settings.
	Output *DomainAudioALSAChannel  // Output channel settings.
}

// DomainAudioALSAChannel embeds DomainAudioChannel, representing an audio channel for the ALSA backend.
type DomainAudioALSAChannel struct {
	DomainAudioChannel
	Dev string  // Path to the host device node.
}

// DomainAudioCoreAudio represents the `<coreaudio>` sub-element of `<audio>`, configuring a CoreAudio backend.
type DomainAudioCoreAudio struct {
	Input  *DomainAudioCoreAudioChannel   // Input channel settings.
	Output *DomainAudioCoreAudioChannel  // Output channel settings.
}

// DomainAudioCoreAudioChannel embeds DomainAudioChannel, representing an audio channel for the CoreAudio backend.
type DomainAudioCoreAudioChannel struct {
	DomainAudioChannel
	BufferCount uint  // Number of buffers.
}

// DomainAudioJack represents the `<jack>` sub-element of `<audio>`, configuring a Jack audio backend.
type DomainAudioJack struct {
	Input  *DomainAudioJackChannel   // Input channel settings.
	Output *DomainAudioJackChannel  // Output channel settings.
}

// DomainAudioJackChannel embeds DomainAudioChannel, representing an audio channel for the Jack backend.
type DomainAudioJackChannel struct {
	DomainAudioChannel
	ServerName   string    // Jack server instance name.
	ClientName   string    // Client name.
	ConnectPorts string  // Regular expression of Jack client port names.
	ExactName    string     // Use exact client name ("yes", "no").
}

// DomainAudioOSS represents the `<oss>` sub-element of `<audio>`, configuring an OSS audio backend.
type DomainAudioOSS struct {
	TryMMap   string    // Attempt to use mmap for data transfer ("yes", "no").
	Exclusive string  // Enforce exclusive access to host device ("yes", "no").
	DSPPolicy *int              // Timing policy of the device.
	Input  *DomainAudioOSSChannel   // Input channel settings.
	Output *DomainAudioOSSChannel  // Output channel settings.
}

// DomainAudioOSSChannel embeds DomainAudioChannel, representing an audio channel for the OSS backend.
type DomainAudioOSSChannel struct {
	DomainAudioChannel
	Dev         string     // Path to the host device node.
	BufferCount uint    // Number of buffers.
	TryPoll     string  // Attempt to use polling mode ("yes", "no").
}

// DomainAudioPipeWire represents the `<pipewire>` sub-element of `<audio>`, configuring a PipeWire audio backend.
type DomainAudioPipeWire struct {
	RuntimeDir string                         // Path to PipeWire daemon socket.
	Input      *DomainAudioPulseAudioChannel                      // Input channel settings.
	Output     *DomainAudioPulseAudioChannel                     // Output channel settings.
}

// DomainAudioPipeWireChannel embeds DomainAudioChannel, representing an audio channel for the PipeWire backend.
type DomainAudioPipeWireChannel struct {
	DomainAudioChannel
	Name       string     // Sink/source name.
	StreamName string  // Stream name.
	Latency    uint    // Desired latency in microseconds.
}

// DomainAudioPulseAudio represents the `<pulseaudio>` sub-element of `<audio>`, configuring a PulseAudio backend.
type DomainAudioPulseAudio struct {
	ServerName string                         // Hostname of the PulseAudio server.
	Input      *DomainAudioPulseAudioChannel                      // Input channel settings.
	Output     *DomainAudioPulseAudioChannel                     // Output channel settings.
}

// DomainAudioPulseAudioChannel embeds DomainAudioChannel, representing an audio channel for the PulseAudio backend.
type DomainAudioPulseAudioChannel struct {
	DomainAudioChannel
	Name       string     // Sink/source name.
	StreamName string  // Stream name.
	Latency    uint    // Desired latency in microseconds.
}

// DomainAudioSDL represents the `<sdl>` sub-element of `<audio>`, configuring an SDL audio backend.
type DomainAudioSDL struct {
	Driver string                  // SDL audio driver ("esd", "alsa", "arts", "pulseaudio").
	Input  *DomainAudioSDLChannel                  // Input channel settings.
	Output *DomainAudioSDLChannel                 // Output channel settings.
}

// DomainAudioSDLChannel embeds DomainAudioChannel, representing an audio channel for the SDL backend.
type DomainAudioSDLChannel struct {
	DomainAudioChannel
	BufferCount uint  // Number of buffers.
}

// DomainAudioSPICE represents the `<spice>` sub-element of `<audio>`, configuring a SPICE audio backend.
type DomainAudioSPICE struct {
	Input  *DomainAudioSPICEChannel   // Input channel settings.
	Output *DomainAudioSPICEChannel  // Output channel settings.
}

// DomainAudioSPICEChannel embeds DomainAudioChannel, representing an audio channel for the SPICE backend.
type DomainAudioSPICEChannel struct {
	DomainAudioChannel
}

// DomainAudioFile represents the `<file>` sub-element of `<audio>`, configuring a file audio backend.
type DomainAudioFile struct {
	Path   string                   // Path to the audio file.
	Input  *DomainAudioFileChannel                // Input channel settings.
	Output *DomainAudioFileChannel               // Output channel settings.
}

// DomainAudioFileChannel embeds DomainAudioChannel, representing an audio channel for the file backend.
type DomainAudioFileChannel struct {
	DomainAudioChannel
}

// DomainAudioDBus represents the `<dbus>` sub-element of `<audio>`, configuring a D-Bus audio backend.
type DomainAudioDBus struct {
	Input  *DomainAudioDBusChannel   // Input channel settings.
	Output *DomainAudioDBusChannel  // Output channel settings.
}

// DomainAudioDBusChannel embeds DomainAudioChannel, representing an an audio channel for the D-Bus backend.
type DomainAudioDBusChannel struct {
	DomainAudioChannel
}

// DomainAudioChannel is a base struct for audio channel settings.
type DomainAudioChannel struct {
	MixingEngine  string                        // Controls host mixing engine usage ("yes", "no").
	FixedSettings string                       // Controls dynamic setting choice ("yes", "no").
	Voices        uint                                // Number of voices to use.
	Settings      *DomainAudioChannelSettings                      // Fixed settings for frequency, channels, format.
	BufferLength  uint                          // Length of audio buffer in microseconds.
}

// DomainAudioChannelSettings represents the `<settings>` sub-element of audio channels, configuring fixed settings.
type DomainAudioChannelSettings struct {
	Frequency uint    // Frequency in HZ.
	Channels  uint     // Number of channels.
	Format    string     // Audio format ("s8", "u8", "s16", "u16", "s32", "u32", "f32").
}

// DomainVideo represents a `<video>` element, configuring video devices.
type DomainVideo struct {
	XMLName xml.Name           
	Model   DomainVideoModel      // Video model.
	Driver  *DomainVideoDriver   // Driver specific options.
	ACPI    *DomainDeviceACPI      // ACPI device configuration.
	Alias   *DomainAlias          // Identifier for the device.
	Address *DomainAddress      // Device address on the virtual bus.
}

// DomainVideoModel represents the `<model>` sub-element of `<video>`, configuring video model.
type DomainVideoModel struct {
	Type       string                  // Model type.
	Heads      uint                    // Number of screens.
	Ram        uint                    // Size of primary bar.
	VRam       uint                    // Video memory in KiB.
	VRam64     uint                    // Extends secondary bar for 64bit memory.
	VGAMem     uint                    // Size of VGA framebuffer.
	Primary    string                  // Marks as primary video device ("yes").
	Blob       string                  // Enables blob resources ("on", "off").
	EDID       string                  // Exposes device EDID blob to guest ("on", "off").
	Accel      *DomainVideoAccel       // Video acceleration settings.
	Resolution *DomainVideoResolution    // Minimum resolution.
}

// DomainVideoAccel represents the `<acceleration>` sub-element of `<model>`, configuring video acceleration.
type DomainVideoAccel struct {
	Accel3D    string     // Enables 3D acceleration ("yes", "no").
	Accel2D    string     // Enables 2D acceleration ("yes", "no").
	RenderNode string  // Path to host's DRI device.
}

// DomainVideoResolution represents the `<resolution>` sub-element of `<model>`, configuring minimum resolution.
type DomainVideoResolution struct {
	X uint  // Minimum X resolution.
	Y uint  // Minimum Y resolution.
}

// DomainVideoDriver represents the `<driver>` sub-element of `<video>`, specifying driver-specific options for a video device.
type DomainVideoDriver struct {
	Name      string       // Backend driver to use ("qemu", "vhostuser").
	VGAConf   string    // VGA configuration ("io", "on", "off").
	IOMMU     string      // Enables emulated IOMMU.
	ATS       string        // Address Translation Service support.
	Packed    string     // Whether to use packed virtqueues.
	PagePerVQ string  // Layout of notification capabilities.
}

// DomainHostdev represents a `<hostdev>` element, configuring host device assignment.
type DomainHostdev struct {
	Managed        string                         // Whether libvirt manages the PCI device ("yes", "no").
	SubsysUSB      *DomainHostdevSubsysUSB                           // USB host device.
	SubsysSCSI     *DomainHostdevSubsysSCSI                          // SCSI host device.
	SubsysSCSIHost *DomainHostdevSubsysSCSIHost                      // SCSI host bus adapter.
	SubsysPCI      *DomainHostdevSubsysPCI                           // PCI host device.
	SubsysMDev     *DomainHostdevSubsysMDev                          // Mediated device.
	CapsStorage    *DomainHostdevCapsStorage                         // Block/character storage device.
	CapsMisc       *DomainHostdevCapsMisc                            // Block/character misc device.
	CapsNet        *DomainHostdevCapsNet                             // Block/character network device.
	Boot           *DomainDeviceBoot                              // Specifies that the device is bootable.
	ROM            *DomainROM                                      // PCI device's ROM configuration.
	ACPI           *DomainDeviceACPI                              // ACPI device configuration.
	Alias          *DomainAlias                                  // Identifier for the device.
	Address        *DomainAddress                              // Device address on the virtual bus.
}

// DomainHostdevSubsysUSB represents the `<usb>` sub-element of `<hostdev>`, configuring a USB host device.
type DomainHostdevSubsysUSB struct {
	Source *DomainHostdevSubsysUSBSource  // USB device source.
}

// DomainHostdevSubsysUSBSource represents the `<source>` sub-element of `<usb>`, configuring a USB device source.
type DomainHostdevSubsysUSBSource struct {
	GuestReset    string                            // Guest initiated device reset requests ("off", "uninitialized", "on").
	StartUpPolicy string                         // Policy if device is not found ("mandatory", "requisite", "optional").
	Address       *DomainAddressUSB                                   // USB device address.
	Product       *DomainHostDevProductVendorID                       // USB product ID.
	Vendor        *DomainHostDevProductVendorID                        // USB vendor ID.
}

// DomainHostDevProductVendorID represents vendor/product IDs for USB devices.
type DomainHostDevProductVendorID struct {
	ID string  // Vendor/product ID.
}

// DomainHostdevSubsysSCSI represents the `<scsi>` sub-element of `<hostdev>`, configuring a SCSI host device.
type DomainHostdevSubsysSCSI struct {
	SGIO      string                              // Whether unprivileged SG_IO commands are filtered.
	RawIO     string                             // Whether the LUN needs rawio capability.
	Source    *DomainHostdevSubsysSCSISource                    // SCSI device source.
	ReadOnly  *DomainDiskReadOnly                            // Device cannot be modified.
	Shareable *DomainDiskShareable                          // Device is shared.
}

// DomainHostdevSubsysSCSISource represents the `<source>` sub-element of `<scsi>`, configuring a SCSI device source.
type DomainHostdevSubsysSCSISource struct {
	Host  *DomainHostdevSubsysSCSISourceHost   // SCSI host adapter.
	ISCSI *DomainHostdevSubsysSCSISourceISCSI  // iSCSI target.
}

// DomainHostdevSubsysSCSISourceHost represents a SCSI host adapter source.
type DomainHostdevSubsysSCSISourceHost struct {
	Adapter *DomainHostdevSubsysSCSIAdapter  // SCSI adapter.
	Address *DomainAddressDrive              // SCSI device address.
}

// DomainHostdevSubsysSCSIAdapter represents a SCSI adapter.
type DomainHostdevSubsysSCSIAdapter struct {
	Name string  // SCSI adapter name.
}

// DomainHostdevSubsysSCSISourceISCSI represents an iSCSI target source.
type DomainHostdevSubsysSCSISourceISCSI struct {
	Name      string                                      // iSCSI target name.
	Host      []DomainDiskSourceHost                           // List of iSCSI hosts.
	Auth      *DomainDiskAuth                                  // Authentication credentials.
	Initiator *DomainHostdevSubsysSCSISourceInitiator     // Initiator IQN.
}

// DomainHostdevSubsysSCSISourceInitiator represents an initiator IQN.
type DomainHostdevSubsysSCSISourceInitiator struct {
	IQN DomainHostdevSubsysSCSISourceIQN  // Initiator IQN.
}

// DomainHostdevSubsysSCSISourceIQN represents an initiator IQN name.
type DomainHostdevSubsysSCSISourceIQN struct {
	Name string  // Initiator IQN name.
}

// DomainHostdevSubsysSCSIHost represents the `<scsi_host>` sub-element of `<hostdev>`, configuring a SCSI host bus adapter.
type DomainHostdevSubsysSCSIHost struct {
	Model  string                              // SCSI host model.
	Source *DomainHostdevSubsysSCSIHostSource                // SCSI host source.
}

// DomainHostdevSubsysSCSIHostSource represents the `<source>` sub-element of `<scsi_host>`, configuring a SCSI host source.
type DomainHostdevSubsysSCSIHostSource struct {
	Protocol string  // Protocol ("vhost").
	WWPN     string      // vhost_scsi WWPN.
}

// DomainHostdevSubsysPCI represents the `<pci>` sub-element of `<hostdev>`, configuring a PCI host device.
type DomainHostdevSubsysPCI struct {
	Display string                         // Enables vGPU as display device ("on", "off").
	RamFB   string                           // Provides a memory framebuffer device ("on", "off").
	Driver  *DomainHostdevSubsysPCIDriver                  // Host driver to bind to device.
	Source  *DomainHostdevSubsysPCISource                  // PCI device source.
	Teaming *DomainInterfaceTeaming                       // Teaming configuration.
}

// DomainHostdevSubsysPCIDriver represents the `<driver>` sub-element of `<pci>`, configuring the host driver for a PCI device.
type DomainHostdevSubsysPCIDriver struct {
	Name  string   // Host driver name.
	Model string  // Driver model.
}

// DomainHostdevSubsysPCISource represents the `<source>` sub-element of `<pci>`, configuring a PCI device source.
type DomainHostdevSubsysPCISource struct {
	WriteFiltering string             // Controls write access to PCI configuration space ("yes", "no").
	Address        *DomainAddressPCI                      // PCI device address.
}

// DomainHostdevSubsysMDev represents the `<mdev>` sub-element of `<hostdev>`, configuring a mediated device.
type DomainHostdevSubsysMDev struct {
	Model   string                          // Device API ("vfio-pci", "vfio-ccw", "vfio-ap").
	Display string                          // Enables accelerated remote desktop ("on", "off").
	RamFB   string                            // Provides a memory framebuffer device ("on", "off").
	Source  *DomainHostdevSubsysMDevSource                  // Mediated device source.
}

// DomainHostdevSubsysMDevSource represents the `<source>` sub-element of `<mdev>`, configuring a mediated device source.
type DomainHostdevSubsysMDevSource struct {
	Address *DomainAddressMDev  // Mediated device address.
}

// DomainAddressMDev represents the `<address>` sub-element of `<mdev>`, specifying the UUID of the mediated device.
type DomainAddressMDev struct {
	UUID string  // UUID of the mediated device.
}

// DomainHostdevCapsStorage represents the `<storage>` sub-element of `<hostdev>`, configuring a block/character storage device.
type DomainHostdevCapsStorage struct {
	Source *DomainHostdevCapsStorageSource  // Block device source.
}

// DomainHostdevCapsStorageSource represents the `<source>` sub-element of `<storage>`, configuring a block device source.
type DomainHostdevCapsStorageSource struct {
	Block string  // Path to the block device.
}

// DomainHostdevCapsMisc represents the `<misc>` sub-element of `<hostdev>`, configuring a block/character misc device.
type DomainHostdevCapsMisc struct {
	Source *DomainHostdevCapsMiscSource  // Character device source.
}

// DomainHostdevCapsMiscSource represents the `<source>` sub-element of `<misc>`, configuring a character device source.
type DomainHostdevCapsMiscSource struct {
	Char string  // Path to the character device.
}

// DomainHostdevCapsNet represents the `<net>` sub-element of `<hostdev>`, configuring a block/character network device.
type DomainHostdevCapsNet struct {
	Source *DomainHostdevCapsNetSource  // Network interface source.
	IP     []DomainIP                       // IP addresses.
	Route  []DomainRoute                 // Routes.
}

// DomainHostdevCapsNetSource represents the `<source>` sub-element of `<net>`, configuring a network interface source.
type DomainHostdevCapsNetSource struct {
	Interface string  // Name of the interface.
}

// DomainIP represents an `<ip>` sub-element, configuring IP addresses.
type DomainIP struct {
	Address string  // IP address.
	Family  string   // Address family.
	Prefix  *uint              // Prefix length.
}

// DomainRoute represents a `<route>` sub-element, configuring routes.
type DomainRoute struct {
	Family  string  // Address family.
	Address string  // Network address.
	Gateway string  // Gateway address.
}

// DomainRedirDev represents a `<redirdev>` element, configuring redirected devices.
type DomainRedirDev struct {
	XMLName  xml.Name               
	Bus      string                  // Bus type ("usb").
	Source   *DomainChardevSource                // Host side of the tunnel.
	Protocol *DomainChardevProtocol            // Protocol for the character device.
	Boot     *DomainDeviceBoot                     // Specifies that the device is bootable.
	ACPI     *DomainDeviceACPI                     // ACPI device configuration.
	Alias    *DomainAlias                         // Identifier for the device.
	Address  *DomainAddress                     // Device address on the virtual bus.
}

// DomainRedirFilter represents the `<redirfilter>` element, creating filter rules for redirection.
type DomainRedirFilter struct {
	USB []DomainRedirFilterUSB  // List of USB device filter rules.
}

// DomainRedirFilterUSB represents a `<usbdev>` sub-element of `<redirfilter>`, defining a USB device filter rule.
type DomainRedirFilterUSB struct {
	Class   *uint            // USB Class code.
	Vendor  *uint           // USB vendor ID.
	Product *uint          // USB product ID.
	Version string  // Device revision.
	Allow   string           // Whether to allow or deny ("yes", "no").
}

// DomainHub represents a `<hub>` element, configuring a hub device.
type DomainHub struct {
	Type    string             // Hub type ("usb").
	ACPI    *DomainDeviceACPI     // ACPI device configuration.
	Alias   *DomainAlias         // Identifier for the device.
	Address *DomainAddress     // Device address on the virtual bus.
}

// DomainWatchdog represents a `<watchdog>` element, configuring a virtual hardware watchdog device.
type DomainWatchdog struct {
	XMLName xml.Name          
	Model   string                        // Emulated watchdog device model.
	Action  string             // Action to take when watchdog expires.
	ACPI    *DomainDeviceACPI                   // ACPI device configuration.
	Alias   *DomainAlias                       // Identifier for the device.
	Address *DomainAddress                   // Device address on the virtual bus.
}

// DomainMemBalloon represents a `<memballoon>` element, configuring a virtual memory balloon device.
type DomainMemBalloon struct {
	XMLName           xml.Name                
	Model             string                              // Type of balloon device.
	AutoDeflate       string                   // Enables/disables memory release ("on", "off").
	FreePageReporting string                   // Enables/disables returning unused pages ("on", "off").
	Driver            *DomainMemBalloonDriver                 // Driver specific options.
	Stats             *DomainMemBalloonStats                   // Statistics collection.
	ACPI              *DomainDeviceACPI                         // ACPI device configuration.
	Alias             *DomainAlias                             // Identifier for the device.
	Address           *DomainAddress                         // Device address on the virtual bus.
}

// DomainMemBalloonDriver represents the `<driver>` sub-element of `<memballoon>`, specifying driver-specific options.
type DomainMemBalloonDriver struct {
	IOMMU     string      // Enables emulated IOMMU.
	ATS       string        // Address Translation Service support.
	Packed    string     // Whether to use packed virtqueues.
	PagePerVQ string  // Layout of notification capabilities.
}

// DomainMemBalloonStats represents the `<stats>` sub-element of `<memballoon>`, configuring statistics collection.
type DomainMemBalloonStats struct {
	Period uint  // Period for statistics collection.
}

// DomainRNG represents an `<rng>` element, configuring a virtual random number generator device.
type DomainRNG struct {
	XMLName xml.Name        
	Model   string                      // Type of RNG device.
	Driver  *DomainRNGDriver                  // Driver specific options.
	Rate    *DomainRNGRate                      // Limits rate at which entropy can be consumed.
	Backend *DomainRNGBackend                // Source of entropy.
	ACPI    *DomainDeviceACPI                   // ACPI device configuration.
	Alias   *DomainAlias                     // Identifier for the device.
	Address *DomainAddress                 // Device address on the virtual bus.
}

// DomainRNGDriver represents the `<driver>` sub-element of `<rng>`, specifying driver-specific options.
type DomainRNGDriver struct {
	IOMMU     string      // Enables emulated IOMMU.
	ATS       string        // Address Translation Service support.
	Packed    string     // Whether to use packed virtqueues.
	PagePerVQ string  // Layout of notification capabilities.
}

// DomainRNGRate represents the `<rate>` sub-element of `<rng>`, configuring the rate at which entropy can be consumed.
type DomainRNGRate struct {
	Bytes  uint           // Bytes permitted to be consumed per period.
	Period uint  // Duration of a period in milliseconds.
}

// DomainRNGBackend represents the `<backend>` sub-element of `<rng>`, specifying the source of entropy.
type DomainRNGBackend struct {
	Random  *DomainRNGBackendRandom   // Non-blocking character device.
	EGD     *DomainRNGBackendEGD      // EGD protocol source.
	BuiltIn *DomainRNGBackendBuiltIn  // QEMU builtin random generator.
}

// DomainRNGBackendRandom represents a non-blocking character device source for RNG.
type DomainRNGBackendRandom struct {
	Device string  // Path to the character device.
}

// DomainRNGBackendEGD represents an EGD protocol source for RNG.
type DomainRNGBackendEGD struct {
	Source   *DomainChardevSource      // Character device for EGD source.
	Protocol *DomainChardevProtocol  // Protocol for the character device.
}

// DomainRNGBackendBuiltIn represents a QEMU builtin random generator source for RNG.
type DomainRNGBackendBuiltIn struct {
}

// DomainNVRAM represents an `<nvram>` element, configuring NVRAM devices.
type DomainNVRAM struct {
	ACPI    *DomainDeviceACPI     // ACPI device configuration.
	Alias   *DomainAlias         // Identifier for the device.
	Address *DomainAddress     // Device address on the virtual bus.
}

// DomainPanic represents a `<panic>` element, configuring a panic device.
type DomainPanic struct {
	XMLName xml.Name          
	Model   string             // Type of panic device.
	ACPI    *DomainDeviceACPI                  // ACPI device configuration.
	Alias   *DomainAlias                      // Identifier for the device.
	Address *DomainAddress                  // Device address on the virtual bus.
}

// DomainShmem represents a `<shmem>` element, configuring a shared memory device.
type DomainShmem struct {
	XMLName xml.Name          
	Name    string                        // Name to identify the shared memory.
	Role    string              // Shared memory migratability role ("master", "peer").
	Size    *DomainShmemSize                   // Size of the shared memory.
	Model   *DomainShmemModel                 // Model of the underlying device.
	Server  *DomainShmemServer                // Server socket configuration.
	MSI     *DomainShmemMSI                     // MSI interrupts configuration.
	ACPI    *DomainDeviceACPI                  // ACPI device configuration.
	Alias   *DomainAlias                      // Identifier for the device.
	Address *DomainAddress                  // Device address on the virtual bus.
}

// DomainShmemSize represents the `<size>` sub-element of `<shmem>`, configuring the size of the shared memory.
type DomainShmemSize struct {
	Value uint             // Size value.
	Unit  string .
}

// DomainShmemModel represents the `<model>` sub-element of `<shmem>`, configuring the model of the underlying device.
type DomainShmemModel struct {
	Type string  // Model type ("ivshmem", "ivshmem-plain", "ivshmem-doorbell").
}

// DomainShmemServer represents the `<server>` sub-element of `<shmem>`, configuring a server socket.
type DomainShmemServer struct {
	Path string  // Absolute path to the unix socket.
}

// DomainShmemMSI represents the `<msi>` sub-element of `<shmem>`, configuring MSI interrupts.
type DomainShmemMSI struct {
	Enabled   string    // Enables/disables MSI interrupts ("on", "off").
	Vectors   uint      // Number of interrupt vectors.
	IOEventFD string  // Enables/disables ioeventfd ("on", "off").
}

// DomainMemorydev represents a `<memory>` element within `<devices>`, configuring memory devices.
type DomainMemorydev struct {
	XMLName xml.Name          
	Model   string                        // Type of memory module.
	Access  string             // Memory access mode ("shared", "private").
	Discard string             // Discard of data on per module basis ("yes", "no").
	UUID    string                    // UUID to identify the nvdimm module.
	Source  *DomainMemorydevSource             // Source of the memory used for the device.
	Target  *DomainMemorydevTarget             // Placement and sizing of the added memory.
	ACPI    *DomainDeviceACPI                   // ACPI device configuration.
	Alias   *DomainAlias                       // Identifier for the device.
	Address *DomainAddress                   // Device address on the virtual bus.
}

// DomainMemorydevSource represents the `<source>` sub-element of `<memorydev>`, configuring the source of memory.
type DomainMemorydevSource struct {
	NodeMask  string                           // Overrides default set of NUMA nodes.
	PageSize  *DomainMemorydevSourcePagesize             // Overrides default host page size.
	Path      string                               // Path in the host that backs the nvdimm module.
	AlignSize *DomainMemorydevSourceAlignsize           // Page size alignment for mmap.
	PMem      *DomainMemorydevSourcePMem                     // Enables persistent memory.
}

// DomainMemorydevSourcePagesize represents the `<pagesize>` sub-element of `<source>`, configuring page size.
type DomainMemorydevSourcePagesize struct {
	Value uint64           // Page size value.
	Unit  string .
}

// DomainMemorydevSourceAlignsize represents the `<alignsize>` sub-element of `<source>`, configuring page size alignment.
type DomainMemorydevSourceAlignsize struct {
	Value uint64           // Alignment size value.
	Unit  string .
}

// DomainMemorydevSourcePMem represents the `<pmem>` sub-element of `<source>`, enabling persistent memory.
type DomainMemorydevSourcePMem struct {
}

// DomainMemorydevTarget represents the `<target>` sub-element of `<memorydev>`, configuring placement and sizing of added memory.
type DomainMemorydevTarget struct {
	DynamicMemslots string                           // Allows hypervisor to spread memory into multiple memory slots.
	Size            *DomainMemorydevTargetSize                                 // Size of the added memory.
	Node            *DomainMemorydevTargetNode                                 // Guest NUMA node to attach memory to.
	Label           *DomainMemorydevTargetLabel                               // Configures namespace label storage.
	Block           *DomainMemorydevTargetBlock                               // Size of an individual block.
	Requested       *DomainMemorydevTargetRequested                       // Total size exposed to the guest.
	ReadOnly        *DomainMemorydevTargetReadOnly                         // Marks vNVDIMM as read-only.
	Address         *DomainMemorydevTargetAddress                           // Physical address in memory where device is mapped.
}

// DomainMemorydevTargetSize represents the `<size>` sub-element of `<target>`, configuring the size of added memory.
type DomainMemorydevTargetSize struct {
	Value uint             // Size value.
	Unit  string .
}

// DomainMemorydevTargetNode represents the `<node>` sub-element of `<target>`, configuring the guest NUMA node.
type DomainMemorydevTargetNode struct {
	Value uint  // Guest NUMA node ID.
}

// DomainMemorydevTargetLabel represents the `<label>` sub-element of `<target>`, configuring namespace label storage.
type DomainMemorydevTargetLabel struct {
	Size *DomainMemorydevTargetSize  // Size of namespaces label storage.
}

// DomainMemorydevTargetBlock represents the `<block>` sub-element of `<target>`, configuring the size of an individual block.
type DomainMemorydevTargetBlock struct {
	Value uint             // Block size value.
	Unit  string .
}

// DomainMemorydevTargetRequested represents the `<requested>` sub-element of `<target>`, configuring the total size exposed to the guest.
type DomainMemorydevTargetRequested struct {
	Value uint             // Requested size value.
	Unit  string .
}

// DomainMemorydevTargetReadOnly represents the `<readonly>` sub-element of `<target>`, marking vNVDIMM as read-only.
type DomainMemorydevTargetReadOnly struct {
}

// DomainMemorydevTargetAddress represents the `<address>` sub-element of `<target>`, configuring the physical address in memory.
type DomainMemorydevTargetAddress struct {
	Base *uint  // Base address.
}

// DomainIOMMU represents an `<iommu>` element, configuring an IOMMU device.
type DomainIOMMU struct {
	Model   string                // IOMMU model.
	Driver  *DomainIOMMUDriver        // Driver specific options.
	ACPI    *DomainDeviceACPI           // ACPI device configuration.
	Alias   *DomainAlias               // Identifier for the device.
	Address *DomainAddress           // Device address on the virtual bus.
}

// DomainIOMMUDriver represents the `<driver>` sub-element of `<iommu>`, specifying driver-specific options for an IOMMU device.
type DomainIOMMUDriver struct {
	IntRemap       string        // Enables interrupt remapping ("on", "off").
	CachingMode    string    // Enables VT-d caching mode ("on", "off").
	EIM            string             // Configures Extended Interrupt Mode ("on", "off").
	IOTLB          string           // Enables IOTLB ("on", "off").
	AWBits         uint           // Sets address width.
	DMATranslation string  // Turns off DMA translation ("on", "off").
	Passthrough    string     // Enables passthrough ("on", "off").
	XTSup          string           // Enables x2APIC mode ("on", "off").
}

// DomainVSock represents a `<vsock>` element, configuring a vsock host/guest interface.
type DomainVSock struct {
	XMLName xml.Name          
	Model   string             // Vsock model.
	CID     *DomainVSockCID                     // CID assigned to the guest.
	Driver  *DomainVSockDriver                // Driver specific options.
	ACPI    *DomainDeviceACPI                  // ACPI device configuration.
	Alias   *DomainAlias                      // Identifier for the device.
	Address *DomainAddress                  // Device address on the virtual bus.
}

// DomainVSockCID represents the `<cid>` sub-element of `<vsock>`, configuring the CID assigned to the guest.
type DomainVSockCID struct {
	Auto    string     // Automatically assign a free CID ("yes", "no").
	Address string  // CID assigned to the guest.
}

// DomainVSockDriver represents the `<driver>` sub-element of `<vsock>`, specifying driver-specific options.
type DomainVSockDriver struct {
	IOMMU     string      // Enables emulated IOMMU.
	ATS       string        // Address Translation Service support.
	Packed    string     // Whether to use packed virtqueues.
	PagePerVQ string  // Layout of notification capabilities.
}

// DomainCrypto represents a `<crypto>` element, configuring a crypto device.
type DomainCrypto struct {
	Model   string              // Crypto model.
	Type    string               // Crypto type.
	Backend *DomainCryptoBackend             // Backend configuration.
	Alias   *DomainAlias                     // Identifier for the device.
	Address *DomainAddress                 // Device address on the virtual bus.
}

// DomainCryptoBackend represents the `<backend>` sub-element of `<crypto>`, configuring the crypto backend.
type DomainCryptoBackend struct {
	BuiltIn *DomainCryptoBackendBuiltIn  // Built-in backend.
	LKCF    *DomainCryptoBackendLKCF     // LKCF backend.
	Queues  uint                         // Number of virt queues.
}

// DomainCryptoBackendBuiltIn represents a built-in crypto backend.
type DomainCryptoBackendBuiltIn struct {
}

// DomainCryptoBackendLKCF represents an LKCF crypto backend.
type DomainCryptoBackendLKCF struct {
}

// DomainPStore represents a `<pstore>` element, configuring a pstore device.
type DomainPStore struct {
	Backend string             // Desired backend ("acpi-erst").
	Path    string                     // Path in the host that backs the pstore device.
	Size    DomainPStoreSize           // Size of the persistent storage.
	ACPI    *DomainDeviceACPI          // ACPI device configuration.
	Alias   *DomainAlias              // Identifier for the device.
	Address *DomainAddress          // Device address on the virtual bus.
}

// DomainPStoreSize represents the `<size>` sub-element of `<pstore>`, configuring the size of persistent storage.
type DomainPStoreSize struct {
	Size uint64  // Size value.
	Unit string .
}

// DomainSecLabel represents the `<seclabel>` element, controlling the operation of security drivers.
type DomainSecLabel struct {
	Type       string        // Security label type ("static", "dynamic", "none").
	Model      string       // Security model name.
	Relabel    string     // Whether automatic relabeling is enabled ("yes", "no").
	Label      string            // Full security label to assign.
	ImageLabel string       // Security label used on resources (output only).
	BaseLabel  string        // Base security label for dynamic generation.
}

// DomainKeyWrap represents the `<keywrap>` element, configuring S390 cryptographic key management operations.
type DomainKeyWrap struct {
	Ciphers []DomainKeyWrapCipher  // List of cipher configurations.
}

// DomainKeyWrapCipher represents a `<cipher>` sub-element of `<keywrap>`, configuring a cryptographic key wrapping algorithm.
type DomainKeyWrapCipher struct {
	Name  string   // Algorithm name ("aes", "dea").
	State string  // Whether operations are turned on ("on", "off").
}

// LaunchSecurity represents the `<launchSecurity>` element, providing guest owner input for creating encrypted VMs.
type LaunchSecurity struct {
	SEV    *DomainLaunchSecuritySEV     // AMD SEV feature.
	SEVSNP *DomainLaunchSecuritySEVSNP  // AMD SEV-SNP feature.
	S390PV *DomainLaunchSecurityS390PV  // S390 protected virtualization.
	TDX    *DomainLaunchSecurityTDX     // Intel TDX feature.
}

// DomainLaunchSecuritySEV represents the `<sev>` sub-element of `<launchSecurity>`, configuring AMD SEV.
type DomainLaunchSecuritySEV struct {
	KernelHashes    string  // Whether kernel hashes are included in measurement.
	CBitPos         *uint                       // C-bit location in guest page table entry.
	ReducedPhysBits *uint               // Physical address bit reduction.
	Policy          *uint                        // Guest policy enforced by SEV firmware.
	DHCert          string                       // Guest owner's base64 encoded Diffie-Hellman (DH) key.
	Session         string                       // Guest owner's base64 encoded session blob.
}

// DomainLaunchSecuritySEVSNP represents the `<sev-snp>` sub-element of `<launchSecurity>`, configuring AMD SEV-SNP.
type DomainLaunchSecuritySEVSNP struct {
	KernelHashes            string   // Whether kernel hashes are included in measurement.
	AuthorKey               string   contains 'AUTHOR_KEY' field.
	VCEK                    string           // Whether guest can choose between VLEK or VCEK.
	CBitPos                 *uint                        // C-bit location in guest page table entry.
	ReducedPhysBits         *uint                // Physical address bit reduction.
	Policy                  *uint64                       // Guest policy enforced by SEV-SNP firmware.
	GuestVisibleWorkarounds string   // Hypervisor-defined workarounds.
	IDBlock                 string             // 'ID Block' structure for SNP_LAUNCH_FINISH command.
	IDAuth                  string              // 'ID Authentication Information Structure'.
	HostData                string            // User-defined blob to provide to the guest.
}

// DomainLaunchSecurityS390PV represents the `<s390-pv>` sub-element of `<launchSecurity>`, configuring S390 protected virtualization.
type DomainLaunchSecurityS390PV struct {
}

// DomainLaunchSecurityTDX represents the `<tdx>` sub-element of `<launchSecurity>`, configuring Intel TDX.
type DomainLaunchSecurityTDX struct {
	Policy                 *uint                                                 // Guest TD attributes passed as TD_PARAMS.ATTRIBUTES.
	MrConfigId             string                                  // ID for non-owner-defined configuration.
	MrOwner                string                                     // ID for the guest TD’s owner.
	MrOwnerConfig          string                               // ID for owner-defined configuration.
	QuoteGenerationService *DomainLaunchSecurityTDXQGS  // Quote Generation Service (QGS) daemon socket address.
}

// DomainLaunchSecurityTDXQGS represents the `<quoteGenerationService>` sub-element of `<tdx>`, configuring QGS daemon socket address.
type DomainLaunchSecurityTDXQGS struct {
	Path string  // UNIX socket address.
}

// DomainQEMUCommandline represents the `qemu:commandline` element, allowing passing arbitrary command-line arguments and environment variables to the QEMU process.
type DomainQEMUCommandline struct {
	XMLName xml.Name                   
	Args    []DomainQEMUCommandlineArg  // List of command-line arguments.
	Envs    []DomainQEMUCommandlineEnv  // List of environment variables.
}

// DomainQEMUCommandlineArg represents an `<arg>` sub-element of `qemu:commandline`, specifying a command-line argument.
type DomainQEMUCommandlineArg struct {
	Value string  // Argument value.
}

// DomainQEMUCommandlineEnv represents an `<env>` sub-element of `qemu:commandline`, specifying an environment variable.
type DomainQEMUCommandlineEnv struct {
	Name  string           // Environment variable name.
	Value string  // Environment variable value.
}

// DomainQEMUCapabilities represents the `qemu:capabilities` element, allowing adding or deleting specific QEMU capabilities.
type DomainQEMUCapabilities struct {
	XMLName xml.Name                    
	Add     []DomainQEMUCapabilitiesEntry  // List of capabilities to add.
	Del     []DomainQEMUCapabilitiesEntry  // List of capabilities to delete.
}

// DomainQEMUCapabilitiesEntry represents an `add` or `del` sub-element of `qemu:capabilities`, specifying a QEMU capability.
type DomainQEMUCapabilitiesEntry struct {
	Name string  // Capability name.
}

// DomainQEMUOverride represents the `qemu:override` element, allowing overriding properties of QEMU devices.
type DomainQEMUOverride struct {
	XMLName xml.Name                   
	Devices []DomainQEMUOverrideDevice  // List of devices with overridden properties.
}

// DomainQEMUOverrideDevice represents a `<device>` sub-element of `qemu:override`, specifying overridden device properties.
type DomainQEMUOverrideDevice struct {
	Alias    string                         // Device alias.
	Frontend DomainQEMUOverrideFrontend  // Frontend properties.
}

// DomainQEMUOverrideFrontend represents the `<frontend>` sub-element of `<device>`, specifying frontend properties.
type DomainQEMUOverrideFrontend struct {
	Properties []DomainQEMUOverrideProperty  // List of overridden properties.
}

// DomainQEMUOverrideProperty represents a `<property>` sub-element of `<frontend>`, specifying an overridden property.
type DomainQEMUOverrideProperty struct {
	Name  string             // Property name.
	Type  string   // Property type.
	Value string  // Property value.
}

// DomainQEMUDeprecation represents the `qemu:deprecation` element, configuring QEMU deprecation behavior.
type DomainQEMUDeprecation struct {
	XMLName  xml.Name 
	Behavior string    // Deprecation behavior.
}

// LXCNamespace represents the `lxc:namespace` element, configuring LXC namespace sharing.
type DomainLXCNamespace struct {
	XMLName  xml.Name              
	ShareNet *DomainLXCNamespaceMap  // Share network namespace.
	ShareIPC *DomainLXCNamespaceMap  // Share IPC namespace.
	ShareUTS *DomainLXCNamespaceMap  // Share UTS namespace.
}

// DomainLXCNamespaceMap represents a namespace mapping within `lxc:namespace`.
type DomainLXCNamespaceMap struct {
	Type  string   // Namespace type.
	Value string  // Namespace value.
}

// BHyveCommandline represents the `bhyve:commandline` element, allowing passing arbitrary command-line arguments and environment variables to the BHyve process.
type DomainBHyveCommandline struct {
	XMLName xml.Name                   
	Args    []DomainBHyveCommandlineArg  // List of command-line arguments.
	Envs    []DomainBHyveCommandlineEnv  // List of environment variables.
}

// DomainBHyveCommandlineArg represents an `<arg>` sub-element of `bhyve:commandline`, specifying a command-line argument.
type DomainBHyveCommandlineArg struct {
	Value string  // Argument value.
}

// DomainBHyveCommandlineEnv represents an `<env>` sub-element of `bhyve:commandline`, specifying an environment variable.
type DomainBHyveCommandlineEnv struct {
	Name  string           // Environment variable name.
	Value string  // Environment variable value.
}

// VMWareDataCenterPath represents the `vmware:datacenterpath` element, specifying the VMware datacenter path.
type DomainVMWareDataCenterPath struct {
	XMLName xml.Name 
	Value   string    // Datacenter path value.
}

// XenCommandline represents the `xen:commandline` element, allowing passing arbitrary command-line arguments to the Xen process.
type DomainXenCommandline struct {
	XMLName xml.Name                   
	Args    []DomainXenCommandlineArg  // List of command-line arguments.
}

// DomainXenCommandlineArg represents an `<arg>` sub-element of `xen:commandline`, specifying a command-line argument.
type DomainXenCommandlineArg struct {
	Value string  // Argument value.
}
