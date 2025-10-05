
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

## XML to HCL Mapping Patterns

### General Mapping Rules

1. **XML Elements → HCL Blocks**
   - Nested XML elements become nested HCL blocks
   - Example: `<os>...</os>` → `os { ... }`

2. **XML Attributes → HCL Attributes**
   - XML element attributes become HCL attributes within their block
   - Example: `<timer name="rtc" tickpolicy="catchup"/>` → `timer { name = "rtc"; tickpolicy = "catchup" }`

3. **Repeated Elements → HCL Lists**
   - Multiple XML elements of the same type become repeated HCL blocks
   - Example: Multiple `<timer>` elements → multiple `timer { }` blocks

### Elements with Text Content + Attributes

For XML elements with both text content and attributes, see the "Handling Elements with Text Content and Attributes" section in README.md.

**Quick reference:**
- Unit only → flatten with fixed unit: `memory = 512`
- Unit + 1 other → flatten both: `max_memory = 2048; max_memory_slots = 16`
- Multiple attributes → nested block: `vcpu { value = 4; placement = "static"; cpuset = "0-3" }`
- Type-dependent source → nested block: `source { network = "default" }`

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

1. **Check TODO.md for current tasks** - single source of truth for what needs to be done
2. **Keep TODO.md updated** - as you complete tasks, mark items as done and update the "Current Status" section
3. **Do NOT create random documentation files** - use existing files (TODO.md, README.md, AGENTS.md, DOMAIN_COVERAGE_ANALYSIS.md)
4. **Run `make lint` before committing** - all code must pass linting
5. **Run `make fmt` to format code** - use standard Go formatting
6. **Preserve design principles** - schema must follow libvirt XML closely
7. **Reference old provider minimally** - mainly for connection handling and test ideas
8. **Track progress continuously** - update TODO.md after completing each feature or task

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
   - **DO NOT add promotional text, links, or "Generated with" messages**

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

## Critical Pattern: Preserve User Intent

**Problem**: Libvirt often sets default values for optional fields. If we naively read back all values from libvirt XML, Terraform will detect a diff between what the user specified (null) and what libvirt returned (a default value).

**Solution**: Only populate fields in the model during XML→model conversion if the user originally specified them.

### Example

```go
// ❌ WRONG - causes "inconsistent result" errors
if domain.OnPoweroff != "" {
    model.OnPoweroff = types.StringValue(domain.OnPoweroff)
}

// ✅ CORRECT - only set if user originally specified it
if !model.OnPoweroff.IsNull() && !model.OnPoweroff.IsUnknown() && domain.OnPoweroff != "" {
    model.OnPoweroff = types.StringValue(domain.OnPoweroff)
}
```

### When to Apply This Pattern

Apply this for **optional** fields where libvirt provides defaults:
- Lifecycle actions (on_poweroff, on_reboot, on_crsh - libvirt defaults to "destroy"/"restart")
- Current memory (libvirt defaults to same as memory)
- Boot devices (libvirt may add default boot order)
- Machine type (libvirt may expand "q35" to "pc-q35-10.1")

### When NOT to Apply

Don't apply for:
- **Required** fields (name, memory, vcpu)
- **Computed** fields that should always reflect libvirt's state (uuid, id)
- Fields where libvirt doesn't set defaults

### Testing for This Issue

If you see errors like:
```
Error: Provider produced inconsistent result after apply
When applying changes to libvirt_domain.test, provider produced an
unexpected new value: .on_poweroff: was null, but now cty.StringVal("destroy")
```

This means you need to add the user-input check to that field's XML→model conversion.
