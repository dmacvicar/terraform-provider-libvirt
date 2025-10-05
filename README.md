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

### Key Points

- **Flattening**: We don't distinguish between XML attributes and elements in HCL - both become HCL attributes
- **Consistency**: The same XML structure always maps to the same HCL structure
- **Readability**: HCL blocks follow Terraform conventions for better readability
- **Pragmatic Exceptions**: Some XML patterns (like `<memory unit="MiB">512</memory>`) are simplified to single values with fixed units for better UX
- **Migration**: This consistent mapping enables automated migration from the old provider or from raw libvirt XML

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

**Working:**
- ✅ Domain resource with basic OS configuration
- ✅ Memory and VCPU management
- ✅ Connection to qemu:///system
- ✅ Acceptance tests passing

**In Progress:**
- ⏳ Device support (disks, networks, graphics)
- ⏳ Additional connection types
- ⏳ Documentation generation

**Planned:**
- 📋 Storage pool and volume resources
- 📋 Network resources
- 📋 Additional domain features (CPU, features, etc.)

## Contributing

This is early stage development. The focus is on getting core functionality working before accepting contributions.

## Author

Duncan Mac-Vicar P.

## License

Same as the original provider (Apache 2.0).
