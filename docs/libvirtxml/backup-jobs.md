# libvirtxml.DomainBackup

The `DomainBackup` struct in `libvirt.org/go/libvirtxml` represents the XML configuration for libvirt domain backups, supporting both full and incremental backups, as well as push and pull modes. This document details the `DomainBackup` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Backup XML format](https://libvirt.org/formatbackup.html).

## `DomainBackup` Struct Definition

```go
// DomainBackup represents the XML configuration for libvirt domain backups.
type DomainBackup struct {
	XMLName     xml.Name  // XMLName is the XML element name, typically "domainbackup".
	Incremental string    // Optional: Name of an existing checkpoint for incremental backups.
	Push        *DomainBackupPush  // Represents the push mode configuration.
	Pull        *DomainBackupPull  // Represents the pull mode configuration.
}
```

## Nested Structs

### `DomainBackupPush`

The `DomainBackupPush` struct configures a push-mode backup.

```go
// DomainBackupPush configures a push-mode backup.
type DomainBackupPush struct {
	Disks *DomainBackupPushDisks 
}
```

### `DomainBackupPushDisks`

The `DomainBackupPushDisks` struct contains a list of disks to be included in a push-mode backup.

```go
// DomainBackupPushDisks contains a list of disks for a push-mode backup.
type DomainBackupPushDisks struct {
	Disks []DomainBackupPushDisk 
}
```

### `DomainBackupPushDisk`

The `DomainBackupPushDisk` struct describes the backup properties for a specific disk in push mode.

```go
// DomainBackupPushDisk describes backup properties for a specific disk in push mode.
type DomainBackupPushDisk struct {
	Name        string               
	Backup      string                // "yes" or "no"
	BackupMode  string                // "full" or "incremental"
	Incremental string                // Checkpoint name for incremental backup
	Driver      *DomainBackupDiskDriver 
	Target      *DomainDiskSource    
}
```

### `DomainBackupDiskDriver`

The `DomainBackupDiskDriver` struct specifies the driver type for the backup destination.

```go
// DomainBackupDiskDriver specifies the driver type for the backup destination.
type DomainBackupDiskDriver struct {
	Type string  // e.g., "raw", "qcow2"
}
```

### `DomainBackupPull`

The `DomainBackupPull` struct configures a pull-mode backup.

```go
// DomainBackupPull configures a pull-mode backup.
type DomainBackupPull struct {
	Server *DomainBackupPullServer 
	Disks  *DomainBackupPullDisks  
}
```

### `DomainBackupPullServer`

The `DomainBackupPullServer` struct configures the NBD server for pull-mode backups.

```go
// DomainBackupPullServer configures the NBD server for pull-mode backups.
type DomainBackupPullServer struct {
	TLS  string                   
	TCP  *DomainBackupPullServerTCP 
	UNIX *DomainBackupPullServerUNIX 
	FD   *DomainBackupPullServerFD   
}
```

### `DomainBackupPullServerTCP`

Configures TCP transport for the NBD server.

```go
// DomainBackupPullServerTCP configures TCP transport for the NBD server.
type DomainBackupPullServerTCP struct {
	Name string 
	Port uint   
}
```

### `DomainBackupPullServerUNIX`

Configures UNIX socket transport for the NBD server.

```go
// DomainBackupPullServerUNIX configures UNIX socket transport for the NBD server.
type DomainBackupPullServerUNIX struct {
	Socket string 
}
```

### `DomainBackupPullServerFD`

Configures file descriptor transport for the NBD server.

```go
// DomainBackupPullServerFD configures file descriptor transport for the NBD server.
type DomainBackupPullServerFD struct {
	FDGroup string 
}
```

### `DomainBackupPullDisks`

The `DomainBackupPullDisks` struct contains a list of disks for a pull-mode backup.

```go
// DomainBackupPullDisks contains a list of disks for a pull-mode backup.
type DomainBackupPullDisks struct {
	Disks []DomainBackupPullDisk 
}
```

### `DomainBackupPullDisk`

The `DomainBackupPullDisk` struct describes the backup properties for a specific disk in pull mode.

```go
// DomainBackupPullDisk describes backup properties for a specific disk in pull mode.
type DomainBackupPullDisk struct {
	Name         string               
	Backup       string                // "yes" or "no"
	BackupMode   string                // "full" or "incremental"
	Incremental  string                // Checkpoint name for incremental backup
	ExportName   string                // NBD export name
	ExportBitmap string                // Bitmap name for incremental NBD export
	Driver       *DomainBackupDiskDriver 
	Scratch      *DomainDiskSource     // Local scratch storage configuration
}
```

## Mapping to Libvirt XML

The `DomainBackup` struct corresponds to the `<domainbackup>` element in libvirt XML.

*   **`Incremental`**: Maps to the `<incremental>` element.
*   **`Push`**: Maps to the `<disks>` element within a push-mode backup configuration.
    *   `DomainBackupPushDisks` maps to `<disks>`, containing `DomainBackupPushDisk` structs which map to `<disk>` elements.
    *   `DomainBackupPushDisk` maps to `<disk>`, with `Name` to `name`, `Backup` to `backup`, `BackupMode` to `backupmode`, `Incremental` to `incremental`, `Driver` to `<driver>`, and `Target` to `<target>`.
*   **`Pull`**: Maps to the `<server>` element and `<disks>` element within a pull-mode backup configuration.
    *   `DomainBackupPullServer` maps to `<server>`, with its TCP, UNIX, or FD fields mapping to corresponding elements like `<server><tcp>`, `<server><unix>`, or `<server><fd>`.
    *   `DomainBackupPullDisks` maps to `<disks>`, containing `DomainBackupPullDisk` structs which map to `<disk>` elements.
    *   `DomainBackupPullDisk` maps to `<disk>`, with `Name` to `name`, `Backup` to `backup`, `BackupMode` to `backupmode`, `Incremental` to `incremental`, `ExportName` to `exportname`, `ExportBitmap` to `exportbitmap`, `Driver` to `<driver>`, and `Scratch` to `<scratch>`.

For example, the following Go struct for a full push backup:

```go
pushBackup := libvirtxml.DomainBackup{
	Push: &libvirtxml.DomainBackupPush{
		Disks: &libvirtxml.DomainBackupPushDisks{
			Disks: []libvirtxml.DomainBackupPushDisk{
				{
					Name: "vda",
					Backup: "yes",
					Target: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: "/path/to/vda.backup",
						},
					},
					Driver: &libvirtxml.DomainBackupDiskDriver{
						Type: "raw",
					},
				},
				{
					Name:   "vdb",
					Backup: "no",
				},
			},
		},
	},
}
```

Would correspond to the following libvirt XML:

```xml
<domainbackup>
  <disks>
    <disk name='vda' backup='yes'>
      <target file='/path/to/vda.backup'/>
      <driver type='raw'/>
    </disk>
    <disk name='vdb' backup='no'/>
  </disks>
</domainbackup>
```

And for an incremental pull backup with a specific checkpoint:

```go
pullBackup := libvirtxml.DomainBackup{
	Incremental: "1525889631",
	Pull: &libvirtxml.DomainBackupPull{
		Server: &libvirtxml.DomainBackupPullServer{
			TCP: &libvirtxml.DomainBackupPullServerTCP{
				Name: "localhost",
				Port: 12345,
			},
		},
		Disks: &libvirtxml.DomainBackupPullDisks{
			Disks: []libvirtxml.DomainBackupPullDisk{
				{
					Name:        "vda",
					Backup:      "yes",
					BackupMode:  "incremental",
					Incremental: "1525889631",
					ExportName:  "vda_export",
					Scratch: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: "/path/to/vda.scratch",
						},
					},
				},
			},
		},
	},
}
```

Would correspond to:

```xml
<domainbackup incremental='1525889631' mode='pull'>
  <server>
    <tcp name='localhost' port='12345'/>
  </server>
  <disks>
    <disk name='vda' backup='yes' backupmode='incremental' incremental='1525889631' exportname='vda_export'>
      <driver type='qcow2'/>
      <scratch file='/path/to/vda.scratch'/>
    </disk>
  </disks>
</domainbackup>
