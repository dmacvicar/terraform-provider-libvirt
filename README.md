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

Everything exposed in these resources maps directly to the corresponding libvirt XML; if libvirtxml adds new fields we regenerate and pick them up automatically. Additional libvirt resources (secrets, nodes, interfaces, etc.) may be added in the future.

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
  name         = "example-vm"
  memory       = 512
  memory_unit  = "MiB"
  vcpu         = 1

  os = {
    type         = "hvm"
    type_arch    = "x86_64"
    type_machine = "q35"
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

## Documentation

You can find an overview of the XML <-> HCL mapping and the resource schema in the documentation hosted at the [Terraform registry](https://registry.terraform.io/providers/dmacvicar/libvirt/latest/docs) and [OpenTofu registry](https://search.opentofu.org/provider/dmacvicar/libvirt/latest).

For non-released versions, see [XML <-> HCL mapping](docs/schema-mapping.md) and for an overview of the HCL to XML mapping patterns and [the documentation source](docs/index.md).

## Development information

### Design Principles

- **Schema Coverage**: We support all fields that `libvirt.org/go/libvirtxml` implements from the official libvirt schemas (located at https://gitlab.com/libvirt/libvirt/-/tree/master/src/conf/schemas). If libvirtxml doesn't support a feature yet, neither do we, as we depend on libvirtxml internally.
- **No Abstraction**: The Terraform schema mirrors the libvirt XML structure as closely as possible, providing full access to underlying features rather than simplified abstractions.
- **Preserve intent**: We try to generate state for what the user specified. That is, if the user did not specify something, we let libvirt handle it. We ignore then those when diffing state.

### Schema Mapping

All resources use a single XML <-> HCL mapping spec (flattening rules, unions, presence booleans, nested attributes). See [XML <-> HCL mapping](docs/schema-mapping.md) documentation for the canonical rules.

### Development

This is the first project where I leveraged AI quite heavily not only to do a major cleanup and rewrite of pieces of code, and to implement a new design, but we also use it to inject documentation into the schema.

I am aware of the consequences, advantages and drawbacks. It is my learning platform, and I own the outcome.

### Code Generation

To reduce boilerplate and ensure consistency, this provider uses a code generation system that automatically produces:

- **Terraform models** with proper `tfsdk` tags
- **Plugin Framework schemas** with correct types and optionality
- **XML conversion functions** implementing the "preserve user intent" pattern

See [codegen documentation](internal/codegen/README.md) for architecture and for the documentation tooling (docindex/docgen) used to inject schema descriptions.

### Building and running the provider from source

```bash
git clone https://github.com/dmacvicar/terraform-provider-libvirt
cd terraform-provider-libvirt
make build
```

To install the provider locally:

```bash
make install
```

This installs to `~/.terraform.d/plugins/registry.terraform.io/dmacvicar/libvirt/dev/linux_amd64/`

Then override the provider like this in `$HOME/.terraformrc`:

```hcl
provider_installation {
    dev_overrides {
      "registry.terraform.io/dmacvicar/libvirt" = "/home/duncan/src/terraform-provider-libvirt"
    }

    # For all other providers, install them directly from their origin provider
    # registries as normal.
    direct {}
  }
```

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


## Contributing

### Opening issues

Issues should be open for clearly actionable bugs. For getting help on your stack not working, please [use the discussions first](https://github.com/dmacvicar/terraform-provider-libvirt/discussions/categories/q-a).

### Pull Requests

In general, you should not contribute significant features or code without a previous discussion and agreement with the maintainers. This involve transferring code maintenance burden to the maintainers and in general is not desired.

### Use of AI

The author uses AI for this project, but as the maintainer, he owns the outcome and consequences.

If you contribute code or issues and used AI, you are required to disclose it, including full details (tools, prompts).

## Author

* Duncan Mac-Vicar P.

## License

* Apache 2.0
