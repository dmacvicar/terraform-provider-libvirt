# Terraform provider for libvirt

This provider allows managing libvirt resources (virtual machines, storage pools, networks) using Terraform. It communicates with libvirt using its API to define, configure, and manage virtualization resources.

This is a complete rewrite of the [legacy provider](https://github.com/dmacvicar/terraform-provider-libvirt). The legacy provider (v0.8.x and earlier) is maintained in the [v0.8 branch](https://github.com/dmacvicar/terraform-provider-libvirt/tree/v0.8). Starting from v0.9.0, all releases will be based on this new rewrite.

## Goals

This rewrite improves upon the legacy provider in several ways:

1. **API Fidelity** - Models the [libvirt XML schemas](https://libvirt.org/format.html) directly instead of abstracting them, giving users full access to libvirt features. Schema coverage is bounded by what [libvirtxml](https://pkg.go.dev/libvirt.org/go/libvirtxml) supports.
2. **Current Framework** - Built with Terraform Plugin Framework, as the SDK v2 used in the legacy provider is deprecated
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
        type  = "network"
        model = "virtio"
        source = {
          network = "default"
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
- This consistent mapping enables automated migration from the legacy provider or from raw libvirt XML
- **Nested Attributes vs Blocks**: Following [HashiCorp's guidance](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/blocks), new features use nested attributes (e.g., `devices = { ... }`) instead of blocks. Some existing features (`os`, `features`, `clock`, etc.) incorrectly use blocks and need conversion (see TODO.md).

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

This table shows implementation status and compatibility with the [legacy provider](https://github.com/dmacvicar/terraform-provider-libvirt/tree/v0.8) (v0.8.x):

### Provider Configuration

| Feature | Status | Legacy Provider | Notes |
|---------|--------|--------------|-------|
| qemu:///system | ✅ | ✅ | Local system connection |
| qemu:///session | ✅ | ✅ | Local user session connection |
| qemu+ssh:// | ✅ | ✅ | SSH transport (Go SSH library) |
| qemu+sshcmd:// | ✅ | ✅ | SSH transport (native command) |
| qemu+tcp:// | ✅ | ✅ | TCP transport |
| qemu+tls:// | ✅ | ✅ | TLS transport |

### Domain Resource (libvirt_domain)

| Feature Category | Status | Legacy Provider | Notes |
|-----------------|--------|--------------|-------|
| Basic config | ✅ | ✅ | name, memory, vcpu, type, description |
| Metadata | ○ | ✅ | Custom metadata XML |
| OS & boot | ✅ | ✅ | type, arch, machine, firmware, boot devices |
| Kernel boot | ✅ | ✅ | kernel, initrd, cmdline |
| CPU | ⚠️ | ⚠️ | Basic (mode) only; topology/features planned |
| Memory | ⚠️ | ⚠️ | Basic only; hugepages planned |
| Features | ✅ | ⚠️ | 20+ features; more than legacy provider |
| Clock & timers | ✅ | ○ | Full support including nested catchup |
| Power management | ✅ | ○ | suspend_to_mem, suspend_to_disk |
| Disks (basic) | ✅ | ✅ | File-based disks with device, target, bus |
| Disks (volume) | ✅ | ✅ | Nested `source` with pool/volume reference |
| Disks (driver) | ○ | ⚠️ | cache, io, discard options |
| Disks (URL) | ○ | ✅ | URL download support |
| Disks (block) | ○ | ✅ | Block device passthrough |
| Disks (SCSI) | ○ | ✅ | SCSI bus, WWN identifier |
| Network (basic) | ✅ | ✅ | network, bridge types |
| Network (user) | ✅ | ✅ | User-mode networking |
| Network (macvtap) | ○ | ✅ | macvtap, vepa, passthrough modes |
| Network (wait_for_lease) | ○ | ✅ | Wait for DHCP lease |
| Graphics | ✅ | ✅ | VNC/Spice display (autoport, listen, port) |
| Video | ○ | ✅ | Video device (cirrus, etc.) |
| Console/Serial | ○ | ✅ | Console and serial devices |
| Filesystem (9p) | ○ | ✅ | Host directory sharing via virtio-9p |
| TPM | ○ | ✅ | TPM device emulation |
| NVRAM | ⚠️ | ✅ | Basic UEFI loader; template support planned |
| State management | ✅ | ✅ | running attribute |
| Autostart | ○ | ✅ | Start domain on host boot |
| Cloud-init | ○ | ✅ | libvirt_cloudinit_disk resource |
| CoreOS Ignition | ○ | ✅ | libvirt_ignition resource |
| Combustion | ○ | ✅ | libvirt_combustion resource |
| QEMU agent | ○ | ✅ | Integration with qemu-guest-agent |
| XML XSLT | ○ | ✅ | XSLT transforms for custom XML |

### Volume Resource (libvirt_volume)

| Feature | Status | Legacy Provider | Notes |
|---------|--------|--------------|-------|
| Resource | ✅ | ✅ | Create and manage volumes |
| Type | ✅ | ✅ | Volume type (file, block, dir, etc.) |
| Format | ✅ | ✅ | qcow2, raw format support |
| Backing volumes | ✅ | ✅ | `backing_store` applies when creating volumes |
| Permissions | ✅ | ✅ | owner, group, mode, label |
| URL download | ✅ | ✅ | Download via create.content.url (HTTPS + local files) |
| XML XSLT | ○ | ✅ | XSLT transforms |

> libvirt’s `<backingStore>` element on domain disks is informational unless the hypervisor advertises `backingStoreInput`. The provider therefore configures copy-on-write overlays only on `libvirt_volume` resources; domain-level `backing_store` inputs are intentionally not exposed.

### Pool Resource (libvirt_pool)

| Feature | Status | Legacy Provider | Notes |
|---------|--------|--------------|-------|
| Resource | ✅ | ✅ | Create and manage storage pools |
| Pool types | ✅ | ✅ | dir (directory) type |
| Target permissions | ✅ | ✅ | owner, group, mode, label |
| Source | ✅ | ✅ | name, device (for LVM) |
| Logical pools | ⚠️ | ✅ | Partial - needs testing |

### Network Resource (libvirt_network)

| Feature | Status | Legacy Provider | Notes |
|---------|--------|--------------|-------|
| Resource | ✅ | ✅ | Create and manage networks |
| Network modes | ⚠️ | ✅ | nat and isolated (none) modes implemented |
| IP addresses | ✅ | ✅ | CIDR configuration (e.g., 10.17.3.0/24) |
| Autostart | ✅ | ✅ | Start network on host boot |
| DHCP | ○ | ✅ | DHCP ranges and static hosts (deferred) |
| DNS | ○ | ✅ | DNS hosts, forwarders, SRV records (deferred) |
| Routes | ○ | ✅ | Static routes (deferred) |
| Dnsmasq options | ○ | ✅ | Custom dnsmasq configuration (deferred) |

### Data Sources

| Feature | Status | Legacy Provider | Notes |
|---------|--------|--------------|-------|
| Node info | ○ | ✅ | Host system information (CPU, memory) |
| Node devices | ○ | ✅ | Device enumeration by capability |
| Node device info | ○ | ✅ | Detailed device information (PCI, USB, etc.) |
| Network lookup | ○ | ✅ | Lookup existing networks (deferred) |
| Network templates | ○ | ✅ | DNS/dnsmasq templates (deferred - use HCL instead) |

**Legend:**
- ✅ Fully implemented
- ⚠️ Partially implemented
- ○ Not yet implemented

See [TODO.md](./TODO.md) for detailed implementation tracking

## Contributing

This is early stage development. The focus is on getting core functionality working before accepting contributions.

## Author

Duncan Mac-Vicar P.

## License

Same as the legacy provider (Apache 2.0).
