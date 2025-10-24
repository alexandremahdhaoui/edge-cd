# libvirtxml.NetworkPort

The `NetworkPort` struct in `libvirt.org/go/libvirtxml` represents the XML configuration for a libvirt network port. It defines the connection between a virtual interface of a virtual domain and the virtual network it is attached to. This document details the `NetworkPort` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Network Port XML format](https://libvirt.org/formatnetworkport.html).

## `NetworkPort` Struct Definition

```go
// NetworkPort represents the XML configuration for a libvirt network port.
type NetworkPort struct {
	XMLName     xml.Name
	UUID        string // The globally unique identifier for the virtual network port. RFC 4122 compliant. If omitted, a random UUID is generated.
	Owner       *NetworkPortOwner // Records the domain object that owns the network port. It contains the domain's UUID and Name.
	MAC         *NetworkPortMAC // Defines the MAC address for a network port. The address attribute provides the MAC address of the virtual port that will be seen by the guest. It must not start with 0xFE.
	Group       string // The port group in the virtual network to which the port belongs. Can be omitted if no port groups are defined on the network.
	Bandwidth   *NetworkBandwidth // Configures Quality of Service settings for the network port. Incoming and outgoing traffic can be shaped independently.
	VLAN        *NetworkPortVLAN // Configures VLAN tagging for the port.
	PortOptions *NetworkPortPortOptions // Configures port isolation. The 'isolated' property, when set to 'yes', isolates this port's network traffic from other ports on the same network.
	VirtualPort *NetworkVirtualPort // Describes metadata for virtual port configuration.
	RXFilters   *NetworkPortRXFilters // Configures receive filters for the port. The 'trustGuest' property allows the host to trust reports from the guest regarding changes to the interface MAC address and receive filters.
	Plug        *NetworkPortPlug // Describes how a port is plugged into the network, specifying the connection type (e.g., network, bridge, direct, hostdev-pci).
}
```

## Nested Structs

```go
// NetworkPortOwner records the domain object that owns the network port.
type NetworkPortOwner struct {
	UUID string // The globally unique identifier for the virtual domain.
	Name string // The unique name of the virtual domain.
}

// NetworkPortMAC defines the MAC address for a network port.
type NetworkPortMAC struct {
	Address string // The MAC address of the virtual port that will be seen by the guest. Must not start with 0xFE.
}

// NetworkPortGroup defines a group of network ports with common configurations.
type NetworkPortGroup struct {
	Name                string // The name of the port group.
	Default             string // Indicates if this is the default port group.
	TrustGuestRxFilters string // Trust reports from the guest regarding changes to the interface MAC address and receive filters. Supported for virtio and macvtap connections.
	VLAN                *NetworkPortVLAN // Configures VLAN tagging for the port group.
	VirtualPort         *NetworkVirtualPort // Describes metadata for virtual port configuration within the port group.
}

// NetworkPortPortOptions configures port isolation.
type NetworkPortPortOptions struct {
	Isolated string // If set to 'yes', isolates this port's network traffic from other ports on the same network. Only supported for emulated network devices connected to a Linux host bridge via a standard tap device. Default is 'no'.
}

// NetworkPortRXFilters configures receive filters for the port.
type NetworkPortRXFilters struct {
	TrustGuest string // Trust reports from the guest regarding changes to the interface MAC address and receive filters.
}

// NetworkPortVLAN configures VLAN tagging for the port.
type NetworkPortVLAN struct {
	Trunk string // Specifies if the port is part of a trunk.
	Tags  []NetworkPortVLANTag // A list of VLAN tags configured for the port.
}

// NetworkPortVLANTag configures a specific VLAN tag.
type NetworkPortVLANTag struct {
	ID         uint   // The VLAN ID.
	NativeMode string // Specifies if this VLAN tag is used in native mode.
}

// NetworkPortPlug describes how a port is plugged into the network.
type NetworkPortPlug struct {
	Bridge     *NetworkPortPlugBridge // Describes plugging into an external bridge.
	Network    *NetworkPortPlugNetwork // Describes plugging into a libvirt-managed network.
	Direct     *NetworkPortPlugDirect // Describes a direct connection to a physical interface.
	HostDevPCI *NetworkPortPlugHostDevPCI // Describes passthrough of a PCI device.
}

// NetworkPortPlugBridge describes plugging into an external bridge.
type NetworkPortPlugBridge struct {
	Bridge          string // The name of the externally managed bridge device.
	MacTableManager string // Specifies the MAC table manager for the bridge.
}

// NetworkPortPlugNetwork describes plugging into a libvirt-managed network.
type NetworkPortPlugNetwork struct {
	Bridge          string // The name of the privately managed bridge device associated with the virtual network.
	MacTableManager string // Specifies the MAC table manager for the bridge.
}

// NetworkPortPlugDirect describes a direct connection to a physical interface.
type NetworkPortPlugDirect struct {
	Dev  string // The name of the physical network interface to which the port will be connected.
	Mode string // Describes how the connection will be set up (e.g., 'vepa').
}

// NetworkPortPlugHostDevPCI describes passthrough of a PCI device.
type NetworkPortPlugHostDevPCI struct {
	Managed string // Indicates who is responsible for managing the PCI device (e.g., 'yes' for libvirt).
	Driver  *NetworkPortPlugHostDevPCIDriver // Specifies driver options for the PCI device.
	Address *NetworkPortPlugHostDevPCIAddress // Specifies the PCI address details.
}

// NetworkPortPlugHostDevPCIAddress specifies PCI address details.
type NetworkPortPlugHostDevPCIAddress struct {
	Domain   *uint // The PCI domain address.
	Bus      *uint // The PCI bus address.
	Slot     *uint // The PCI slot address.
	Function *uint // The PCI function address.
}

// NetworkPortPlugHostDevPCIDriver specifies driver options for a PCI device.
type NetworkPortPlugHostDevPCIDriver struct {
	Name string // The name of the driver.
}
```

## Libvirt Network Port XML Format Reference

*   [https://libvirt.org/formatnetworkport.html](https://libvirt.org/formatnetworkport.html)
