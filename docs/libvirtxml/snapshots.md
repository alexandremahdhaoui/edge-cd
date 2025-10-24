# libvirtxml.DomainSnapshot

The `DomainSnapshot` struct in `libvirt.org/go/libvirtxml` represents a snapshot of a virtual machine's state, which can include disk states and memory state. This document details the `DomainSnapshot` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Snapshot XML format](https://libvirt.org/formatsnapshot.html).

## `DomainSnapshot` Struct Definition

```go
// DomainSnapshot represents a snapshot of a virtual machine's state.
type DomainSnapshot struct {
	XMLName        xml.Name
	Name           string                
	Description    string                
	State          string                
	CreationTime   string                
	Parent         *DomainSnapshotParent 
	Memory         *DomainSnapshotMemory 
	Disks          *DomainSnapshotDisks  
	Domain         *Domain                // Represents the domain configuration at the time of the snapshot.
	InactiveDomain *DomainSnapshotInactiveDomain  // Represents the inactive domain configuration.
	Active         *uint                 
	Cookie         *DomainSnapshotCookie 
}
```

## Nested Structs

### `DomainSnapshotParent`

The `DomainSnapshotParent` struct indicates the parent snapshot in a hierarchical snapshot tree.

```go
// DomainSnapshotParent indicates the parent snapshot in a hierarchical snapshot tree.
type DomainSnapshotParent struct {
	Name string 
}
```

### `DomainSnapshotMemory`

The `DomainSnapshotMemory` struct specifies how the VM memory state is handled during snapshotting.

```go
// DomainSnapshotMemory specifies how the VM memory state is handled during snapshotting.
type DomainSnapshotMemory struct {
	Snapshot string  // e.g., "no", "internal", "external"
	File     string  // Path to the memory state file if snapshot="external"
}
```

### `DomainSnapshotDisks`

The `DomainSnapshotDisks` struct contains a list of disk snapshots.

```go
// DomainSnapshotDisks contains a list of disk snapshots.
type DomainSnapshotDisks struct {
	Disks []DomainSnapshotDisk 
}
```

### `DomainSnapshotDisk`

The `DomainSnapshotDisk` struct describes the snapshot properties for a specific disk.

```go
// DomainSnapshotDisk describes the snapshot properties for a specific disk.
type DomainSnapshotDisk struct {
	Name       string         // Matches the disk's target device name or source file.
	Snapshot   string         // Snapshot mode: "no", "internal", "external", "bitmap", "manual".
	Driver     *DomainDiskDriver  // Specifies the driver type for the snapshot file (e.g., qcow2).
	Source     *DomainDiskSource  // Specifies the source file or path for the snapshot.
}
```

### `DomainSnapshotInactiveDomain`

The `DomainSnapshotInactiveDomain` struct represents the domain configuration at the time the snapshot was taken, specifically for an inactive domain.

```go
// DomainSnapshotInactiveDomain represents the domain configuration at the time of the snapshot.
type DomainSnapshotInactiveDomain struct {
	Domain // Embeds the Domain struct to represent the full domain configuration.
}
```

### `DomainSnapshotCookie`

The `DomainSnapshotCookie` struct holds arbitrary data that libvirt might need to restore a domain from an active snapshot.

```go
// DomainSnapshotCookie holds arbitrary data for restoring a domain from an active snapshot.
type DomainSnapshotCookie struct {
	XML string  // Contains the raw XML data for the cookie.
}
```

## Mapping to Libvirt XML

The `DomainSnapshot` struct corresponds to the `<domainsnapshot>` element in libvirt XML.

*   **`Name`**: Maps to the `<name>` element.
*   **`Description`**: Maps to the `<description>` element.
*   **`State`**: Maps to the `<state>` element (read-only).
*   **`CreationTime`**: Maps to the `<creationTime>` element (read-only).
*   **`Parent`**: Maps to the `<parent>` element, containing a `<name>` sub-element.
*   **`Memory`**: Maps to the `<memory>` element, with its `Snapshot` field mapping to the `snapshot` attribute.
*   **`Disks`**: Maps to the `<disks>` element, containing a list of `<disk>` elements. Each `DomainSnapshotDisk` maps to a `<disk>` element, with `Name` mapping to the `name` attribute, `Snapshot` to the `snapshot` attribute, `Driver` to the `<driver>` element, and `Source` to the `<source>` element.
*   **`Domain`**: Maps to the `<domain>` element, representing the domain configuration at the time of the snapshot.
*   **`InactiveDomain`**: Maps to the `<inactiveDomain>` element.
*   **`Active`**: Maps to the `<active>` element.
*   **`Cookie`**: Maps to the `<cookie>` element.

For example, the following Go struct for a disk snapshot:

```go
diskSnapshot := libvirtxml.DomainSnapshot{
	Name:        "snapshot-123",
	Description: "Snapshot after OS install",
	Disks: &libvirtxml.DomainSnapshotDisks{
		Disks: []libvirtxml.DomainSnapshotDisk{
			{
				Name:     "vda",
				Snapshot: "external",
				Driver: &libvirtxml.DomainDiskDriver{
					Type: "qcow2",
				},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: "/path/to/snapshot/vda.qcow2",
					},
				},
			},
			{
				Name:     "vdb",
				Snapshot: "no",
			},
		},
	},
}
```

Would correspond to the following libvirt XML:

```xml
<domainsnapshot>
  <name>snapshot-123</name>
  <description>Snapshot after OS install</description>
  <disks>
    <disk name='vda' snapshot='external'>
      <driver type='qcow2'/>
      <source file='/path/to/snapshot/vda.qcow2'/>
    </disk>
    <disk name='vdb' snapshot='no'/>
  </disks>
</domainsnapshot>
