
# Context for AI Assistants

## Project Overview

This project is a complete v2 rewrite of the Terraform provider for libvirt to replace https://github.com/dmacvicar/terraform-provider-libvirt.

## Core Principles

1. **Close API Modeling**: Follow https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles especially around modelling the underlying API closely. This means the provider will not try to simplify things but follow the libvirt XML schema closely. See https://libvirt.org/format.html

2. **Modern Framework**: Use the https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider instead of the v2 SDK

3. **Proven Client**: Continue using the pure Go libvirt client from DigitalOcean (github.com/digitalocean/go-libvirt)

## Project Structure

```
.
├── AGENTS.md                 # This file - context for AI assistants
├── README.md                 # Project status, roadmap, and TODO tracking
├── internal/
│   ├── provider/            # Provider implementation
│   ├── resources/           # Resource implementations
│   ├── datasources/         # Data source implementations
│   └── libvirt/             # Libvirt client wrapper
├── examples/                 # Usage examples
└── docs/                     # Generated documentation
```

## Important References

- **Old Provider**: Located at ../terraform-provider-libvirt - reference for test cases and connection logic, but DO NOT copy the schema design
- **Libvirt XML Schemas**: https://libvirt.org/format.html - the source of truth for resource schemas
- **Plugin Framework Examples**: https://github.com/hashicorp/terraform-provider-scaffolding-framework
- **Good Plugin Framework Providers**: Look at terraform-provider-docker, terraform-provider-kubernetes for patterns

## Technical Decisions Made

1. **Initial Connection Support**: Start with `qemu:///system`, then port other transports from old provider
2. **Resource Priority**: Domain (VM) → Storage → Network
3. **Go Version**: 1.21+ (for modern Plugin Framework support)
4. **Schema Design**: Hand-crafted Terraform schemas, but use official libvirtxml for XML marshaling
5. **XML Library**: Using `libvirt.org/go/libvirtxml` for all XML operations instead of custom structs
6. **Testing**: Port test cases from old provider where applicable, adapt to new schema
7. **Computed Fields**: Preserve user input for optional+computed fields (machine type, boot devices) to avoid unnecessary diffs

## Current State

Check README.md for current implementation status and the roadmap. The README contains:
- Implementation roadmap with checkboxes
- Current phase and next steps
- Pending technical decisions
- Questions for future sessions

## Working with This Project

1. **Always check README.md first** - it tracks the current state and TODOs
2. **Update README.md** - as you complete tasks, update the checkboxes and status
3. **Run `make lint` before committing** - all code must pass linting
4. **Run `make fmt` to format code** - use standard Go formatting
5. **Preserve design principles** - schema must follow libvirt XML closely
6. **Reference old provider minimally** - mainly for connection handling and test ideas
7. **Document decisions** - add technical decisions to README.md as they're made

## Key Gotchas

- The old provider simplified the libvirt API - we explicitly do NOT want that
- Plugin Framework uses different patterns than SDK v2 - check framework docs
- Libvirt XML schemas are complex - expect nested structures with many optional fields
- **Use libvirtxml library**: Don't create custom XML structs - use `libvirt.org/go/libvirtxml`
- Connection management is tricky - see old provider for proven patterns
- Testing requires libvirt daemon - tests should be skippable in CI if needed
- Libvirt normalizes values (e.g., "q35" → "pc-q35-10.1") - preserve user input to avoid diffs
