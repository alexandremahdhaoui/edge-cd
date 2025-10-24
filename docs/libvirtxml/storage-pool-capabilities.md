# libvirtxml.StoragePoolCapabilities

The `StoragePoolCapabilities` structure, while not a single top-level struct in `libvirt.org/go/libvirtxml`, is represented by the collective capabilities described within the `StoragePool` and related structs. These capabilities define the types of storage pools supported, their features, and the options available for sources and targets. This document details how these Go structs map to the information described in the [Libvirt Storage Pool Capabilities XML format](https://libvirt.org/formatstoragecaps.html).

## Overview of Storage Pool Capabilities

Libvirt provides information about storage pool capabilities through the `virConnectGetStoragePoolCapabilities` function, which returns an XML document describing various storage pool types and their supported features. The Go structs that represent these capabilities are primarily derived from the `StoragePool`, `StoragePoolSource`, and `StoragePoolTarget` definitions.

## Key Structs and Their Mapping

### `StoragePool`

The `StoragePool` struct itself represents a storage pool definition, and its fields provide insights into the capabilities of different pool types.

```go
// StoragePool represents a storage pool definition.
type StoragePool struct {
	XMLName    xml.Name  // XMLName is the XML element name, typically "pool".
	Type       string   ).
	Name       string    // Name of the storage pool.
	UUID       string    // UUID of the storage pool.
	Allocation *StoragePoolSize  // Maps to <allocation> element, representing allocated space.
	Capacity   *StoragePoolSize  // Maps to <capacity> element, representing total capacity.
	Available  *StoragePoolSize  // Maps to <available> element, representing available space.
	Features   *StoragePoolFeatures  // Maps to <features> element, describing pool features.
	Target     *StoragePoolTarget    // Maps to <target> element, describing the pool's target configuration.
	Source     *StoragePoolSource    // Maps to <source> element, describing the pool's source configuration.
	Refresh    *StoragePoolRefresh  // Refresh configuration for the storage pool.
	// ... other fields
}
```

*   **`Type`**: This field directly maps to the `type` attribute of the `<pool>` element in the capabilities XML, indicating the type of storage pool (e.g., `dir`, `fs`, `rbd`). The `supported` attribute is implicitly handled by the presence of a pool type.

### `StoragePoolSource`

The `StoragePoolSource` struct describes the source of the storage pool.

```go
// StoragePoolSource describes the source of the storage pool.
type StoragePoolSource struct {
	Name      string
	Dir       *StoragePoolSourceDir
	Host      []StoragePoolSourceHost
	Device    []StoragePoolSourceDevice
	Auth      *StoragePoolSourceAuth
	Vendor    *StoragePoolSourceVendor
	Product   *StoragePoolSourceProduct
	Format    *StoragePoolSourceFormat // Maps to <source><format type='...'/>, relevant for sourceFormatType enum.
	Protocol  *StoragePoolSourceProtocol
	Adapter   *StoragePoolSourceAdapter
	Initiator *StoragePoolSourceInitiator
}
```

*   **`Format`**: When present within `StoragePoolSource`, this maps to the `<source><format type='...'/>` element. The `type` attribute here corresponds to the `sourceFormatType` enum values described in the libvirt documentation.

### `StoragePoolTarget`

The `StoragePoolTarget` struct describes the target configuration for the storage pool.

```go
// StoragePoolTarget describes the target configuration for the storage pool.
type StoragePoolTarget struct {
	Path        string
	Permissions *StoragePoolTargetPermissions
	Timestamps  *StoragePoolTargetTimestamps
	Encryption  *StorageEncryption // Related to storage encryption capabilities.
}
```

*   **`Format`**: While not directly a field in `StoragePoolTarget`, the `StorageVolumeTarget` struct (which is related to pool capabilities for volumes) contains a `Format` field. This field maps to the `<target><format type='...'/>` element, corresponding to the `targetFormatType` enum values.

### `StoragePoolFeatures`

The `StoragePoolFeatures` struct describes specific features supported by the storage pool.

```go
// StoragePoolFeatures describes specific features supported by the storage pool.
type StoragePoolFeatures struct {
	COW *StoragePoolFeatureCOW  // Maps to the <features><cow> element.
}
```

*   **`COW`**: This field, if present, maps to the `<features><cow>` element, indicating support for Copy-on-Write functionality.

This documentation provides a mapping between the Go structs in `libvirt-go-xml` and the libvirt XML format for storage pool capabilities, enabling users to understand and utilize these features effectively.
