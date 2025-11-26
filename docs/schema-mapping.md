# Schema Mapping Reference

How libvirt XML maps to Terraform HCL in this provider. All resources follow these rules.

## Core Rules
- XML elements become nested attributes (`schema.SingleNestedAttribute`/`schema.ListNestedAttribute`). New features should not use Terraform blocks.
- Repeated elements become lists of nested objects. Order matters.
- XML attributes stay scalar inside the containing object.
- Names are converted to snake_case; common acronyms stay intact (`mac_address`, `uuid`, `nvram`).
- Optional fields are only populated on reads when the user set them. Computed fields are always read. Required fields are always populated.

## Common Patterns
- Containers become nested objects that hold their children, e.g. `<devices>` -> `devices = { disks = [...], interfaces = [...] }`.
- Value + unit flattening: `<memory unit="MiB">512</memory>` -> `memory = 512`, `memory_unit = "MiB"`. If there is one extra attribute, flatten it too (e.g. `max_memory` + `max_memory_slots`). With two or more extra attributes, use a nested object (e.g. `vcpu = { value = 4, placement = "static", cpuset = "0-3" }`).
- Presence-only elements (like `<acpi/>`) map to booleans. `true` emits the element; `false`/`null` omits it.
- Union/variant branches stay nested: `source = { file = "...", block = "...", volume = { pool = "...", volume = "..." } }`. Only set one branch.
- XML `yes/no` attributes become booleans (e.g. loader readonly).
- Where the XML type is implied by which branch is set (e.g. interfaces), omit a `type` attribute; the chosen branch defines the type.

## Examples
```xml
<domain type="kvm">
  <memory unit="MiB">512</memory>
  <vcpu placement="static" cpuset="0-3" current="2">4</vcpu>
  <devices>
    <disk type="file">
      <source file="/var/lib/libvirt/images/disk.qcow2"/>
      <target dev="vda" bus="virtio"/>
    </disk>
    <interface type="network">
      <source network="default"/>
      <model type="virtio"/>
    </interface>
  </devices>
</domain>
```

```hcl
resource "libvirt_domain" "example" {
  memory       = 512
  memory_unit  = "MiB"
  vcpu         = { value = 4, placement = "static", cpuset = "0-3", current = 2 }

  devices = {
    disks = [{
      source = { file = "/var/lib/libvirt/images/disk.qcow2" }
      target = { dev = "vda", bus = "virtio" }
    }]
    interfaces = [{
      source = { network = { network = "default" } }
      model  = { type = "virtio" }
    }]
  }
}
```

## Validation Source
- The authoritative schemas are the libvirt RNG files in `/usr/share/libvirt/schemas/`.
- We only expose fields supported by `libvirt.org/go/libvirtxml`; if libvirtxml lacks a field, the provider omits it.
