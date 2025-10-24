# libvirtxml.DomainCaps

The `DomainCaps` struct in `libvirt.org/go/libvirtxml` represents the capabilities of a specific hypervisor for a given domain type and architecture. This includes information about supported virtual CPUs, I/O threads, operating system configurations, CPU models and features, memory backing options, devices, and various hypervisor features. This document details the `DomainCaps` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Domain capabilities XML format](https://libvirt.org/formatdomaincaps.html).

## `DomainCaps` Struct Definition

```go
// DomainCaps represents the capabilities of a specific hypervisor for a given domain type and architecture.
type DomainCaps struct {
	XMLName       [xml.Name](https://pkg.go.dev/encoding/xml#Name)                  // XMLName is the XML element name, typically "domainCapabilities".
	Path          string                                            // Path to the hypervisor binary.
	Domain        string                                            // Hypervisor type (e.g., "kvm", "qemu").
	Machine       string                                            // Machine type supported by the hypervisor.
	Arch          string                                            // CPU architecture.
	VCPU          *DomainCapsVCPU                                   // Virtual CPU capabilities.
	IOThreads     *DomainCapsIOThreads                              // I/O thread capabilities.
	OS            *DomainCapsOS                                     // Operating system capabilities.
	CPU           *DomainCapsCPU                                    // CPU capabilities.
	MemoryBacking *DomainCapsMemoryBacking                          // Memory backing capabilities.
	Devices       *DomainCapsDevices                                // Device capabilities.
	Features      *DomainCapsFeatures                               // Hypervisor features.
}
```

## Nested Structs

### `DomainCapsVCPU`

The `DomainCapsVCPU` struct describes the maximum number of virtual CPUs supported.

```go
// DomainCapsVCPU describes the maximum number of virtual CPUs supported.
type DomainCapsVCPU struct {
	Max uint 
}
```

### `DomainCapsIOThreads`

The `DomainCapsIOThreads` struct indicates whether I/O threads are supported.

```go
// DomainCapsIOThreads indicates whether I/O threads are supported.
type DomainCapsIOThreads struct {
	Supported string 
}
```

### `DomainCapsOS`

The `DomainCapsOS` struct describes the operating system capabilities, including firmware and loader options.

```go
// DomainCapsOS describes the operating system capabilities.
type DomainCapsOS struct {
	Supported string             
	Loader    *DomainCapsOSLoader 
	Enums     []DomainCapsEnum   
}
```

### `DomainCapsOSLoader`

The `DomainCapsOSLoader` struct provides details about supported OS loaders.

```go
// DomainCapsOSLoader provides details about supported OS loaders.
type DomainCapsOSLoader struct {
	Supported string       
	Values    []string     
	Enums     []DomainCapsEnum 
}
```

### `DomainCapsCPU`

The `DomainCapsCPU` struct describes the CPU capabilities, including modes, models, features, and topology.

```go
// DomainCapsCPU describes the CPU capabilities.
type DomainCapsCPU struct {
	Modes       []DomainCapsCPUMode       
	Blockers    []DomainCapsCPUBlockers   
	Enums       []DomainCapsEnum          
	MaxPhysAddr *DomainCapsCPUMaxPhysAddr 
	Features    []DomainCapsCPUFeature    
}
```

### `DomainCapsCPUMode`

The `DomainCapsCPUMode` struct describes a specific CPU mode supported by the hypervisor.

```go
// DomainCapsCPUMode describes a specific CPU mode supported by the hypervisor.
type DomainCapsCPUMode struct {
	Name        string               
	Supported   string               
	Models      []DomainCapsCPUModel 
	Vendor      string               
	MaxPhysAddr *DomainCapsCPUMaxPhysAddr 
	Features    []DomainCapsCPUFeature    
	Blockers    []DomainCapsCPUBlockers   
	Enums       []DomainCapsEnum          
}
```

### `DomainCapsCPUModel`

The `DomainCapsCPUModel` struct describes a specific CPU model.

```go
// DomainCapsCPUModel describes a specific CPU model.
type DomainCapsCPUModel struct {
	Name       string 
	Usable     string 
	Fallback   string 
	Deprecated string 
	Vendor     string 
	Canonical  string 
}
```

### `DomainCapsDevices`

The `DomainCapsDevices` struct lists the capabilities for various device types.

```go
// DomainCapsDevices lists the capabilities for various device types.
type DomainCapsDevices struct {
	Disk       *DomainCapsDevice 
	Graphics   *DomainCapsDevice 
	Video      *DomainCapsDevice 
	HostDev    *DomainCapsDevice 
	RNG        *DomainCapsDevice 
	FileSystem *DomainCapsDevice 
	TPM        *DomainCapsDevice 
	Redirdev   *DomainCapsDevice 
	Channel    *DomainCapsDevice 
	Crypto     *DomainCapsDevice 
	Interface  *DomainCapsDevice 
	Panic      *DomainCapsDevice 
	Console    *DomainCapsDevice 
}
```

### `DomainCapsDevice`

The `DomainCapsDevice` struct represents the capabilities of a generic device type.

```go
// DomainCapsDevice represents the capabilities of a generic device type.
type DomainCapsDevice struct {
	Supported string         
	Enums     []DomainCapsEnum 
}
```

### `DomainCapsFeatures`

The `DomainCapsFeatures` struct lists the capabilities for various hypervisor features.

```go
// DomainCapsFeatures lists the capabilities for various hypervisor features.
type DomainCapsFeatures struct {
	GIC               *DomainCapsFeatureGIC               
	VMCoreInfo        *DomainCapsFeatureVMCoreInfo        
	GenID             *DomainCapsFeatureGenID             
	BackingStoreInput *DomainCapsFeatureBackingStoreInput 
	Backup            *DomainCapsFeatureBackup            
	AsyncTeardown     *DomainCapsFeatureAsyncTeardown     
	S390PV            *DomainCapsFeatureS390PV            
	PS2               *DomainCapsFeaturePS2               
	TDX               *DomainCapsFeatureTDX               
	SEV               *DomainCapsFeatureSEV               
	SGX               *DomainCapsFeatureSGX               
	HyperV            *DomainCapsFeatureHyperV            
	LaunchSecurity    *DomainCapsFeatureLaunchSecurity    
}
```

## Mapping to Libvirt XML

The `DomainCaps` struct corresponds to the `<domainCapabilities>` element in libvirt XML.

*   The `Path` field maps to the `<path>` element.
*   The `Domain` field maps to the `<domain>` element.
*   The `Machine` field maps to the `<machine>` element.
*   The `Arch` field maps to the `<arch>` element.
*   The `VCPU` field maps to the `<vcpu>` element.
*   The `IOThreads` field maps to the `<iothreads>` element.
*   The `OS` field maps to the `<os>` element.
*   The `CPU` field maps to the `<cpu>` element.
*   The `MemoryBacking` field maps to the `<memoryBacking>` element.
*   The `Devices` field maps to the `<devices>` element.
*   The `Features` field maps to the `<features>` element.

For example, the following Go struct:

```go
domainCaps := libvirtxml.DomainCaps{
	Path:   "/usr/bin/qemu-system-x86_64",
	Domain: "kvm",
	Machine: "pc-i440fx-7.1",
	Arch:   "x86_64",
	VCPU: &libvirtxml.DomainCapsVCPU{
		Max: 255,
	},
	CPU: &libvirtxml.DomainCapsCPU{
		Modes: []libvirtxml.DomainCapsCPUMode{
			{
				Name:      "host-passthrough",
				Supported: "yes",
				Enums: []libvirtxml.DomainCapsEnum{
					{Name: "hostPassthroughMigratable", Values: []string{"on", "off"}},
				},
			},
			{
				Name:      "host-model",
				Supported: "yes",
				Model: &libvirtxml.DomainCapsCPUModel{
					Fallback: "allow",
					Name:     "Broadwell",
					Vendor:   "Intel",
				},
			},
		},
	},
	Devices: &libvirtxml.DomainCapsDevices{
		Disk: &libvirtxml.DomainCapsDevice{
			Supported: "yes",
			Enums: []libvirtxml.DomainCapsEnum{
				{Name: "diskDevice", Values: []string{"disk", "cdrom", "floppy", "lun"}},
				{Name: "bus", Values: []string{"ide", "fdc", "scsi", "virtio", "xen", "usb", "sata", "sd"}},
			},
		},
	},
	Features: &libvirtxml.DomainCapsFeatures{
		GIC: &libvirtxml.DomainCapsFeatureGIC{
			Supported: "yes",
			Enums: []libvirtxml.DomainCapsEnum{
				{Name: "version", Values: []string{"2", "3"}},
			},
		},
		SEV: &libvirtxml.DomainCapsFeatureSEV{
			Supported: "yes",
			CBitPos:   uint(47),
			ReducedPhysBits: uint(1),
		},
	},
}
```

Would correspond to the following libvirt XML:

```xml
<domainCapabilities>
  <path>/usr/bin/qemu-system-x86_64</path>
  <domain>kvm</domain>
  <machine>pc-i440fx-7.1</machine>
  <arch>x86_64</arch>
  <vcpu max='255'/>
  <cpu>
    <mode name='host-passthrough' supported='yes'>
      <enum name='hostPassthroughMigratable'>
        <value>on</value>
        <value>off</value>
      </enum>
    </mode>
    <mode name='host-model' supported='yes'>
      <model fallback='allow' vendor='Intel'>Broadwell</model>
    </mode>
  </cpu>
  <devices>
    <disk supported='yes'>
      <enum name='diskDevice'>
        <value>disk</value>
        <value>cdrom</value>
        <value>floppy</value>
        <value>lun</value>
      </enum>
      <enum name='bus'>
        <value>ide</value>
        <value>fdc</value>
        <value>scsi</value>
        <value>virtio</value>
        <value>xen</value>
        <value>usb</value>
        <value>sata</value>
        <value>sd</value>
      </enum>
    </disk>
  </devices>
  <features>
    <gic supported='yes'>
      <enum name='version'>
        <value>2</value>
        <value>3</value>
      </enum>
    </gic>
    <sev supported='yes'>
      <cbitpos>47</cbitpos>
      <reduced-phys-bits>1</reduced-phys-bits>
    </sev>
  </features>
</domainCapabilities>
