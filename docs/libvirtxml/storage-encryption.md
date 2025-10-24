# libvirtxml.StorageEncryption

The `StorageEncryption` struct in `libvirt.org/go/libvirtxml` represents the XML configuration for libvirt storage volume encryption. This allows for the encryption of storage volumes, providing data security. This document details the `StorageEncryption` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Storage volume encryption XML format](https://libvirt.org/formatstorageencryption.html).

## `StorageEncryption` Struct Definition

```go
// StorageEncryption represents the XML configuration for libvirt storage volume encryption.
type StorageEncryption struct {
	Format string  // Encryption format (e.g., "luks", "qcow").
	Secret *StorageEncryptionSecret  // Secret used for encryption/decryption.
	Cipher *StorageEncryptionCipher  // Cipher algorithm used for encryption/decryption.
	Ivgen  *StorageEncryptionIvgen  // Initialization vector generation algorithm.
}
```

## Nested Structs

### `StorageEncryptionSecret`

The `StorageEncryptionSecret` struct represents the secret used for encryption/decryption.

```go
// StorageEncryptionSecret represents the secret used for encryption/decryption.
type StorageEncryptionSecret struct {
	Type string // Mandatory attribute, e.g., 'passphrase'.
	UUID string // Mandatory attribute, the UUID of the secret.
}
```

### `StorageEncryptionCipher`

The `StorageEncryptionCipher` struct describes the cipher algorithm used for encryption or decryption.

```go
// StorageEncryptionCipher describes the cipher algorithm used for encryption or decryption.
type StorageEncryptionCipher struct {
	Name string // The name of the cipher algorithm (e.g., 'aes').
	Size uint64 // The size of the cipher in bits (e.g., '256').
	Mode string // Optional cipher algorithm mode (e.g., 'cbc', 'xts').
	Hash string // Optional master key hash algorithm (e.g., 'sha256').
}
```

### `StorageEncryptionIvgen`

The `StorageEncryptionIvgen` struct describes the initialization vector generation algorithm.

```go
// StorageEncryptionIvgen describes the initialization vector generation algorithm.
type StorageEncryptionIvgen struct {
	Name string // The name of the algorithm (e.g., 'plain64').
	Hash string // Optional hash algorithm (e.g., 'sha256').
}
```

## Mapping to Libvirt XML

The `StorageEncryption` struct corresponds to the `<encryption>` element in libvirt XML.

*   The `Format` field maps to the `format` attribute of the `<encryption>` element (e.g., `luks`, `qcow`).
*   The `Secret` field maps to the `<secret>` child element, with `Type` mapping to the `type` attribute and `UUID` mapping to the `uuid` attribute.
*   The `Cipher` field maps to the `<cipher>` child element, with its fields mapping to the `name`, `size`, `mode`, and `hash` attributes.
*   The `Ivgen` field maps to the `<ivgen>` child element, with its fields mapping to the `name` and `hash` attributes.

For example, the following Go struct:

```go
encryption := libvirtxml.StorageEncryption{
	Format: "luks",
	Secret: &libvirtxml.StorageEncryptionSecret{
		Type: "passphrase",
		UUID: "f52a81b2-424e-490c-823d-6bd4235bc572",
	},
	Cipher: &libvirtxml.StorageEncryptionCipher{
		Name: "twofish",
		Size: 256,
		Mode: "cbc",
		Hash: "sha256",
	},
	Ivgen: &libvirtxml.StorageEncryptionIvgen{
		Name: "plain64",
		Hash: "sha256",
	},
}
```

Would correspond to the following libvirt XML:

```xml
<encryption format='luks'>
  <secret type='passphrase' uuid='f52a81b2-424e-490c-823d-6bd4235bc572'/>
  <cipher name='twofish' size='256' mode='cbc' hash='sha256'/>
  <ivgen name='plain64' hash='sha256'/>
</encryption>
