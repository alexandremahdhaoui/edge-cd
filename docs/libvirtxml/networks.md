# libvirtxml.Network

The `Network` struct in `libvirt.org/go/libvirtxml` represents the XML configuration for a libvirt network. It is used to define, create, and manage virtual networks. This document details the `Network` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Network XML format](https://libvirt.org/formatnetwork.html).

## `Network` Struct Definition

```go
// Network represents the XML configuration for a libvirt network.
type Network struct {
	XMLName             xml.Name
	IPv6                string
	TrustGuestRxFilters string
	Name                string
	UUID                string
	Metadata            *NetworkMetadata
	Forward             *NetworkForward
	Bridge              *NetworkBridge
	MTU                 *NetworkMTU
	MAC                 *NetworkMAC
	Domain              *NetworkDomain
	DNS                 *NetworkDNS
	VLAN                *NetworkVLAN
	Bandwidth           *NetworkBandwidth
	PortOptions         *NetworkPortOptions
	IPs                 []NetworkIP
	Routes              []NetworkRoute
	VirtualPort         *NetworkVirtualPort
	PortGroups          []NetworkPortGroup
	DnsmasqOptions      *NetworkDnsmasqOptions
}
```

## Nested Structs

```go
// NetworkMetadata represents general metadata about the virtual network.
type NetworkMetadata struct {
	XML string
}

// NetworkForward controls how a virtual network is provided connectivity to the physical LAN.
type NetworkForward struct {
	Mode       string
	Dev        string
	Managed    string
	Driver     *NetworkForwardDriver
	PFs        []NetworkForwardPF
	NAT        *NetworkForwardNAT
	Interfaces []NetworkForwardInterface
	Addresses  []NetworkForwardAddress
}

// NetworkBridge defines the host bridge device for the virtual network.
type NetworkBridge struct {
	Name            string
	STP             string
	Delay           string
	MacTableManager string
	Zone            string
}

// NetworkMTU specifies the Maximum Transmission Unit for the network.
type NetworkMTU struct {
	Size uint
}

// NetworkMAC defines the MAC address for the virtual network's bridge.
type NetworkMAC struct {
	Address string
}

// NetworkDomain defines the DNS domain of the DHCP server.
type NetworkDomain struct {
	Name      string
	LocalOnly string
	Register  string
}

// NetworkDNS configures DNS services for the virtual network.
type NetworkDNS struct {
	Enable            string
	ForwardPlainNames string
	Forwarders        []NetworkDNSForwarder
	TXTs              []NetworkDNSTXT
	Host              []NetworkDNSHost
	SRVs              []NetworkDNSSRV
}

// NetworkVLAN configures VLAN tagging for the network.
type NetworkVLAN struct {
	Trunk string
	Tags  []NetworkVLANTag
}

// NetworkBandwidth configures Quality of Service settings for the network.
type NetworkBandwidth struct {
	Inbound  *NetworkBandwidthParams
	Outbound *NetworkBandwidthParams
}

// NetworkBandwidthParams specifies average, peak, burst, and floor for traffic shaping.
type NetworkBandwidthParams struct {
	Average *uint
	Peak    *uint
	Burst   *uint
	Floor   *uint
}

// NetworkPortOptions configures port isolation.
type NetworkPortOptions struct {
	Isolated string
}

// NetworkIP defines an IP address configuration for the network.
type NetworkIP struct {
	Address  string
	Family   string
	Netmask  string
	Prefix   uint
	LocalPtr string
	DHCP     *NetworkDHCP
	TFTP     *NetworkTFTP
}

// NetworkRoute defines a static route for the network.
type NetworkRoute struct {
	Family  string
	Address string
	Netmask string
	Prefix  uint
	Gateway string
	Metric  string
}

// NetworkVirtualPort describes metadata for virtual port configuration.
type NetworkVirtualPort struct {
	Params *NetworkVirtualPortParams
}

// NetworkPortGroup defines a group of network ports with common configurations.
type NetworkPortGroup struct {
	Name                string
	Default             string
	TrustGuestRxFilters string
	VLAN                *NetworkVLAN
	VirtualPort         *NetworkVirtualPort
	Bandwidth           *NetworkBandwidth
}

// NetworkDnsmasqOptions provides options for the dnsmasq configuration.
type NetworkDnsmasqOptions struct {
	XMLName xml.Name
	Option  []NetworkDnsmasqOption
}

// NetworkDnsmasqOption represents a single dnsmasq option.
type NetworkDnsmasqOption struct {
	Value string
}

// NetworkDNSForwarder specifies an alternate DNS server for forwarding requests.
type NetworkDNSForwarder struct {
	Domain string
	Addr   string
}

// NetworkDNSHost defines a DNS host entry.
type NetworkDNSHost struct {
	XMLName   xml.Name
	IP        string
	Hostnames []NetworkDNSHostHostname
}

// NetworkDNSHostHostname is a hostname for a DNS host entry.
type NetworkDNSHostHostname struct {
	Hostname string
}

// NetworkDNSSRV defines a DNS SRV record.
type NetworkDNSSRV struct {
	XMLName  xml.Name
	Service  string
	Protocol string
	Target   string
	Port     uint
	Priority uint
	Weight   uint
	Domain   string
}

// NetworkDNSTXT defines a DNS TXT record.
type NetworkDNSTXT struct {
	XMLName xml.Name
	Name    string
	Value   string
}

// NetworkForwardDriver specifies driver-specific options for forwarding.
type NetworkForwardDriver struct {
	Name  string
	Model string
}

// NetworkForwardInterface specifies a physical interface for direct connection modes.
type NetworkForwardInterface struct {
	XMLName xml.Name
	Dev     string
}

// NetworkForwardNAT configures NAT settings for the network.
type NetworkForwardNAT struct {
	IPv6      string
	Addresses []NetworkForwardNATAddress
	Ports     []NetworkForwardNATPort
}

// NetworkForwardNATAddress defines an IPv4 address range for NAT.
type NetworkForwardNATAddress struct {
	Start string
	End   string
}

// NetworkForwardNATPort defines a port range for NAT.
type NetworkForwardNATPort struct {
	Start uint
	End   uint
}

// NetworkForwardPF specifies a physical function for passthrough mode.
type NetworkForwardPF struct {
	Dev string
}

// NetworkDHCP configures DHCP services for the network.
type NetworkDHCP struct {
	Ranges []NetworkDHCPRange
	Hosts  []NetworkDHCPHost
	Bootp  []NetworkBootp
}

// NetworkDHCPHost defines a static DHCP host entry.
type NetworkDHCPHost struct {
	XMLName   xml.Name
	ID        string
	MAC       string
	Name      string
	IP        string
	Lease     *NetworkDHCPLease
}

// NetworkDHCPLease defines the lease time for a DHCP entry.
type NetworkDHCPLease struct {
	Expiry uint
	Unit   string
}

// NetworkDHCPRange defines an IP address range for DHCP.
type NetworkDHCPRange struct {
	XMLName xml.Name
	Start   string
	End     string
	Lease   *NetworkDHCPLease
}

// NetworkBootp specifies BOOTP options for DHCP.
type NetworkBootp struct {
	File   string
	Server string
}

// NetworkForwardAddress defines an address for NAT forwarding.
type NetworkForwardAddress struct {
	PCI *NetworkForwardAddressPCI
}

// NetworkForwardAddressPCI specifies PCI address details for forwarding.
type NetworkForwardAddressPCI struct {
	Domain   *uint
	Bus      *uint
	Slot     *uint
	Function *uint
}

// NetworkTFTP specifies TFTP services for the network.
type NetworkTFTP struct {
	Root string
}

// NetworkVLANTag configures a specific VLAN tag.
type NetworkVLANTag struct {
	ID         uint
	NativeMode string
}

// NetworkVirtualPortParams holds parameters for various virtual port types.
type NetworkVirtualPortParams struct {
	Any          *NetworkVirtualPortParamsAny
	VEPA8021QBG  *NetworkVirtualPortParamsVEPA8021QBG
	VNTag8011QBH *NetworkVirtualPortParamsVNTag8021QBH
	OpenVSwitch  *NetworkVirtualPortParamsOpenVSwitch
	MidoNet      *NetworkVirtualPortParamsMidoNet
}

// NetworkVirtualPortParamsAny represents generic virtual port parameters.
type NetworkVirtualPortParamsAny struct {
	ManagerID     *uint
	TypeID        *uint
	TypeIDVersion *uint
	InstanceID    string
	ProfileID     string
	InterfaceID   string
}

// NetworkVirtualPortParamsVEPA8021QBG represents 802.1Qbg virtual port parameters.
type NetworkVirtualPortParamsVEPA8021QBG struct {
	ManagerID     *uint
	TypeID        *uint
	TypeIDVersion *uint
	InstanceID    string
}

// NetworkVirtualPortParamsVNTag8021QBH represents 802.1Qbh virtual port parameters.
type NetworkVirtualPortParamsVNTag8021QBH struct {
	ProfileID string
}

// NetworkVirtualPortParamsOpenVSwitch represents Open vSwitch virtual port parameters.
type NetworkVirtualPortParamsOpenVSwitch struct {
	InterfaceID string
	ProfileID   string
}

// NetworkVirtualPortParamsMidoNet represents Midonet virtual port parameters.
type NetworkVirtualPortParamsMidoNet struct {
	InterfaceID string
}

// NetworkPortMAC defines the MAC address for a network port.
type NetworkPortMAC struct {
	Address string
}

// NetworkPortOwner records the domain object that owns the network port.
type NetworkPortOwner struct {
	UUID string
	Name string
}

// NetworkPortPlug describes how a port is plugged into the network.
type NetworkPortPlug struct {
	Bridge     *NetworkPortPlugBridge
	Network    *NetworkPortPlugNetwork
	Direct     *NetworkPortPlugDirect
	HostDevPCI *NetworkPortPlugHostDevPCI
}

// NetworkPortPlugBridge describes plugging into an external bridge.
type NetworkPortPlugBridge struct {
	Bridge          string
	MacTableManager string
}

// NetworkPortPlugDirect describes a direct connection to a physical interface.
type NetworkPortPlugDirect struct {
	Dev  string
	Mode string
}

// NetworkPortPlugHostDevPCI describes passthrough of a PCI device.
type NetworkPortPlugHostDevPCI struct {
	Managed string
	Driver  *NetworkPortPlugHostDevPCIDriver
	Address *NetworkPortPlugHostDevPCIAddress
}

// NetworkPortPlugHostDevPCIAddress specifies PCI address details.
type NetworkPortPlugHostDevPCIAddress struct {
	Domain   *uint
	Bus      *uint
	Slot     *uint
	Function *uint
}

// NetworkPortPlugHostDevPCIDriver specifies driver options for a PCI device.
type NetworkPortPlugHostDevPCIDriver struct {
	Name string
}

// NetworkPortPlugNetwork describes plugging into a libvirt-managed network.
type NetworkPortPlugNetwork struct {
	Bridge          string
	MacTableManager string
}

// NetworkPortRXFilters configures receive filters for the port.
type NetworkPortRXFilters struct {
	TrustGuest string
}

// NetworkPortVLAN configures VLAN tagging for the port.
type NetworkPortVLAN struct {
	Trunk string
	Tags  []NetworkVLANTag
}

// NetworkPortVLANTag configures a specific VLAN tag.
type NetworkPortVLANTag struct {
	ID         uint
	NativeMode string
}
```