# Terraform provider for libvirt

This provider allows managing libvirt resources (virtual machines, storage pools, networks) using Terraform. It communicates with libvirt using its API to define, configure, and manage virtualization resources.

This is a complete rewrite of the [original provider](https://github.com/dmacvicar/terraform-provider-libvirt).

## Goals

This rewrite improves upon the original provider in several ways:

1. **API Fidelity** - Models the [libvirt XML schemas](https://libvirt.org/format.html) directly instead of abstracting them, giving users full access to libvirt features
2. **Current Framework** - Built with Terraform Plugin Framework, as the SDK v2 used in the original is deprecated
3. **Best Practices** - Follows [HashiCorp's provider design principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)

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
