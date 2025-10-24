# libvirtxml.NodeDevice

The `NodeDevice` struct in `libvirt.org/go/libvirtxml` represents a device on the host's bus, which can be passed through to a guest. This document details the `NodeDevice` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Node devices XML format](https://libvirt.org/formatnode.html).

## `NodeDevice` Struct Definition

```go
// NodeDevice represents a device on the host's bus.
type NodeDevice struct {
	XMLName    xml.Name
	Name       string                 
	Path       string                 
	Parent     string                 
	Driver     *NodeDeviceDriver      
	DevNodes   []NodeDeviceDevNode    
	Capability NodeDeviceCapability // This represents one of the specific capability types.
}
```

## Common Fields

*   **`Name`**: The name of the device. This is typically derived from the bus type and address (e.g., `pci_0000_00_02_1`, `usb_1_5_3`), but can sometimes be more specific (e.g., `net_eth1_00_27_13_6a_fe_00`). This is a read-only field.
*   **`Path`**: The fully qualified sysfs path to the device. This is a read-only field.
*   **`Parent`**: Identifies the parent node in the device hierarchy. Its value corresponds to the `name` element of the parent device, or "computer" if the device has no parent.
*   **`Driver`**: Reports the driver in use for this device. The presence of this element depends on whether the underlying device manager exposes driver information.
    ```go
    // NodeDeviceDriver reports the driver in use for this device.
    type NodeDeviceDriver struct {
    	Name string 
    }
    ```
*   **`DevNodes`**: A list of associated `/dev` special files.
    ```go
    // NodeDeviceDevNode represents an associated /dev special file.
    type NodeDeviceDevNode struct {
    	Type string  // e.g., "dev" or "link"
    	Path string 
    }
    ```

## Capability Types

The `Capability` field represents various capabilities associated with a node device. The specific type of capability determines the structure and attributes within this field. The primary capability types are detailed below:

### `NodeDeviceSystemCapability`

Describes the overall host system.

```go
// NodeDeviceSystemCapability describes the overall host system.
type NodeDeviceSystemCapability struct {
	Product  string                  
	Hardware *NodeDeviceSystemHardware 
	Firmware *NodeDeviceSystemFirmware 
}

// NodeDeviceSystemHardware describes the hardware of the system.
type NodeDeviceSystemHardware struct {
	Vendor  string 
	Version string 
	Serial  string 
	UUID    string 
}

// NodeDeviceSystemFirmware describes the firmware of the system.
type NodeDeviceSystemFirmware struct {
	Vendor      string 
	Version     string 
	ReleaseData string 
}
```

### `NodeDevicePCICapability`

Describes a device on the host's PCI bus.

```go
// NodeDevicePCICapability describes a device on the host's PCI bus.
type NodeDevicePCICapability struct {
	Class        string                           
	Domain       *uint                            
	Bus          *uint                            
	Slot         *uint                            
	Function     *uint                            
	Product      NodeDeviceIDName                 
	Vendor       NodeDeviceIDName                 
	IOMMUGroup   *NodeDeviceIOMMUGroup            
	NUMA         *NodeDeviceNUMA                  
	PCIExpress   *NodeDevicePCIExpress            
	Capabilities []NodeDevicePCISubCapability      // Can contain virt_functions, pci-bridge, etc.
}

// NodeDevicePCIAddress represents a PCI address.
type NodeDevicePCIAddress struct {
	Domain   *uint 
	Bus      *uint 
	Slot     *uint 
	Function *uint 
}

// NodeDeviceIOMMUGroup describes the IOMMU group a device belongs to.
type NodeDeviceIOMMUGroup struct {
	Number  int                
	Address []NodeDevicePCIAddress 
}

// NodeDeviceNUMA describes the NUMA node associated with a PCI device.
type NodeDeviceNUMA struct {
	Node int 
}

// NodeDevicePCIExpress describes PCI Express capabilities.
type NodeDevicePCIExpress struct {
	Links []NodeDevicePCIExpressLink 
}

// NodeDevicePCIExpressLink describes PCI Express link details.
type NodeDevicePCIExpressLink struct {
	Validity string   // e.g., "cap" or "sta"
	Speed    float64     // Speed in GigaTransfers per second
	Port     *uint   
	Width    *uint   
}

// NodeDevicePCISubCapability represents sub-capabilities within PCI.
type NodeDevicePCISubCapability struct {
	VirtFunctions *NodeDevicePCIVirtFunctionsCapability 
	PhysFunction  *NodeDevicePCIPhysFunctionCapability  
	MDevTypes     *NodeDevicePCIMDevTypesCapability     
	Bridge        *NodeDevicePCIBridgeCapability        
	VPD           *NodeDevicePCIVPDCapability           
}

// NodeDevicePCIVirtFunctionsCapability describes SRIOV Virtual Functions.
type NodeDevicePCIVirtFunctionsCapability struct {
	Address  []NodeDevicePCIAddress 
	MaxCount int                  
}

// NodeDevicePCIPhysFunctionCapability describes the Physical Function (PF) of an SRIOV device.
type NodeDevicePCIPhysFunctionCapability struct {
	Address NodeDevicePCIAddress 
}

// NodeDevicePCIMDevTypesCapability lists mediated device types supported by a PCI device.
type NodeDevicePCIMDevTypesCapability struct {
	Types []NodeDeviceMDevType 
}

// NodeDevicePCIBridgeCapability indicates if a device is a PCI bridge.
type NodeDevicePCIBridgeCapability struct{}

// NodeDevicePCIVPDCapability describes the VPD PCI/PCIe capability.
type NodeDevicePCIVPDCapability struct {
	Name      string                      
	ReadOnly  *NodeDevicePCIVPDFieldsRO   
	ReadWrite *NodeDevicePCIVPDFieldsRW   
}

// NodeDevicePCIVPDFieldsRO represents read-only VPD fields.
type NodeDevicePCIVPDFieldsRO struct {
	ChangeLevel   string                  
	ManufactureID string                  
	PartNumber    string                  
	SerialNumber  string                  
	VendorFields  []NodeDevicePCIVPDCustomField 
}

// NodeDevicePCIVPDFieldsRW represents read-write VPD fields.
type NodeDevicePCIVPDFieldsRW struct {
	AssetTag     string                      
	VendorFields []NodeDevicePCIVPDCustomField 
	SystemFields []NodeDevicePCIVPDCustomField 
}

// NodeDevicePCIVPDCustomField represents a custom VPD field.
type NodeDevicePCIVPDCustomField struct {
	Index string 
	Value string 
}
```

### `NodeDeviceUSBDeviceCapability`

Describes a USB device based on its location.

```go
// NodeDeviceUSBDeviceCapability describes a USB device based on its location.
type NodeDeviceUSBDeviceCapability struct {
	Bus     int        
	Device  int        
	Product NodeDeviceIDName 
	Vendor  NodeDeviceIDName 
}

// NodeDeviceIDName represents an ID and Name for a device.
type NodeDeviceIDName struct {
	ID   string 
	Name string 
}
```

### `NodeDeviceUSBCapability`

Describes a USB device based on its advertised driver interface.

```go
// NodeDeviceUSBCapability describes a USB device based on its advertised driver interface.
type NodeDeviceUSBCapability struct {
	Number      int    
	Class       int    
	Subclass    int    
	Protocol    int    
	Description string 
}
```

### `NodeDeviceNetCapability`

Describes a device capable for use as a network interface.

```go
// NodeDeviceNetCapability describes a device capable for use as a network interface.
type NodeDeviceNetCapability struct {
	Interface string                     
	Address   string                     
	Link      *NodeDeviceNetLink         
	Features  []NodeDeviceNetOffloadFeatures 
	Capability []NodeDeviceNetSubCapability  // e.g., 80203 or 80211
}

// NodeDeviceNetLink describes the link status of a network device.
type NodeDeviceNetLink struct {
	State string 
	Speed string 
}

// NodeDeviceNetOffloadFeatures lists hardware offloads supported by a network interface.
type NodeDeviceNetOffloadFeatures struct {
	Name string 
}

// NodeDeviceNetSubCapability represents a network protocol capability.
type NodeDeviceNetSubCapability struct {
	Wireless80211 *NodeDeviceNet80211Capability 
	Ethernet80203 *NodeDeviceNet80203Capability 
}
```

### `NodeDeviceSCSIHostCapability`

Describes a SCSI host device.

```go
// NodeDeviceSCSIHostCapability describes a SCSI host device.
type NodeDeviceSCSIHostCapability struct {
	Host       uint                         
	UniqueID   *uint                        
	Capability []NodeDeviceSCSIHostSubCapability 
}

// NodeDeviceSCSIHostSubCapability represents sub-capabilities of a SCSI host.
type NodeDeviceSCSIHostSubCapability struct {
	VPortOps *NodeDeviceSCSIVPortOpsCapability 
	FCHost   *NodeDeviceSCSIFCHostCapability   
}

// NodeDeviceSCSIFCHostCapability describes Fibre Channel host details.
type NodeDeviceSCSIFCHostCapability struct {
	WWNN      string 
	WWPN      string 
	FabricWWN string 
}

// NodeDeviceSCSIVPortOpsCapability indicates support for vPort operations.
type NodeDeviceSCSIVPortOpsCapability struct {
	VPorts    int 
	MaxVPorts int 
}
```

### `NodeDeviceSCSIHostSubCapability`

This is a nested type within `NodeDeviceSCSIHostCapability`.

### `NodeDeviceSCSIFCHostCapability`

This is a nested type within `NodeDeviceSCSIHostSubCapability`.

### `NodeDeviceSCSITargetCapability`

Describes a SCSI target device.

```go
// NodeDeviceSCSITargetCapability describes a SCSI target device.
type NodeDeviceSCSITargetCapability struct {
	Target     string                         
	Capability []NodeDeviceSCSITargetSubCapability 
}

// NodeDeviceSCSITargetSubCapability represents sub-capabilities of a SCSI target.
type NodeDeviceSCSITargetSubCapability struct {
	FCRemotePort *NodeDeviceSCSIFCRemotePortCapability 
}

// NodeDeviceSCSIFCRemotePortCapability describes Fibre Channel remote port details.
type NodeDeviceSCSIFCRemotePortCapability struct {
	RPort string 
	WWPN  string 
}
```

### `NodeDeviceSCSI`

Describes a SCSI device.

```go
// NodeDeviceSCSI describes a SCSI device.
type NodeDeviceSCSI struct {
	Host   int    
	Bus    int    
	Target int    
	Lun    int    
	Type   string 
}
```

### `NodeDeviceStorageCapability`

Describes a device usable for storage.

```go
// NodeDeviceStorageCapability describes a device usable for storage.
type NodeDeviceStorageCapability struct {
	Block            string                         
	Bus              string                         
	DriverType       string                         
	Model            string                         
	Vendor           string                         
	Serial           string                         
	Size             *uint                          
	LogicalBlockSize *uint                          
	NumBlocks        *uint                          
	Capability       []NodeDeviceStorageSubCapability 
}

// NodeDeviceStorageSubCapability represents sub-capabilities of a storage device.
type NodeDeviceStorageSubCapability struct {
	Removable *NodeDeviceStorageRemovableCapability 
}

// NodeDeviceStorageRemovableCapability describes capabilities of removable storage.
type NodeDeviceStorageRemovableCapability struct {
	MediaAvailable   *uint 
	MediaSize        *uint 
	MediaLabel       string 
	LogicalBlockSize *uint 
	NumBlocks        *uint 
}
```

### `NodeDeviceDRMCapability`

Describes a Direct Rendering Manager (DRM) device.

```go
// NodeDeviceDRMCapability describes a Direct Rendering Manager (DRM) device.
type NodeDeviceDRMCapability struct {
	Type string 
}
```

### `NodeDeviceMDevCapability`

Describes a mediated device.

```go
// NodeDeviceMDevCapability describes a mediated device.
type NodeDeviceMDevCapability struct {
	Type          *NodeDeviceMDevCapabilityType   
	IOMMUGroup    *NodeDeviceIOMMUGroup           
	UUID          string                          
	ParentAddr    string                          
	Attrs         []NodeDeviceMDevCapabilityAttrs 
}

// NodeDeviceMDevCapabilityType describes the type of a mediated device.
type NodeDeviceMDevCapabilityType struct {
	ID string 
}

// NodeDeviceMDevCapabilityAttrs represents vendor-specific attributes for a mediated device.
type NodeDeviceMDevCapabilityAttrs struct {
	Name  string 
	Value string 
}

// NodeDeviceMDevType describes a mediated device type.
type NodeDeviceMDevType struct {
	ID                 string 
	Name               string 
	DeviceAPI          string 
	AvailableInstances uint   
}
```

### `NodeDeviceCCWCapability`

Describes a Command Channel Word (CCW) device on the S390 architecture.

```go
// NodeDeviceCCWCapability describes a Command Channel Word (CCW) device.
type NodeDeviceCCWCapability struct {
	CSSID        *uint                      
	SSID         *uint                      
	DevNo        *uint                      
	Capabilities []NodeDeviceCCWSubCapability 
}

// NodeDeviceCCWSubCapability represents sub-capabilities for CCW devices.
type NodeDeviceCCWSubCapability struct {
	GroupMember *NodeDeviceCCWGroupMemberCapability 
}

// NodeDeviceCCWGroupCapability describes a CCW group.
type NodeDeviceCCWGroupCapability struct {
	State        string                           
	CSSID        *uint                            
	SSID         *uint                            
	DevNo        *uint                            
	Members      *NodeDeviceCCWGroupMembers       
	Capabilities []NodeDeviceCCWGroupSubCapability 
}

// NodeDeviceCCWGroupMembers lists members of a CCW group.
type NodeDeviceCCWGroupMembers struct {
	CCWDevice []NodeDeviceCCWGroupMembersDevice 
}

// NodeDeviceCCWGroupMembersDevice represents a CCW device member.
type NodeDeviceCCWGroupMembersDevice struct {
	Ref  string 
	Name string 
}

// NodeDeviceCCWGroupSubCapability represents sub-capabilities for CCW groups.
type NodeDeviceCCWGroupSubCapability struct {
	QEthGeneric *NodeDeviceCCWGroupSubCapabilityQEthGeneric 
}

// NodeDeviceCCWGroupSubCapabilityQEthGeneric describes QEthGeneric properties.
type NodeDeviceCCWGroupSubCapabilityQEthGeneric struct {
	CardType string 
	ChpID    string 
}

// NodeDeviceCCWSubCapability represents sub-capabilities for CCW devices.
type NodeDeviceCCWSubCapability struct {
	GroupMember *NodeDeviceCCWGroupMemberCapability 
}
```

### `NodeDeviceCSSCapability`

Describes a Channel SubSystem (CSS) device on the S390 architecture.

```go
// NodeDeviceCSSCapability describes a Channel SubSystem (CSS) device.
type NodeDeviceCSSCapability struct {
	CSSID        *uint                      
	SSID         *uint                      
	DevNo        *uint                      
	ChannelDevAddr *NodeDeviceCSSChannelDevAddr 
	Capabilities []NodeDeviceCSSSubCapability 
}

// NodeDeviceCSSChannelDevAddr describes the channel device address.
type NodeDeviceCSSChannelDevAddr struct {
	CSSID *uint 
	SSID  *uint 
	DevNo *uint 
}

// NodeDeviceCSSSubCapability represents sub-capabilities for CSS devices.
type NodeDeviceCSSSubCapability struct {
	MDevTypes *NodeDeviceCSSMDevTypesCapability 
}

// NodeDeviceCSSMDevTypesCapability lists mediated device types for CSS devices.
type NodeDeviceCSSMDevTypesCapability struct {
	Types []NodeDeviceMDevType 
}
```

## Mapping to Libvirt XML

The `NodeDevice` struct maps to the `<device>` element in libvirt XML.

*   The `Name` field maps to the `<name>` element.
*   The `Path` field maps to the `<path>` element.
*   The `Parent` field maps to the `<parent>` element.
*   The `Driver` field maps to the `<driver>` element.
*   The `DevNodes` field maps to a list of `<devnode>` elements.
*   The `Capability` field maps to a `<capability>` element, whose structure depends on the specific capability type (e.g., `system`, `pci`, `net`).

For example, the following Go struct for a PCI device:

```go
pciDevice := libvirtxml.NodeDevice{
	Name:   "pci_0000_02_00_0",
	Path:   "/sys/devices/pci0000:00/0000:00:04.0/0000:02:00.0",
	Parent: "pci_0000_00_04_0",
	Driver: &libvirtxml.NodeDeviceDriver{Name: "igb"},
	DevNodes: []libvirtxml.NodeDeviceDevNode{
		{Type: "dev", Path: "/dev/dri/by-path/pci-0000:02:00.0-render"},
	},
	Capability: libvirtxml.NodeDeviceCapability{
		PCI: &libvirtxml.NodeDevicePCICapability{
			Class:    "0x020000",
			Domain:   uint(0),
			Bus:      uint(2),
			Slot:     uint(0),
			Function: uint(0),
			Product:  libvirtxml.NodeDeviceIDName{ID: "0x10c9", Name: "82576 Gigabit Network Connection"},
			Vendor:   libvirtxml.NodeDeviceIDName{ID: "0x8086", Name: "Intel Corporation"},
			IOMMUGroup: &libvirtxml.NodeDeviceIOMMUGroup{
				Number: 12,
				Address: []libvirtxml.NodeDevicePCIAddress{
					{Domain: uint(0), Bus: uint(2), Slot: uint(0), Function: uint(0)},
					{Domain: uint(0), Bus: uint(2), Slot: uint(0), Function: uint(1)},
				},
			},
			PCIExpress: &libvirtxml.NodeDevicePCIExpress{
				Links: []libvirtxml.NodeDevicePCIExpressLink{
					{Validity: "cap", Speed: 2.5, Port: uint(1), Width: uint(1)},
					{Validity: "sta", Speed: 2.5, Width: uint(1)},
				},
			},
			Capabilities: []libvirtxml.NodeDevicePCISubCapability{
				{
					VirtFunctions: &libvirtxml.NodeDevicePCIVirtFunctionsCapability{
						Address: []libvirtxml.NodeDevicePCIAddress{
							{Domain: uint(0), Bus: uint(2), Slot: uint(0x10), Function: uint(0)},
							{Domain: uint(0), Bus: uint(2), Slot: uint(0x10), Function: uint(2)},
							{Domain: uint(0), Bus: uint(2), Slot: uint(0x10), Function: uint(4)},
							{Domain: uint(0), Bus: uint(2), Slot: uint(0x10), Function: uint(6)},
							{Domain: uint(0), Bus: uint(2), Slot: uint(0x11), Function: uint(0)},
							{Domain: uint(0), Bus: uint(2), Slot: uint(0x11), Function: uint(2)},
							{Domain: uint(0), Bus: uint(2), Slot: uint(0x11), Function: uint(4)},
						},
					},
				},
			},
		},
	},
}
```

Would correspond to the following libvirt XML:

```xml
<device>
  <name>pci_0000_02_00_0</name>
  <path>/sys/devices/pci0000:00/0000:00:04.0/0000:02:00.0</path>
  <parent>pci_0000_00_04_0</parent>
  <driver>
    <name>igb</name>
  </driver>
  <devnode type='dev'>/sys/devices/pci0000:00/0000:00:04.0/0000:02:00.0/net/eth1</devnode>
  <capability type='pci'>
    <class>0x020000</class>
    <domain>0</domain>
    <bus>2</bus>
    <slot>0</slot>
    <function>0</function>
    <product id='0x10c9'>82576 Gigabit Network Connection</product>
    <vendor id='0x8086'>Intel Corporation</vendor>
    <iommuGroup number='12'>
      <address domain='0x0000' bus='0x02' slot='0x00' function='0x0'/>
      <address domain='0x0000' bus='0x02' slot='0x00' function='0x1'/>
    </iommuGroup>
    <pci-express>
      <link validity='cap' port='1' speed='2.5' width='1'/>
      <link validity='sta' speed='2.5' width='1'/>
    </pci-express>
    <capability type='virt_functions'>
      <address domain='0x0000' bus='0x02' slot='0x10' function='0x0'/>
      <address domain='0x0000' bus='0x02' slot='0x10' function='0x2'/>
      <address domain='0x0000' bus='0x02' slot='0x10' function='0x4'/>
      <address domain='0x0000' bus='0x02' slot='0x10' function='0x6'/>
      <address domain='0x0000' bus='0x02' slot='0x11' function='0x0'/>
      <address domain='0x0000' bus='0x02' slot='0x11' function='0x2'/>
      <address domain='0x0000' bus='0x02' slot='0x11' function='0x4'/>
    </capability>
  </capability>
</device>
