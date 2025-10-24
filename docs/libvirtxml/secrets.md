# libvirtxml.Secret

The `Secret` struct in `libvirt.org/go/libvirtxml` represents secrets managed by libvirt, such as passphrases for encrypted volumes or TLS credentials. This document details the `Secret` struct and its nested components, mapping them to the corresponding elements and attributes in the [Libvirt Secret XML format](https://libvirt.org/formatsecret.html).

## `Secret` Struct Definition

```go
// Secret represents secrets managed by libvirt.
type Secret struct {
	XMLName     xml.Name
	Ephemeral   string // Optional attribute: 'yes' or 'no'. Defaults to 'no'.
	Private     string // Optional attribute: 'yes' or 'no'. Defaults to 'no'.
	Description string // Optional human-readable description.
	UUID        string // Optional unique identifier for the secret.
	Usage       *SecretUsage // Specifies the usage category of the secret.
}
```

## Nested Structs

### `SecretUsage`

The `SecretUsage` struct specifies the category and context for the secret.

```go
// SecretUsage specifies the usage category and context for the secret.
type SecretUsage struct {
	Type   string // Mandatory attribute, e.g., 'volume', 'ceph', 'iscsi', 'tls', 'vtpm'.
	Volume string // Used when Type is 'volume', specifies the path to the volume.
	Name   string // Used for 'ceph', 'iscsi', 'tls', 'vtpm' to specify a usage name.
	Target string // Used for 'iscsi' and 'vtpm', specifies the target name.
}
```

## Mapping to Libvirt XML

The `Secret` struct corresponds to the `<secret>` element in libvirt XML.

*   The `Ephemeral` field maps to the `ephemeral` attribute of the `<secret>` element.
*   The `Private` field maps to the `private` attribute of the `<secret>` element.
*   The `Description` field maps to the `<description>` element.
*   The `UUID` field maps to the `<uuid>` element.
*   The `Usage` field maps to the `<usage>` element, with its `Type` field mapping to the `type` attribute and `Volume` or `Name` fields mapping to corresponding child elements within `<usage>`.

For example, the following Go struct for a volume secret:

```go
volumeSecret := libvirtxml.Secret{
	Ephemeral:   "no",
	Private:     "yes",
	Description: "Super secret name of my first puppy",
	UUID:        "0a81f5b2-8403-7b23-c8d6-21ccc2f80d6f",
	Usage: &libvirtxml.SecretUsage{
		Type:   "volume",
		Volume: "/var/lib/libvirt/images/puppyname.img",
	},
}
```

Would correspond to the following libvirt XML:

```xml
<secret ephemeral='no' private='yes'>
  <description>Super secret name of my first puppy</description>
  <uuid>0a81f5b2-8403-7b23-c8d6-21ccc2f80d6f</uuid>
  <usage type='volume'>
    <volume>/var/lib/libvirt/images/puppyname.img</volume>
  </usage>
</secret>
```

And for a Ceph secret:

```go
cephSecret := libvirtxml.Secret{
	Description: "CEPH passphrase example",
	Usage: &libvirtxml.SecretUsage{
		Type: "ceph",
		Name: "ceph_example",
	},
}
```

Would correspond to:

```xml
<secret>
  <description>CEPH passphrase example</description>
  <usage type='ceph'>
    <name>ceph_example</name>
  </usage>
</secret>
