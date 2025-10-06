# Terraform provider for libvirt

This provider allows managing libvirt resources (virtual machines, storage pools, networks) using Terraform. It communicates with libvirt using its API to define, configure, and manage virtualization resources.

This is a complete rewrite of the [original provider](https://github.com/dmacvicar/terraform-provider-libvirt).

## Goals

This rewrite improves upon the original provider in several ways:

1. **API Fidelity** - Models the [libvirt XML schemas](https://libvirt.org/format.html) directly instead of abstracting them, giving users full access to libvirt features. Schema coverage is bounded by what [libvirtxml](https://pkg.go.dev/libvirt.org/go/libvirtxml) supports.
2. **Current Framework** - Built with Terraform Plugin Framework, as the SDK v2 used in the original is deprecated
3. **Best Practices** - Follows [HashiCorp's provider design principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)

## Design Principles

- **Schema Coverage**: We support all fields that `libvirt.org/go/libvirtxml` implements from the official libvirt schemas (located at `/usr/share/libvirt/schemas/`). If libvirtxml doesn't support a feature yet, neither do we - we don't create custom XML structs.
- **No Abstraction**: The Terraform schema mirrors the libvirt XML structure as closely as possible, providing full access to underlying features rather than simplified abstractions.
- **User Input Preservation**: For optional+computed fields, we preserve the user's input value even when libvirt normalizes it (e.g., "q35" vs "pc-q35-10.1") to avoid unnecessary diffs.

## XML to HCL Mapping

This provider maps libvirt's XML structure to Terraform's HCL configuration language using a consistent, predictable pattern:

### Mapping Rules

1. **XML Elements → HCL Blocks**
   - Nested XML elements become nested HCL blocks
   - Example: `<os>...</os>` → `os { ... }`

2. **XML Attributes → HCL Attributes**
   - Both XML element attributes and simple text content become HCL attributes
   - Example: `<timer name="rtc" tickpolicy="catchup"/>` → `timer { name = "rtc"; tickpolicy = "catchup" }`

3. **Repeated Elements → HCL Lists**
   - Multiple XML elements of the same type become HCL block lists
   - Example: Multiple `<timer>` elements → `timer { ... }` blocks (can be repeated)

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
  name   = "example-vm"
  type   = "kvm"
  memory = 512  # MiB (unit fixed for simplicity)
  vcpu   = 1

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
  source {
    network   = "default"
    portgroup = "web"
  }
}
```

If the source always has the same pattern, it can be flattened to a simple attribute.

### Notes

- We don't distinguish between XML attributes and elements in HCL - both become HCL attributes
- The same XML structure always maps to the same HCL structure
- This consistent mapping enables automated migration from the old provider or from raw libvirt XML

For detailed XML schemas, see the [libvirt domain format documentation](https://libvirt.org/formatdomain.html).

## Development Approach

Terraform providers are largely scaffolding and domain conversion (Terraform HCL ↔ Provider API). This project leverages AI agents to accelerate development while maintaining code quality through automated linting and testing.

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
  unit   = "MiB"
  vcpu   = 1

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
  }
}
```

See the [examples](./examples) directory for more usage examples.

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

## Current Status

See [TODO.md](./TODO.md) for detailed implementation tracking.

**Completed:**
- ✅ Domain resource with comprehensive configuration
  - OS configuration (type, arch, machine, firmware, boot devices, kernel boot, UEFI loader)
  - Memory and VCPU management with hotplug support
  - Features block (20+ hypervisor features)
  - CPU configuration (mode, match, check, model, vendor)
  - Clock configuration with nested timer blocks and catchup
  - Power management configuration
  - Basic disk support (file-based)
  - Network interfaces (network, bridge, user types)
  - Lifecycle actions (on_poweroff, on_reboot, on_crash)
  - State management with running attribute and create/destroy flags
- ✅ Full CRUD operations with update support
- ✅ Connection to qemu:///system
- ✅ 12 acceptance tests passing
- ✅ Documentation generation

**Planned:**
- 📋 Graphics devices (VNC, Spice)
- 📋 Expanded disk support (driver attributes, network disks)
- 📋 CPU enhancements (topology, features, NUMA)
- 📋 Storage pool and volume resources
- 📋 Network resources
- 📋 Host device passthrough

## Contributing

This is early stage development. The focus is on getting core functionality working before accepting contributions.

## Author

Duncan Mac-Vicar P.

## License

Same as the original provider (Apache 2.0).
