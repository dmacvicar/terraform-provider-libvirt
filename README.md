# Terraform provider for libvirt

![Build and Test](https://github.com/dmacvicar/terraform-provider-libvirt/actions/workflows/test.yml/badge.svg)

This provider allows managing libvirt resources (virtual machines, storage pools, networks) using Terraform. It communicates with libvirt using its API to define, configure, and manage virtualization resources.

This is a complete rewrite of the legacy provider. The legacy provider (v0.8.x and earlier) is maintained in the [v0.8 branch](https://github.com/dmacvicar/terraform-provider-libvirt/tree/v0.8). Starting from v0.9.0, all releases will be based on this new rewrite.

## Goals

This rewrite improves upon the legacy provider in several ways:

1. **API Fidelity** - Models the [libvirt XML schemas](https://libvirt.org/format.html) directly instead of abstracting them, giving users full access to libvirt features. Schema coverage is bounded by what [libvirtxml](https://pkg.go.dev/libvirt.org/go/libvirtxml) supports.
2. **Current Framework** - Built with Terraform Plugin Framework, as the SDK v2 used in the legacy provider is deprecated
3. **Best Practices** - Follows [HashiCorp's provider design principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)

## Supported Resources & XML Coverage

| Resource | Status | XML Coverage |
|----------|--------|--------------|
| `libvirt_domain` | ✅ Supported | Full coverage of libvirtxml’s domain schema (devices, CPU, memory, features, RNG, TPM, etc.). |
| `libvirt_network` | ✅ Supported | Full coverage of libvirtxml network schema (forwarding modes, bridge, DHCP, VLAN, virtual ports, etc.). |
| `libvirt_pool` | ✅ Supported | Full coverage of libvirtxml storage pool schema (dir/logical/iscsi/etc.). |
| `libvirt_volume` | ✅ Supported | Full coverage of libvirtxml storage volume schema (target, backing_store, encryption, timestamps). |

Everything exposed in these resources maps directly to the corresponding libvirt XML; if libvirtxml adds new fields we regenerate and pick them up automatically. Additional libvirt resources (secrets, nodes, interfaces, etc.) will be added later following the same pattern—see TODO.md for the roadmap.

## Design Principles

- **Schema Coverage**: We support all fields that `libvirt.org/go/libvirtxml` implements from the official libvirt schemas (located at `/usr/share/libvirt/schemas/`). If libvirtxml doesn't support a feature yet, neither do we - we don't create custom XML structs.
- **No Abstraction**: The Terraform schema mirrors the libvirt XML structure as closely as possible, providing full access to underlying features rather than simplified abstractions.
- **User Input Preservation**: For optional+computed fields, we preserve the user's input value even when libvirt normalizes it (e.g., "q35" vs "pc-q35-10.1") to avoid unnecessary diffs.

## XML to HCL Mapping

This provider maps libvirt's XML structure to Terraform's HCL configuration language using a consistent, predictable pattern: XML elements become nested attributes, attributes/text become scalar attributes, repeated elements become lists, and elements with chardata plus attributes flatten into sibling fields. The examples below show these rules in practice.

### Example Mapping

**Libvirt XML:**
```xml
<domain type="kvm">
  <name>example-vm</name>
  <memory unit="MiB">512</memory>
  <vcpu>1</vcpu>
  <clock offset="utc">
    <timer name="rtc" tickpolicy="catchup">
      <catchup threshold="123" slew="120" limit="10000"/>
    </timer>
    <timer name="pit" tickpolicy="delay"/>
  </clock>
</domain>
```

**Terraform HCL:**
```hcl
resource "libvirt_domain" "example" {
  name        = "example-vm"
  type        = "kvm"
  memory      = 512        # Flattened from <memory unit='MiB'>512</memory>
  memory_unit = "MiB"      # Optional, defaults based on libvirt
  vcpu        = 1          # Flattened from <vcpu>1</vcpu>

  clock {
    offset = "utc"

    timer {
      name       = "rtc"
      tickpolicy = "catchup"

      catchup {
        threshold = 123
        slew      = 120
        limit     = 10000
      }
    }

    timer {
      name       = "pit"
      tickpolicy = "delay"
    }
  }
}
```

### Quick Mapping Patterns

1. **Elements → Nested Attributes**

```xml
<clock><timer name="rtc"/></clock>
```

```hcl
clock = { timers = [{ name = "rtc" }] }
```

2. **Value-with-Unit Flattening**

```xml
<memory unit='KiB'>524288</memory>
```

```hcl
memory      = 524288
memory_unit = "KiB"
```

   Leave the `_unit` attribute unset to let libvirt pick its default.

3. **Boolean to Presence**

```xml
<acpi/>
```

```hcl
features = { acpi = true }
```

These fields are presence-only: `false` (or `null`) omits the XML element entirely.

4. **Type-Dependent Sources**

```xml
<source file="/var/lib/libvirt/images/disk.qcow2"/>
```

```hcl
source = { file = "/var/lib/libvirt/images/disk.qcow2" }
```

Only set the branch you need (`file`, `block`, `volume = { ... }`, etc.).

5. **Lists Map to Arrays**

```xml
<disk/><disk/><disk/>
```

```hcl
disks = [{ ... }, { ... }, { ... }]
```

   Order matters: Terraform diffs trigger when you reorder array elements.

> **Snake_case naming:** Every XML element/attribute name is converted to snake_case automatically (e.g., `accessmode` → `access_mode`, `portgroup` → `port_group`). Common acronyms stay intact (`MAC`, `UUID`, `VLAN`, etc.), so libvirt’s `MACAddress` becomes `mac_address`, not `m_a_c_address`.

**Example with Devices (Disks and Network Interfaces):**

**Libvirt XML:**
```xml
<domain type="kvm">
  <name>example-vm</name>
  <memory unit="MiB">512</memory>
  <vcpu>1</vcpu>
  <devices>
    <disk type="file" device="disk">
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

**Terraform HCL:**
```hcl
resource "libvirt_domain" "example" {
  name   = "example-vm"
  type   = "kvm"
  memory = 512
  vcpu   = 1

  devices = {
    disks = [
      {
        source = {
          file = "/var/lib/libvirt/images/disk.qcow2"
        }
        target = {
          dev = "vda"
          bus = "virtio"
        }
      }
    ]
    interfaces = [
      {
        model = {
          type = "virtio"
        }
        source = {
          network = {
            network = "default"
          }
        }
      }
    ]
  }
}
```

In this mapping:
- `devices.disks.source` is a nested object whose attributes (e.g., `file`, `pool`, `volume`, `block`) mirror the `<source>` element attributes in libvirt XML. Only one source variant may be provided at a time.
- `devices.disks.target` is a nested object with `dev` and optional `bus`, matching `<target dev="..." bus="..."/>`.
- Disk backing chains are configured on the storage volume (`libvirt_volume.backing_store`); libvirt ignores `<backingStore>` input on domains unless the hypervisor advertises the `backingStoreInput` capability.

### Handling Elements with Text Content and Attributes

Some libvirt XML elements have both text content and attributes. For better ergonomics, we apply these patterns:

#### Simple value with unit only

**XML**: `<memory unit="MiB">512</memory>`

The unit is fixed and the value becomes a simple attribute:
```hcl
memory = 512  # Always MiB
```

This applies to all scaledInteger fields (memory, hard_limit, soft_limit, etc.). We pick a sensible default unit per field.

#### Value with unit plus one other attribute

**XML**: `<maxMemory unit="MiB" slots="16">2048</maxMemory>`

The value is flattened with a fixed unit, the other attribute becomes a separate field:
```hcl
max_memory       = 2048
max_memory_slots = 16
```

#### Value with multiple attributes

**XML**: `<vcpu placement="static" cpuset="0-3" current="2">4</vcpu>`

A nested block is used with the value and all attributes:
```hcl
vcpu {
  value     = 4
  placement = "static"
  cpuset    = "0-3"
  current   = 2
}
```

#### Source elements with type-dependent attributes

When a source element has different attribute sets depending on a type, we use a nested block:

**XML**:
```xml
<interface type="network">
  <source network="default" portgroup="web"/>
</interface>
```

**HCL**:
```hcl
interface {
  type = "network"
  source = {
    network = {
      network   = "default"
      portgroup = "web"
    }
  }
}

#### Interface Schema Migration

The legacy provider exposed network interfaces with flattened attributes such as:

```hcl
devices = {
  interfaces = [{
    type  = "network"
    model = "virtio"
    source = {
      network = "default"
    }
  }]
}
```

This rewrite models the libvirt XML more faithfully. Each interface entry is now a nested object where `model`, `source`, and type-specific fields map directly to their XML counterparts:

```hcl
devices = {
  interfaces = [
    {
      model = {
        type = "virtio"
      }
      source = {
        network = {
          network = "default"
        }
      }
      wait_for_ip = {
        timeout = 300
        source  = "lease"
      }
    },
    {
      source = {
        direct = {
          dev  = "eth0"
          mode = "bridge"
        }
      }
    }
  ]
}
```

Instead of a `type` attribute, the provider derives the interface type from the populated `source` variant (`network`, `direct`, `bridge`, etc.). This keeps the Terraform schema aligned with libvirtxml and unlocks all interface features (macvtap, passthrough, portgroups, wait_for_ip) without ad-hoc flattening.

If the source always has the same pattern, it can be flattened to a simple attribute.

### Notes

- We don't distinguish between XML attributes and elements in HCL - both become HCL attributes
- The same XML structure always maps to the same HCL structure
- This consistent mapping enables automated migration from the legacy provider or from raw libvirt XML
- **Nested Attributes vs Blocks**: Following [HashiCorp's guidance](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/blocks), new features use nested attributes (e.g., `devices = { ... }`) instead of blocks. Some existing features (`os`, `features`, `clock`, etc.) incorrectly use blocks and need conversion (see TODO.md).

For detailed XML schemas, see the [libvirt domain format documentation](https://libvirt.org/formatdomain.html).

## Development Approach

Terraform providers are largely scaffolding and domain conversion (Terraform HCL ↔ Provider API). This project leverages AI agents to accelerate development while maintaining code quality through automated linting and testing.

### Code Generation

To reduce boilerplate and ensure consistency, this provider uses a code generation system that automatically produces:

- **Terraform models** with proper `tfsdk` tags
- **Plugin Framework schemas** with correct types and optionality
- **XML conversion functions** implementing the "preserve user intent" pattern

**Benefits:**
- Reduces manual coding by ~80% for new fields
- Ensures uniform implementation across resources
- Automatically maintains 1:1 mapping with libvirtxml
- Implements best practices consistently

**Current Status:**
- ✅ Core resources (`libvirt_domain`, `libvirt_volume`, `libvirt_network`, `libvirt_pool`) use 100% generated models, schemas, and conversions.
- 🧩 In progress: documentation/validator generation, polishing remaining manual helpers (wait_for_ip overrides, provider-specific fields).
- 📋 Next: leverage RelaxNG data for automatic validators/docs and extend generation to future resources (secrets, node devices, etc.).

**Running the generator:**
```bash
go run ./internal/codegen
```

For detailed architecture, usage, and extension guide, see [`internal/codegen/README.md`](internal/codegen/README.md).

## Building from source

```bash
git clone https://github.com/dmacvicar/terraform-provider-libvirt
cd terraform-provider-libvirt
make build
```

Or manually:

```bash
go build -o terraform-provider-libvirt
```

## Installing

To install the provider locally:

```bash
make install
```

This installs to `~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/dev/linux_amd64/`

## Using the provider

```hcl
terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_domain" "example" {
  name   = "example-vm"
  memory = 512
  memory_unit   = "MiB"
  vcpu   = 1

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
```

### Connection URIs

The provider supports multiple connection transports:

```hcl
# Local system socket
provider "libvirt" {
  uri = "qemu:///system"
}

# Remote via SSH (Go library)
provider "libvirt" {
  uri = "qemu+ssh://user@host.example.com/system"
}

# Remote via SSH (native command, respects ~/.ssh/config)
provider "libvirt" {
  uri = "qemu+sshcmd://user@host.example.com/system"
}

# Remote via TLS
provider "libvirt" {
  uri = "qemu+tls://host.example.com/system"
}
```

See [docs/transports.md](./docs/transports.md) for detailed transport configuration and examples.

See the [examples](./examples) directory for more usage examples.

## Migration from Legacy Provider (v0.8.x)

### Getting Domain IP Addresses

The legacy provider exposed IP addresses directly on the domain resource via `network_interface.*.addresses`. The new provider uses a separate data source for querying IP addresses:

**Legacy provider (v0.8.x):**
```hcl
resource "libvirt_domain" "example" {
  # ... domain config ...
}

output "ip" {
  value = libvirt_domain.example.network_interface[0].addresses[0]
}
```

**New provider (v0.9+):**
```hcl
resource "libvirt_domain" "example" {
  # ... domain config ...
}

data "libvirt_domain_interface_addresses" "example" {
  domain = libvirt_domain.example.id
  source = "lease"  # or "agent" or "any"
}

output "ip" {
  value = data.libvirt_domain_interface_addresses.example.interfaces[0].addrs[0].addr
}
```

Alternatively, use the `wait_for_ip` property on the domain's interface configuration to ensure the domain has an IP before creation completes:

```hcl
resource "libvirt_domain" "example" {
  name   = "example-vm"
  memory = 512
  vcpu   = 1

  devices = {
    interfaces = [
      {
        type = "network"
        source = {
          network = "default"
        }
        wait_for_ip = {
          timeout = 300  # seconds
          source  = "lease"
        }
      }
    ]
  }
}
```

### Volume Source URLs

If you're migrating from the legacy provider and used the `source` attribute on volumes to download cloud images, note that this feature is now available via the `create.content.url` block:

**Legacy provider (v0.8.x):**
```hcl
resource "libvirt_volume" "ubuntu" {
  name   = "ubuntu.qcow2"
  pool   = "default"
  source = "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
  # size was automatically detected from Content-Length
}
```

**New provider (v0.9+):**
```hcl
resource "libvirt_volume" "ubuntu" {
  name   = "ubuntu.qcow2"
  pool   = "default"
  format = "qcow2"  # Must specify format

  create = {
    content = {
      url = "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
    }
  }
  # capacity is automatically detected from Content-Length header
}
```

**Important notes:**
1. **Format is required**: You must explicitly specify the `format` attribute (e.g., `"qcow2"`, `"raw"`). The legacy provider auto-detected format from file extension, but the new provider requires it.
2. **Capacity is computed**: Like the legacy provider, `capacity` is automatically computed from the HTTP `Content-Length` header (or file size for local files). You don't need to specify it.
3. **Local files supported**: You can use absolute paths or `file://` URIs for local files: `url = "/path/to/local.qcow2"` or `url = "file:///path/to/local.qcow2"`
4. **Content-Length required**: For HTTPS URLs, the server must provide a `Content-Length` header. If it doesn't, volume creation will fail.

## Development

### Prerequisites

- Go 1.21+
- libvirt daemon running (for acceptance tests)

### Running tests

```bash
# Run linter
make lint

# Run unit tests
make test

# Run acceptance tests (requires libvirt)
make testacc
```

On Github, the tests use a hack we have in place to override the domain type (`TF_PROVIDER_LIBVIRT_DOMAIN_TYPE=qemu`), which allows to run acceptance tests without nested virtualization, but using the [tcg](https://www.qemu.org/docs/master/devel/index-tcg.html) accelerator instead of KVM.

### Code quality

All code must pass linting before being committed:

```bash
make check  # Runs lint, vet, and tests
```

Format code with:

```bash
make fmt
```

### Available make targets

Run `make help` to see all targets.

## Contributing

This is early stage development. The focus is on getting core functionality working before accepting contributions.

### Opening issues

Issues should be open for clearly actionable bugs. For getting help on your stack not working, please [use the discussions first](https://github.com/dmacvicar/terraform-provider-libvirt/discussions/categories/q-a).

### Pull Requests

In general, you should not contribute significant features or code without a previous discussion and agreement with the maintainers. This involve transferring code maintenance burden to the maintainers and in general is not desired.

### Use of AI

The author uses AI for this project, but as the maintainer, he owns the outcome and consequences.

If you contribute code or issues and used AI, you have to disclose it, including full details (tools, prompts).

## Author

Duncan Mac-Vicar P.

## License

Same as the legacy provider (Apache 2.0).
