# libvirtxml.StoragePool and libvirtxml.StorageVolume

This document details the `StoragePool` and `StorageVolume` structs in `libvirt.org/go/libvirtxml`, which represent the XML configuration for libvirt storage pools and volumes. These structs map to the elements and attributes described in the [Libvirt Storage Pool XML format](https://libvirt.org/formatstorage.html) and [Libvirt Storage Volume XML format](https://libvirt.org/formatstorage.html), as well as incorporating details from the [Libvirt Storage Encryption XML format](https://libvirt.org/formatstorageencryption.html).

## `StoragePool` Struct Definition

```go
// StoragePool represents the XML configuration for a libvirt storage pool.
type StoragePool struct {
	XMLName    xml.Name
	Type       string // The type of storage backend (e.g., "dir", "fs", "logical", "iscsi").
	Name       string // A unique name for the pool.
	UUID       string // A globally unique identifier for the pool.
	Allocation *StoragePoolSize // Total storage allocation for the pool in bytes.
	Capacity   *StoragePoolSize // Total storage capacity for the pool in bytes.
	Available  *StoragePoolSize // Free space available for allocating new volumes in the pool in bytes.
	Features   *StoragePoolFeatures // Optional features supported by the pool, like copy-on-write.
	Target     *StoragePoolTarget // Describes the mapping of the storage pool into the host filesystem.
	Source     *StoragePoolSource // Describes the source of the storage pool.
	Refresh    *StoragePoolRefresh // Overrides for pool refresh behavior.
	// Pool backend namespaces must be last
	FSCommandline  *StoragePoolFSCommandline // Namespace for FS pool specific mount options.
	RBDCommandline *StoragePoolRBDCommandline // Namespace for RBD pool specific configuration options.
}
```

## `StorageVolume` Struct Definition

```go
// StorageVolume represents the XML configuration for a libvirt storage volume.
type StorageVolume struct {
	XMLName      xml.Name
	Type         string // The actual type of the volume (e.g., "file", "block", "dir", "network").
	Name         string // The name for the volume, unique within the pool.
	Key          string // An identifier for the volume.
	Allocation   *StorageVolumeSize // The total storage allocation for the volume in bytes.
	Capacity     *StorageVolumeSize // The logical capacity for the volume in bytes.
	Physical     *StorageVolumeSize // The host physical size of the target storage volume in bytes.
	Target       *StorageVolumeTarget // Describes the mapping of the storage volume into the host filesystem.
	BackingStore *StorageVolumeBackingStore // Describes the optional copy-on-write backing store for the storage volume.
}
```

## Nested Structs

```go
// StoragePoolSize represents a storage size with a unit.
type StoragePoolSize struct {
	Unit  string // The unit of the size (e.g., "B", "KiB", "MiB", "GiB", "TiB").
	Value uint64 // The numeric value of the size.
}

// StoragePoolFeatures represents optional features supported by a storage pool.
type StoragePoolFeatures struct {
	COW *StoragePoolFeatureCOW // Controls copy-on-write behavior for images in the pool.
}

// StoragePoolFeatureCOW represents the copy-on-write feature.
type StoragePoolFeatureCOW struct {
	State string // State of the COW feature ('yes' or 'no').
}

// StoragePoolTarget describes the mapping of the storage pool into the host filesystem.
type StoragePoolTarget struct {
	Path        string // The location at which the pool will be mapped into the local filesystem namespace.
	Permissions *StoragePoolTargetPermissions // Permissions for the pool's directory.
	Timestamps  *StoragePoolTargetTimestamps // Timing information about the pool.
	Encryption  *StorageEncryption // Encryption details for the pool.
}

// StoragePoolTargetPermissions describes the permissions for a storage pool's directory.
type StoragePoolTargetPermissions struct {
	Owner string // The numeric user ID of the owner.
	Group string // The numeric group ID of the owner.
	Mode  string // The octal permission set.
	Label string // The MAC (e.g., SELinux) label string.
}

// StoragePoolTargetTimestamps provides timing information about the pool.
type StoragePoolTargetTimestamps struct {
	Atime string // Access time.
	Mtime string // Modification time.
	Ctime string // Change time.
}

// StoragePoolSource describes the source of the storage pool.
type StoragePoolSource struct {
	Name      string // Name of the source element (e.g., for logical, rbd, gluster).
	Dir       *StoragePoolSourceDir // Source for pools backed by directories.
	Host      []*StoragePoolSourceHost // Source for pools from remote servers.
	Device    []*StoragePoolSourceDevice // Source for pools backed by physical devices.
	Auth      *StoragePoolSourceAuth // Authentication credentials for accessing the source.
	Vendor    *StoragePoolSourceVendor // Optional vendor information of the storage device.
	Product   *StoragePoolSourceProduct // Optional product name of the storage device.
	Format    *StoragePoolSourceFormat // Format of the pool (e.g., filesystem type).
	Protocol  *StoragePoolSourceProtocol // Protocol version for netfs pools.
	Adapter   *StoragePoolSourceAdapter // Source for pools backed by SCSI adapters.
	Initiator *StoragePoolSourceInitiator // Initiator information for iSCSI-direct pools.
}

// StoragePoolSourceDir provides the source path for directory-based pools.
type StoragePoolSourceDir struct {
	Path string // The fully qualified path to the backing directory.
}

// StoragePoolSourceHost specifies a remote server for storage pools.
type StoragePoolSourceHost struct {
	Name string // Hostname or IP address of the server.
	Port string // Optional port number for the protocol.
}

// StoragePoolSourceDevice provides the source path for device-based pools.
type StoragePoolSourceDevice struct {
	Path          string // The fully qualified path to the block device node.
	PartSeparator string // Optional attribute for device mapper multipath devices.
	FreeExtents   []*StoragePoolSourceDeviceFreeExtent // Information about available extents on the device.
}

// StoragePoolSourceDeviceFreeExtent describes an available extent on a device.
type StoragePoolSourceDeviceFreeExtent struct {
	Start uint64 // The starting boundary of the extent in bytes.
	End   uint64 // The ending boundary of the extent in bytes.
}

// StoragePoolSourceAuth provides authentication credentials for accessing the source.
type StoragePoolSourceAuth struct {
	Type     string // Authentication type ('chap' or 'ceph').
	Username string // Username for authentication.
	Secret   *StoragePoolSourceAuthSecret // Reference to the secret object holding credentials.
}

// StoragePoolSourceAuthSecret references a libvirt secret object.
type StoragePoolSourceAuthSecret struct {
	Usage string // Usage category of the secret (e.g., 'libvirtiscsi', 'ceph_cluster').
	UUID  string // UUID of the secret object.
}

// StoragePoolSourceAdapter provides information about SCSI adapters.
type StoragePoolSourceAdapter struct {
	Type       string // Type of adapter ('scsi_host' or 'fc_host').
	Name       string // Name of the SCSI adapter.
	Parent     string // Parent SCSI host for FC adapters.
	Managed    string // Indicates if libvirt manages the adapter ('yes' or 'no').
	WWNN       string // World Wide Node Name for FC adapters.
	WWPN       string // World Wide Port Name for FC adapters.
	ParentAddr *StoragePoolSourceAdapterParentAddr // PCI address of the parent SCSI host.
}

// StoragePoolSourceAdapterParentAddr specifies the PCI address of a parent SCSI host.
type StoragePoolSourceAdapterParentAddr struct {
	UniqueID string // Unique ID of the parent SCSI host.
	Address  *StoragePoolPCIAddress // PCI address details.
}

// StoragePoolPCIAddress specifies PCI address details.
type StoragePoolPCIAddress struct {
	Domain   *uint // PCI domain address.
	Bus      *uint // PCI bus address.
	Slot     *uint // PCI slot address.
	Function *uint // PCI function address.
}

// StoragePoolSourceInitiator provides iSCSI initiator information.
type StoragePoolSourceInitiator struct {
	IQN *StoragePoolSourceInitiatorIQN // iSCSI Qualified Name (IQN).
}

// StoragePoolSourceInitiatorIQN specifies the iSCSI Qualified Name.
type StoragePoolSourceInitiatorIQN struct {
	Name string // The IQN name.
}

// StoragePoolSourceProtocol specifies the NFS protocol version.
type StoragePoolSourceProtocol struct {
	Version string // NFS protocol version (e.g., "3").
}

// StoragePoolSourceVendor provides optional vendor information.
type StoragePoolSourceVendor struct {
	Name string // Vendor name.
}

// StoragePoolSourceProduct provides optional product information.
type StoragePoolSourceProduct struct {
	Name string // Product name.
}

// StoragePoolSourceFormat specifies the format of the pool or volume.
type StoragePoolSourceFormat struct {
	Type string // Format type (e.g., "raw", "qcow2", "nfs", "gpt").
}

// StoragePoolRefresh overrides for pool refresh behavior.
type StoragePoolRefresh struct {
	Volume *StoragePoolRefreshVol // Overrides for volume refresh behavior.
}

// StoragePoolRefreshVol specifies volume refresh behavior overrides.
type StoragePoolRefreshVol struct {
	Allocation string // Method for computing volume allocation ('default' or 'capacity').
}

// StoragePoolFSCommandline provides namespace for FS pool specific mount options.
type StoragePoolFSCommandline struct {
	Options []*StoragePoolFSCommandlineOption // List of mount options.
}

// StoragePoolFSCommandlineOption represents a single mount option.
type StoragePoolFSCommandlineOption struct {
	Name string // The name of the mount option.
}

// StoragePoolRBDCommandline provides namespace for RBD pool specific configuration options.
type StoragePoolRBDCommandline struct {
	Options []*StoragePoolRBDCommandlineOption // List of RBD configuration options.
}

// StoragePoolRBDCommandlineOption represents a single RBD configuration option.
type StoragePoolRBDCommandlineOption struct {
	Name  string // The name of the configuration option.
	Value string // The value of the configuration option.
}

// StorageVolumeBackingStore describes the optional copy-on-write backing store for a storage volume.
type StorageVolumeBackingStore struct {
	Path        string // The absolute path to the backing store file.
	Format      *StorageVolumeTargetFormat // The format of the backing store.
	Permissions *StorageVolumeTargetPermissions // Permissions for the backing file.
}

// StorageVolumeTarget describes the mapping of the storage volume into the host filesystem.
type StorageVolumeTarget struct {
	Path        string // The absolute path where the volume can be accessed on the local filesystem.
	Format      *StorageVolumeTargetFormat // The format of the volume (e.g., "raw", "qcow2").
	Permissions *StorageVolumeTargetPermissions // Permissions for the volume file.
	Timestamps  *StorageVolumeTargetTimestamps // Timing information about the volume.
	Compat      string // Compatibility level for the volume format (e.g., "1.1" for qcow2).
	ClusterSize *StorageVolumeTargetClusterSize // qcow2 cluster size.
	NoCOW       *struct{} // Disables copy-on-write for the volume.
	Features    []*StorageVolumeTargetFeature // Format-specific features (e.g., lazy_refcounts for qcow2).
	Encryption  *StorageEncryption // Encryption details for the volume.
}

// StorageVolumeTargetFormat specifies the format of the storage volume.
type StorageVolumeTargetFormat struct {
	Type string // The format type (e.g., "raw", "qcow2").
}

// StorageVolumeTargetPermissions describes the permissions for a storage volume file.
type StorageVolumeTargetPermissions struct {
	Owner string // The numeric user ID of the owner.
	Group string // The numeric group ID of the owner.
	Mode  string // The octal permission set.
	Label string // The MAC (e.g., SELinux) label string.
}

// StorageVolumeTargetTimestamps provides timing information about the volume.
type StorageVolumeTargetTimestamps struct {
	Atime string // Access time.
	Mtime string // Modification time.
	Ctime string // Change time.
}

// StorageVolumeTargetClusterSize specifies the cluster size for qcow2 volumes.
type StorageVolumeTargetClusterSize struct {
	Unit  string // The unit of the cluster size (e.g., "KiB").
	Value uint64 // The numeric value of the cluster size.
}

// StorageVolumeTargetFeature represents format-specific features.
type StorageVolumeTargetFeature struct {
	LazyRefcounts *struct{} // Enables lazy reference counting updates.
	ExtendedL2    *struct{} // Enables subcluster allocation for qcow2 images.
}

// StorageEncryption specifies how a storage volume is encrypted.
type StorageEncryption struct {
	Format string // The encryption format (e.g., "luks", "luks2", "qcow").
	Secret *StorageEncryptionSecret // Reference to the secret object holding the encryption passphrase.
	Cipher *StorageEncryptionCipher // Specifies the cipher algorithm details.
	Ivgen  *StorageEncryptionIvgen // Specifies the initialization vector generation algorithm.
}

// StorageEncryptionSecret references a libvirt secret object for encryption.
type StorageEncryptionSecret struct {
	Type string // The type of secret (e.g., "passphrase", "volume").
	UUID string // The UUID of the secret object.
}

// StorageEncryptionCipher specifies the cipher algorithm details.
type StorageEncryptionCipher struct {
	Name string // The name of the cipher algorithm (e.g., "aes").
	Size uint64 // The size of the cipher in bits (e.g., "256").
	Mode string // The cipher algorithm mode (e.g., "cbc", "xts").
	Hash string // The master key hash algorithm (e.g., "sha256").
}

// StorageEncryptionIvgen specifies the initialization vector generation algorithm.
type StorageEncryptionIvgen struct {
	Name string // The name of the IV generation algorithm (e.g., "plain64").
	Hash string // The hash algorithm used for IV generation.
}
