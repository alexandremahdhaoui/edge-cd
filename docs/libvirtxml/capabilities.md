# libvirtxml.Caps

The `Caps` struct in `libvirt.org/go/libvirtxml` represents the overall driver capabilities of the libvirt host. This includes information about the host CPU, guest architectures, and other system-level capabilities. This document details the `Caps` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Driver capabilities XML format](https://libvirt.org/formatcaps.html).

## `Caps` Struct Definition

```go
// Caps represents the overall driver capabilities of the libvirt host.
type Caps struct {
	XMLName [xml.Name](https://pkg.go.dev/encoding/xml#Name)  // XMLName is the XML element name, typically "capabilities".
	Host    *CapsHost       // Host provides details about the host's capabilities.
	Guests  []CapsGuest     // Guests describes the capabilities for specific guest architectures.
}
```

## Nested Structs

### `CapsHost`

The `CapsHost` struct provides details about the host's capabilities.

```go
// CapsHost provides details about the host's capabilities.
type CapsHost struct {
	UUID              string
	CPU               *CapsHostCPU
	PowerManagement   *CapsHostPowerManagement
	IOMMU             *CapsHostIOMMU
	MigrationFeatures *CapsHostMigrationFeatures
	NUMA              *CapsHostNUMATopology
	Cache             *CapsHostCache
	MemoryBandwidth   *CapsHostMemoryBandwidth
	SecModel          []CapsHostSecModel
}
```

### `CapsGuest`

The `CapsGuest` struct describes the capabilities for a specific guest architecture.

```go
// CapsGuest describes the capabilities for a specific guest architecture.
type CapsGuest struct {
	OSType   string
	Arch     CapsGuestArch
	Features *CapsGuestFeatures
}
```

## Mapping to Libvirt XML

The `Caps` struct corresponds to the `<capabilities>` element in libvirt XML.

*   The `Host` field maps to the `<host>` child element.
*   The `Guests` field maps to a list of `<guest>` child elements.

The `CapsHost` struct maps to the `<host>` element, and its fields map to the respective child elements like `<cpu>`, `<power_management>`, etc.

The `CapsGuest` struct maps to the `<guest>` element, and its fields map to child elements like `<os_type>`, `<arch>`, and `<features>`.

For example, the following Go struct:

```go
caps := libvirtxml.Caps{
	Host: &libvirtxml.CapsHost{
		UUID: "7b55704c-29f4-11b2-a85c-9dc6ff50623f",
		CPU: &libvirtxml.CapsHostCPU{
			Arch:  "x86_64",
			Model: "Skylake-Client-noTSX-IBRS",
			Vendor: "Intel",
		},
	},
	Guests: []libvirtxml.CapsGuest{
		{
			OSType: "hvm",
			Arch: libvirtxml.CapsGuestArch{
				Name:     "x86_64",
				Emulator: "/usr/bin/qemu-system-x86_64",
				Machines: []libvirtxml.CapsGuestMachine{
					{Name: "pc-i440fx-7.1"},
					{Name: "q35", Canonical: "pc-q35-7.1"},
				},
			},
			Features: &libvirtxml.CapsGuestFeatures{
				ACPI: &libvirtxml.CapsGuestFeatureACPI{Default: "on", Toggle: "yes"},
				APIC: &libvirtxml.CapsGuestFeatureAPIC{Default: "on", Toggle: "no"},
			},
		},
	},
}
```

Would correspond to the following libvirt XML:

```xml
<capabilities>
  <host>
    <uuid>7b55704c-29f4-11b2-a85c-9dc6ff50623f</uuid>
    <cpu>
      <arch>x86_64</arch>
      <model>Skylake-Client-noTSX-IBRS</model>
      <vendor>Intel</vendor>
    </cpu>
    <power_management/>
    <migration_features/>
    <topology/>
    <cache/>
    <secmodel/>
  </host>
  <guest>
    <os_type>hvm</os_type>
    <arch name='x86_64'>
      <wordsize>64</wordsize>
      <emulator>/usr/bin/qemu-system-x86_64</emulator>
      <machine maxCpus='255'>pc-i440fx-7.1</machine>
      <machine canonical='pc-q35-7.1' maxCpus='288'>q35</machine>
    </arch>
    <features>
      <acpi default='on' toggle='yes'/>
      <apic default='on' toggle='no'/>
    </features>
  </guest>
</capabilities>
