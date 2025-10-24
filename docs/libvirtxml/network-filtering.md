```go
# libvirtxml.NWFilter

The `NWFilter` struct in `libvirt.org/go/libvirtxml` represents the XML configuration for libvirt network filters. These filters are used to define and manage network traffic filtering rules for virtual machines. This document details the `NWFilter` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Network Filters XML format](https://libvirt.org/formatnwfilter.html).

## `NWFilter` Struct Definition

```go
// NWFilter represents the XML configuration for libvirt network filters.
type NWFilter struct {
	XMLName  xml.Name  // XMLName is the XML element name, typically "filter".
	Name     string    // Name of the network filter.
	UUID     string    // UUID of the network filter.
	Chain    string    // Chain type (e.g., "ipv4", "ipv6").
	Priority int       // Priority of the filter.
	Entries  []NWFilterEntry  // List of filter entries (rules or references).
}
```

## Nested Structs

```go
// NWFilterEntry represents a single entry within a network filter, which can be a rule or a reference to another filter.
type NWFilterEntry struct {
	Rule *NWFilterRule
	Ref  *NWFilterRef
}

// NWFilterRule represents a single network traffic filtering rule.
type NWFilterRule struct {
	Action     string
	Direction  string
	Priority   int
	StateMatch string
	ARP        *NWFilterRuleARP
	RARP       *NWFilterRuleRARP
	MAC        *NWFilterRuleMAC
	VLAN       *NWFilterRuleVLAN
	STP        *NWFilterRuleSTP
	IP         *NWFilterRuleIP
	IPv6       *NWFilterRuleIPv6
	TCP        *NWFilterRuleTCP
	UDP        *NWFilterRuleUDP
	UDPLite    *NWFilterRuleUDPLite
	ESP        *NWFilterRuleESP
	AH         *NWFilterRuleAH
	SCTP       *NWFilterRuleSCTP
	ICMP       *NWFilterRuleICMP
	All        *NWFilterRuleAll
	IGMP       *NWFilterRuleIGMP
	TCPIPv6    *NWFilterRuleTCPIPv6
	UDPIPv6    *NWFilterRuleUDPIPv6
	UDPLiteIPv6 *NWFilterRuleUDPLiteIPv6
	ESPIPv6    *NWFilterRuleESPIPv6
	AHIPv6     *NWFilterRuleAHIPv6
	SCTPIPv6   *NWFilterRuleSCTPIPv6
	ICMPv6     *NWFilterRuleICMPIPv6
	AllIPv6    *NWFilterRuleAllIPv6
	Comment    string
}

// NWFilterRef references another network filter.
type NWFilterRef struct {
	Filter     string
	Parameters []NWFilterBindingFilterParam
}

// NWFilterParameter represents a parameter for a referenced filter.
type NWFilterParameter struct {
	Name  string
	Value string
}

// NWFilterField represents a field value that can be a string, integer, or MAC/IP address.
type NWFilterField struct {
	Var  string // Variable name (e.g., "$MAC", "$IP").
	Str  string // String value.
	Uint *uint  // Unsigned integer value.
}

// NWFilterRuleARP represents an ARP/RARP filtering rule.
type NWFilterRuleARP struct {
	Match         string
	HWType        NWFilterField
	ProtocolType  NWFilterField
	OpCode        NWFilterField
	ARPSrcMACAddr NWFilterField
	ARPDstMACAddr NWFilterField
	ARPSrcIPAddr  NWFilterField
	ARPSrcIPMask  NWFilterField
	ARPDstIPAddr  NWFilterField
	ARPDstIPMask  NWFilterField
	Gratuitous    NWFilterField
	Comment       string
}

// NWFilterRuleMAC represents a MAC address filtering rule.
type NWFilterRuleMAC struct {
	Match        string
	ProtocolID   NWFilterField
	Comment      string
}

// NWFilterRuleVLAN represents a VLAN filtering rule.
type NWFilterRuleVLAN struct {
	Match           string
	VLANID          NWFilterField
	EncapProtocol   NWFilterField
	Comment         string
}

// NWFilterRuleIP represents an IPv4 filtering rule.
type NWFilterRuleIP struct {
	Match       string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Protocol    NWFilterField
	DSCP        NWFilterField
	Comment     string
}

// NWFilterRuleIPv6 represents an IPv6 filtering rule.
type NWFilterRuleIPv6 struct {
	Match       string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Protocol    NWFilterField
	Type        NWFilterField
	TypeEnd     NWFilterField
	Code        NWFilterField
	CodeEnd     NWFilterField
	Comment     string
}

// NWFilterRuleTCP represents a TCP filtering rule.
type NWFilterRuleTCP struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	SrcPortStart NWFilterField
	SrcPortEnd   NWFilterField
	DstPortStart NWFilterField
	DstPortEnd   NWFilterField
	Option       NWFilterField
	Flags        NWFilterField
	Comment      string
}

// NWFilterRuleUDP represents a UDP filtering rule.
type NWFilterRuleUDP struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	SrcPortStart NWFilterField
	SrcPortEnd   NWFilterField
	DstPortStart NWFilterField
	DstPortEnd   NWFilterField
	Comment  string
}

// NWFilterRuleUDPLite represents a UDPLite filtering rule.
type NWFilterRuleUDPLite struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment  string
}

// NWFilterRuleESP represents an ESP filtering rule.
type NWFilterRuleESP struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}

// NWFilterRuleAH represents an AH filtering rule.
type NWFilterRuleAH struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}

// NWFilterRuleSCTP represents an SCTP filtering rule.
type NWFilterRuleSCTP struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	SrcPortStart NWFilterField
	SrcPortEnd   NWFilterField
	DstPortStart NWFilterField
	DstPortEnd   NWFilterField
	Comment  string
}

// NWFilterRuleICMP represents an ICMP filtering rule.
type NWFilterRuleICMP struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Type        NWFilterField
	Code        NWFilterField
	Comment     string
}

// NWFilterRuleAll represents a filtering rule for all protocols.
type NWFilterRuleAll struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}

// NWFilterRuleIGMP represents an IGMP filtering rule.
type NWFilterRuleIGMP struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}

// NWFilterRuleTCPIPv6 represents a TCP over IPv6 filtering rule.
type NWFilterRuleTCPIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Option       NWFilterField
	Comment      string
}

// NWFilterRuleUDPIPv6 represents a UDP over IPv6 filtering rule.
type NWFilterRuleUDPIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment  string
}

// NWFilterRuleUDPLiteIPv6 represents a UDPLite over IPv6 filtering rule.
type NWFilterRuleUDPLiteIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}

// NWFilterRuleESPIPv6 represents an ESP over IPv6 filtering rule.
type NWFilterRuleESPIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}

// NWFilterRuleAHIPv6 represents an AH over IPv6 filtering rule.
type NWFilterRuleAHIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}

// NWFilterRuleSCTPIPv6 represents an SCTP over IPv6 filtering rule.
type NWFilterRuleSCTPIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	SrcPortStart NWFilterField
	SrcPortEnd   NWFilterField
	DstPortStart NWFilterField
	DstPortEnd   NWFilterField
	Comment  string
}

// NWFilterRuleICMPIPv6 represents an ICMPv6 filtering rule.
type NWFilterRuleICMPIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Type        NWFilterField
	Code        NWFilterField
	Comment     string
}

// NWFilterRuleAllIPv6 represents a filtering rule for all protocols over IPv6.
type NWFilterRuleAllIPv6 struct {
	Match    string
	SrcMACAddr  NWFilterField
	SrcIPAddr   NWFilterField
	SrcIPMask   NWFilterField
	DstIPAddr   NWFilterField
	DstIPMask   NWFilterField
	Comment string
}
```
