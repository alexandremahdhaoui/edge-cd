# libvirtxml.Filesystem (Virtiofs)

This document describes how to configure Virtiofs for file sharing between a host and a guest using libvirt XML, focusing on the relevant Go structs in `libvirt.org/go/libvirtxml`. Virtiofs provides efficient shared file system access with local file system semantics. This information is derived from the [Libvirt Virtiofs documentation](https://libvirt.org/kbase/virtiofs.html).

## Virtiofs Configuration in Domain XML

To share a host directory with a guest using Virtiofs, you need to configure the `<filesystem>` element within the `<devices>` section of the domain XML.

### `DomainFilesystem` Struct

The `DomainFilesystem` struct represents the `<filesystem>` element.

```go
// DomainFilesystem represents a shared filesystem configuration.
type DomainFilesystem struct {
	AccessMode string                   // e.g., "passthrough"
	Model      string                        // e.g., "virtiofs"
	Driver     *DomainFilesystemDriver  // Driver specific options.
	Source     *DomainFilesystemSource  // Source directory on the host.
	Target     *DomainFilesystemTarget  // Mount tag within the guest.
	// ... other fields like ReadOnly, SpaceHardLimit, etc.
}
```

### `DomainFilesystemDriver`

The `DomainFilesystemDriver` struct specifies the driver type for the filesystem.

```go
// DomainFilesystemDriver specifies the driver type for the filesystem.
type DomainFilesystemDriver struct {
	Type string  // e.g., "virtiofs"
	Queue uint  // e.g., "1024"
}
```

### `DomainFilesystemSource`

The `DomainFilesystemSource` struct defines the source directory on the host.

```go
// DomainFilesystemSource defines the source directory on the host.
type DomainFilesystemSource struct {
	Dir string  // The host directory path to be shared.
}
```

### `DomainFilesystemTarget`

The `DomainFilesystemTarget` struct defines the mount tag used within the guest to identify the shared filesystem.

```go
// DomainFilesystemTarget defines the mount tag used within the guest.
type DomainFilesystemTarget struct {
	Dir string  // The mount tag (arbitrary string).
}
```

### `DomainMemoryBacking`

Crucially, for vhost-user connections with the virtiofsd daemon, `<memoryBacking>` elements are necessary.

```go
// DomainMemoryBacking configures memory backing options.
type DomainMemoryBacking struct {
	Source *DomainMemorySource 
	Access *DomainMemoryAccess 
	// ... other memory backing options
}

// DomainMemorySource specifies the source type for memory backing.
type DomainMemorySource struct {
	Type string  // e.g., "memfd", "file", "hugepages"
}

// DomainMemoryAccess specifies memory access mode.
type DomainMemoryAccess struct {
	Mode string  // e.g., "shared"
}
```

## Example Configuration

Here's how you would configure Virtiofs for a guest, sharing the host directory `/path/to/shared/folder` with the guest mount tag `myvirtiofs`:

```go
filesystem := libvirtxml.DomainFilesystem{
	AccessMode: "passthrough",
	Driver: &libvirtxml.DomainFilesystemDriver{
		Type:  "virtiofs",
		Queue: uint(1024),
	},
	Source: &libvirtxml.DomainFilesystemSource{
		Dir: "/path/to/shared/folder",
	},
	Target: &libvirtxml.DomainFilesystemTarget{
		Dir: "myvirtiofs",
	},
	// Memory backing is essential for vhost-user connections
	MemoryBacking: &libvirtxml.DomainMemoryBacking{
		Source: &libvirtxml.DomainMemorySource{
			Type: "memfd", // or "file" if configured in qemu.conf
		},
		Access: &libvirtxml.DomainMemoryAccess{
			Mode: "shared",
		},
	},
}
```

This Go struct would translate to the following XML within the `<devices>` section of a domain definition:

```xml
<devices>
  ...
  <filesystem type='mount' accessmode='passthrough'>
    <driver type='virtiofs' queue='1024'/>
    <source dir='/path/to/shared/folder'/>
    <target dir='myvirtiofs'/>
    <memoryBacking>
      <source type='memfd'/>
      <access mode='shared'/>
    </memoryBacking>
  </filesystem>
  ...
</devices>
```

**Mounting in the Guest:**

After booting the guest with this configuration, you would mount the shared filesystem inside the guest using:

```bash
mount -t virtiofs myvirtiofs /mnt/mount/path
```

**Note:** This requires Virtiofs support in the guest kernel (Linux v5.4 or later).

**Unprivileged Mode:**

In unprivileged mode (`qemu:///session`), user/group ID mapping is available (since libvirt 10.0.0). The root user (ID 0) in the guest maps to the current user on the host. Other IDs map to subordinate user IDs specified in `/etc/subuid` and `/etc/subgid`. The `idmap` element can be used for manual tweaking of user ID mapping.

**Optional Parameters:**

Additional parameters like `<binary>` for specifying the `virtiofsd` path and attributes like `xattr` and `<cache>`, `<lock>` can be configured within the `<driver>` element or as separate elements depending on the specific libvirt version and `virtiofsd` capabilities.

**Externally-launched virtiofsd:**

Libvirtd can also connect to a `virtiofsd` daemon launched outside of libvirtd. In this scenario, socket permissions, the mount tag, and all `virtiofsd` options are managed by the external application. The XML configuration would typically involve a `<source socket='/path/to/socket'/>` element.
