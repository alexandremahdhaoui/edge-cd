# libvirtxml.DomainCheckpoint

The `DomainCheckpoint` struct in `libvirt.org/go/libvirtxml` represents a checkpoint of a virtual machine's disk state, used for incremental backups. This document details the `DomainCheckpoint` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Checkpoint XML format](https://libvirt.org/formatcheckpoint.html).

## `DomainCheckpoint` Struct Definition

```go
// DomainCheckpoint represents a checkpoint of a virtual machine's disk state.
type DomainCheckpoint struct {
	XMLName      xml.Name                 // XMLName is the XML element name, typically "domaincheckpoint".
	Name         string                   // Name of the checkpoint.
	Description  string                   // Human-readable description of the checkpoint.
	State        string                   // Current state of the checkpoint (read-only).
	CreationTime string                   // Timestamp when the checkpoint was created (read-only).
	Parent       *DomainCheckpointParent  // Parent checkpoint in a hierarchical tree.
	Disks        *DomainCheckpointDisks   // List of disk states at the time of the checkpoint.
	Domain       *Domain                  // Domain configuration at the time of the checkpoint.
}
```

## Nested Structs

### `DomainCheckpointParent`

The `DomainCheckpointParent` struct indicates the parent checkpoint in a hierarchical checkpoint tree.

```go
// DomainCheckpointParent indicates the parent checkpoint in a hierarchical checkpoint tree.
type DomainCheckpointParent struct {
	Name string 
}
```

### `DomainCheckpointDisks`

The `DomainCheckpointDisks` struct contains a list of disk checkpoints.

```go
// DomainCheckpointDisks contains a list of disk checkpoints.
type DomainCheckpointDisks struct {
	Disks []DomainCheckpointDisk 
}
```

### `DomainCheckpointDisk`

The `DomainCheckpointDisk` struct describes the checkpoint properties for a specific disk.

```go
// DomainCheckpointDisk describes the checkpoint properties for a specific disk.
type DomainCheckpointDisk struct {
	Name       string         // Matches the disk's target device name or source file.
	Checkpoint string         // Snapshot mode: "no", "bitmap".
	Bitmap     string         // Name of the tracking bitmap if checkpoint='bitmap'.
	Size       uint64         // Estimated size of changes since checkpoint (output only).
	Driver     *DomainDiskDriver  // Specifies the driver type for the disk.
	Source     *DomainDiskSource  // Specifies the source file or path for the disk.
}
```

### `Domain`

The `Domain` struct represents the domain configuration at the time the checkpoint was created. This is embedded here to show its relation, but its full definition is extensive and found in `domain.go`.

### `DomainCheckpointDiskDriver`

The `DomainDiskDriver` struct, relevant here for the disk's driver properties during checkpointing.

```go
// DomainDiskDriver specifies the driver properties for a disk.
type DomainDiskDriver struct {
	Name string 
	Type string 
	// ... other fields
}
```

### `DomainDiskSource`

The `DomainDiskSource` struct, relevant here for specifying the source of the disk being checkpointed.

```go
// DomainDiskSource specifies the source of a disk.
type DomainDiskSource struct {
	File *DomainDiskSourceFile 
	// ... other source types
}

// DomainDiskSourceFile specifies a disk source from a file.
type DomainDiskSourceFile struct {
	File string 
}
```

## Mapping to Libvirt XML

The `DomainCheckpoint` struct corresponds to the `<domaincheckpoint>` element in libvirt XML.

*   **`Name`**: Maps to the `<name>` element.
*   **`Description`**: Maps to the `<description>` element.
*   **`State`**: Maps to the `<state>` element (read-only).
*   **`CreationTime`**: Maps to the `<creationTime>` element (read-only).
*   **`Parent`**: Maps to the `<parent>` element, containing a `<name>` sub-element.
*   **`Disks`**: Maps to the `<disks>` element, containing a list of `<disk>` elements. Each `DomainCheckpointDisk` maps to a `<disk>` element, with `Name` mapping to the `name` attribute, `Checkpoint` to the `checkpoint` attribute, `Bitmap` to the `bitmap` attribute, `Size` to the `size` attribute, `Driver` to the `<driver>` element, and `Source` to the `<source>` element.
*   **`Domain`**: Maps to the `<domain>` element, representing the domain configuration at the time of the checkpoint.

For example, the following Go struct for a disk checkpoint:

```go
diskCheckpoint := libvirtxml.DomainCheckpoint{
	Name:        "checkpoint-456",
	Description: "Checkpoint after updates",
	Disks: &libvirtxml.DomainCheckpointDisks{
		Disks: []libvirtxml.DomainCheckpointDisk{
			{
				Name:       "vda",
				Checkpoint: "bitmap",
				Bitmap:     "checkpoint-456",
				Driver: &libvirtxml.DomainDiskDriver{
					Type: "qcow2",
				},
				Source: &libvirtxml.DomainDiskSource{
					File: &libvirtxml.DomainDiskSourceFile{
						File: "/path/to/disk.qcow2",
					},
				},
			},
			{
				Name:       "vdb",
				Checkpoint: "no",
			},
		},
	},
}
```

Would correspond to the following libvirt XML:

```xml
<domaincheckpoint>
  <name>checkpoint-456</name>
  <description>Checkpoint after updates</description>
  <disks>
    <disk name='vda' checkpoint='bitmap' bitmap='checkpoint-456'>
      <driver type='qcow2'/>
      <source file='/path/to/disk.qcow2'/>
    </disk>
    <disk name='vdb' checkpoint='no'/>
  </disks>
</domaincheckpoint>
