
# Context for AI Assistants

## Project Overview

This project is a complete v2 rewrite of the Terraform provider for libvirt to replace https://github.com/dmacvicar/terraform-provider-libvirt.

## Core Principles

1. **Close API Modeling**: Follow https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles especially around modelling the underlying API closely. This means the provider will not try to simplify things but follow the libvirt XML schema closely. See https://libvirt.org/format.html

2. **Modern Framework**: Use the https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider instead of the v2 SDK

3. **Proven Client**: Continue using the pure Go libvirt client from DigitalOcean (github.com/digitalocean/go-libvirt)

## Schema Coverage Policy

**We support the full libvirt XML schemas as implemented by `libvirt.org/go/libvirtxml`.**

- **Source of Truth**: The official libvirt XML schemas located at `/usr/share/libvirt/schemas/` (domain.rng, network.rng, storagepool.rng, etc.)
- **Implementation Boundary**: If libvirtxml has not added support for a schema element, we will not support it yet. We do not create custom XML structs.
- **Coverage Goal**: Expose all fields that libvirtxml supports in our Terraform schemas, maintaining full API fidelity
- **When libvirtxml is missing features**: Document the gap and consider contributing upstream to libvirtxml rather than working around it

This ensures we provide comprehensive libvirt feature coverage while maintaining clean, maintainable code that leverages the official XML marshaling library.

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
- **Libvirt XML Schemas**:
  - Online docs: https://libvirt.org/format.html
  - Local RNG schemas: `/usr/share/libvirt/schemas/` (domain.rng, network.rng, storagepool.rng, etc.)
  - These are the source of truth for what features exist
- **libvirtxml Library**: https://pkg.go.dev/libvirt.org/go/libvirtxml - defines what we can support
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

## Development Workflow - Work Incrementally

**IMPORTANT**: Always work field-by-field or feature-by-feature with commits in between. Never implement multiple complex features in one iteration.

### The Pattern: Add → Test → Commit → Repeat

1. **Add ONE field or small group of related fields**
   - Update model struct
   - Add schema definition
   - Implement conversion functions (model ↔ XML)

2. **Verify it works**
   - `make lint` - must pass with 0 issues
   - `make build` - must compile
   - `make testacc` - acceptance tests must pass

3. **Commit immediately**
   - Small, focused commit message
   - Example: `feat: add title and description fields`
   - Keep it simple - avoid verbose explanations

4. **Repeat for next field**

### What NOT to Do

- ❌ Don't implement 10+ fields at once
- ❌ Don't batch multiple commits
- ❌ Don't skip testing between changes
- ❌ Don't write verbose commit messages explaining everything
- ❌ Don't say "all tests passing" or obvious statements in commits

### Examples

**Good approach:**
```
Commit 1: feat: add title and description fields
Commit 2: feat: add lifecycle action fields
Commit 3: feat: add iothreads field
Commit 4: fix: preserve user input for optional fields with defaults
```

**Bad approach:**
```
Commit 1: feat: add title, description, lifecycle, iothreads, memory fields
  - Added 15 new fields
  - Updated all schemas
  - Implemented conversions
  - All tests passing
  - Linting clean
  (Then discover 20 compilation errors and test failures)
```

### Why This Matters

- Easier to review individual changes
- Faster to identify and fix issues
- Each commit is a working state
- Can revert problematic changes without losing good work
- User can see steady progress

## Key Gotchas

- The old provider simplified the libvirt API - we explicitly do NOT want that
- Plugin Framework uses different patterns than SDK v2 - check framework docs
- Libvirt XML schemas are complex - expect nested structures with many optional fields
- **Use libvirtxml library**: Don't create custom XML structs - use `libvirt.org/go/libvirtxml`
- Connection management is tricky - see old provider for proven patterns
- Testing requires libvirt daemon - tests should be skippable in CI if needed
- Libvirt normalizes values (e.g., "q35" → "pc-q35-10.1") - preserve user input to avoid diffs
