
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

## Feature Implementation Priority

**CRITICAL: Read this section before implementing any features.**

When deciding what to implement next, follow this priority order:

### Priority 1: Pure Libvirt Features from Old Provider
**These are the highest priority features to implement.**

- Features that existed in the old provider (github.com/dmacvicar/terraform-provider-libvirt)
- AND are part of native libvirt functionality
- NOT provider-specific additions or conveniences

**Examples of Priority 1 (implement these):**
- Console/Serial devices (libvirtxml.DomainConsole, DomainSerial)
- Video devices (libvirtxml.DomainVideo)
- TPM devices (libvirtxml.DomainTPM)
- Emulator path (libvirtxml.DomainDeviceList.Emulator)
- SCSI disks with WWN (libvirtxml.DomainDisk)
- Block device disks (libvirtxml.DomainDisk with type="block")
- Direct network types: macvtap, vepa, passthrough (libvirtxml.DomainInterface)
- NVRAM template (libvirtxml.DomainLoader.Template)
- Metadata custom XML (libvirtxml.Domain.Metadata)
- RNG devices (libvirtxml.DomainRNG)
- Input devices (libvirtxml.DomainInput)

**Examples of NOT Priority 1 (defer these):**
- ❌ Cloud-init support (libvirt_cloudinit_disk resource) - provider addition, not libvirt
- ❌ URL download for volumes (source = "http://...") - provider convenience, not libvirt
- ❌ CoreOS Ignition (libvirt_ignition resource) - provider addition
- ❌ Combustion support - provider addition
- ❌ XML XSLT transforms - provider addition, against design principles

### Priority 2: Advanced Libvirt Features
- CPU topology, features, NUMA (libvirtxml supports)
- Memory backing, hugepages (libvirtxml supports)
- Advanced features blocks (HyperV, KVM, SMM)
- Host device passthrough (libvirtxml.DomainHostdev)
- Advanced tuning (CPUTune, NUMATune, BlockIOTune)

### Priority 3: Provider Conveniences (Maybe Never)
- Cloud-init integration
- URL download
- XSLT transforms
- Other abstractions on top of libvirt

**Why this order matters:**
- Users migrating from the old provider expect feature parity with pure libvirt features
- Provider-specific conveniences can be built later (or never, if they conflict with design principles)
- Focus on exposing libvirt's full API first, conveniences second

**When in doubt:** Check if the feature exists in libvirtxml and was in the old provider. If yes to both and it's not a provider addition, it's Priority 1.

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

### Provider Schema Implementation Notes

**Use Nested Attributes, Not Blocks**

Per [HashiCorp guidance](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/blocks), new provider implementations should use nested attribute types instead of block types. Blocks are mainly for migrating legacy SDK-based providers.

**Why nested attributes?**
- Best practice for new providers per Terraform Plugin Framework documentation
- More explicit syntax with `=` and array brackets `[...]`
- Better type safety and validation
- Consistent with modern Terraform providers

**Implementation:**
- Use `schema.SingleNestedAttribute` for single objects (e.g., `target = { path = "/data" }`)
- Use `schema.ListNestedAttribute` for lists of objects (e.g., `disks = [{ target = "vda" }]`)
- Model fields should be `types.Object` or `types.List` accordingly
- Use `.As(ctx, &model, basetypes.ObjectAsOptions{})` to extract from `types.Object`
- Use `.ElementsAs(ctx, &array, false)` to extract from `types.List`

**Current State:**
- **Devices** (`devices`): Uses nested attributes ✅
  - `devices = { disks = [...], interfaces = [...] }`
- **Existing blocks** (`os`, `features`, `cpu`, `clock`, `pm`, `create`, `destroy`): Still use blocks for ergonomics
  - These may be converted later, but it's not urgent
- **New features**: Should use nested attributes

**Follow the libvirt XML Schema Structure**

The Terraform schema must mirror the libvirt XML structure exactly. This is critical for correctness.

Example:
- XML: `<domain><devices><disk>...</disk><interface>...</interface></devices></domain>`
- HCL: `devices = { disks = [...], interfaces = [...] }`

**Key principle**: If libvirt XML has a container element (like `<devices>`), we must have a corresponding container in HCL.

**Preserve User Intent Pattern**

**Rule**: Only populate optional fields in state if the user explicitly specified them in their configuration.

**Why**: Libvirt sets defaults for many optional fields. If we naively read all values back, Terraform detects drift between null (user didn't specify) and libvirt's default, causing unwanted plan diffs.

**Apply to**: Optional fields where libvirt provides defaults (on_poweroff, current_memory, boot_devices, autostart, unit, type, etc.)

**Don't apply to**: Required fields (name, memory) or purely computed fields (uuid, id)

Example:
```go
// ❌ WRONG
if domain.OnPoweroff != "" {
    model.OnPoweroff = types.StringValue(domain.OnPoweroff)
}

// ✅ CORRECT
if !model.OnPoweroff.IsNull() && !model.OnPoweroff.IsUnknown() && domain.OnPoweroff != "" {
    model.OnPoweroff = types.StringValue(domain.OnPoweroff)
}
```

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
4. **NEVER create new .md files without explicit user authorization** - ask first before creating any documentation
5. **NEVER install software or modify system configuration** - only work within the source directories. If dependencies are missing, inform the user.
6. **NEVER use sudo or any system administration commands** - no system modifications, no service restarts, no package installs
7. **Run `make lint` before committing** - all code must pass linting
8. **Run `make fmt` to format code** - use standard Go formatting
9. **Preserve design principles** - schema must follow libvirt XML closely
10. **Reference old provider minimally** - mainly for connection handling and test ideas
11. **Track progress continuously** - update TODO.md after completing each feature or task

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

## Testing Patterns

### Acceptance Tests

Acceptance tests verify the provider works against real libvirt infrastructure. They use the Terraform Plugin Testing framework.

**Test Structure:**
```go
func TestAccDomainResource_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckDomainDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccDomainResourceConfigBasic("test-domain-basic"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("libvirt_domain.test", "name", "test-domain-basic"),
                ),
            },
        },
    })
}
```

**Key Fields:**
- `PreCheck` - Verify prerequisites (libvirt available)
- `ProtoV6ProviderFactories` - Provider factory for Plugin Framework
- `CheckDestroy` - Verify resources cleaned up after test
- `Steps` - Test steps (create, update, delete)

### Test Sweepers

Test sweepers clean up leaked resources from failed tests. They're especially useful when tests fail before reaching the destroy phase.

**Setup (required once):**

Add to `provider_test.go`:
```go
func TestMain(m *testing.M) {
    resource.TestMain(m)
}
```

**Register sweepers** in resource test files:
```go
func init() {
    resource.AddTestSweepers("libvirt_domain", &resource.Sweeper{
        Name: "libvirt_domain",
        F: func(uri string) error {
            ctx := context.Background()
            client, err := libvirtclient.NewClient(ctx, uri)
            if err != nil {
                return fmt.Errorf("failed to create libvirt client: %w", err)
            }
            defer client.Close()

            // List all domains
            domains, err := client.Libvirt().Domains()
            if err != nil {
                return err
            }

            // Delete test domains (prefix: test-)
            for _, domain := range domains {
                if strings.HasPrefix(domain.Name, "test-") {
                    _ = client.Libvirt().DomainDestroy(domain)
                    _ = client.Libvirt().DomainUndefine(domain)
                }
            }
            return nil
        },
    })
}
```

**Dependencies:** Sweepers can specify dependencies to ensure cleanup order:
```go
resource.AddTestSweepers("libvirt_domain", &resource.Sweeper{
    Name: "libvirt_domain",
    Dependencies: []string{"libvirt_volume"},  // Clean volumes first
    F: func(uri string) error { /* ... */ },
})
```

**Running sweepers:**
```bash
# List available sweepers
go test -sweep-list

# Run all sweepers for qemu:///system
go test -sweep=qemu:///system

# Run specific sweeper
go test -sweep=qemu:///system -sweep-run=libvirt_domain

# Add to Makefile
make sweep  # Should run: go test -sweep=qemu:///system -timeout 10m
```

**Best Practices:**
- Prefix all test resources with `test-` for easy identification
- Only delete resources matching test prefix
- Handle errors gracefully (sweeper failures shouldn't block other sweepers)
- Register sweepers for resources that may leak (domains, volumes, networks, pools)
- Set dependencies to ensure proper cleanup order

**Important Notes:**
- The string parameter (e.g., "qemu:///system") is provider-specific, NOT just for AWS regions
- For libvirt, it's the connection URI
- Sweepers run manually, not automatically on test failure
- Run sweepers before test runs to ensure clean state, or after to cleanup failures

## Key Gotchas

- The old provider simplified the libvirt API - we explicitly do NOT want that
- Plugin Framework uses different patterns than SDK v2 - check framework docs
- Libvirt XML schemas are complex - expect nested structures with many optional fields
- **Use libvirtxml library**: Don't create custom XML structs - use `libvirt.org/go/libvirtxml`
- Connection management is tricky - see old provider for proven patterns
- Testing requires libvirt daemon - tests should be skippable in CI if needed
- Libvirt normalizes values (e.g., "q35" → "pc-q35-10.1") - preserve user input to avoid diffs
- **Use go-libvirt constants**: Never use magic numbers for libvirt enums - always use the proper constants from `golibvirt`

## Critical Pattern: Use libvirt Constants

**Problem**: Using magic numbers (like `1` for "running" state) makes code unreadable and error-prone.

**Solution**: Always use the proper constants from the `go-libvirt` package (imported as `golibvirt`).

### Import Pattern

```go
import (
    golibvirt "github.com/digitalocean/go-libvirt"
)
```

### Examples

#### Domain States

```go
// ❌ WRONG - magic numbers
if state == 1 {  // What does 1 mean?
    // ...
}

// ✅ CORRECT - use constants
if uint32(state) == uint32(golibvirt.DomainRunning) {
    // ...
}
```

#### Domain Creation Flags

```go
// ❌ WRONG - magic numbers
var flags uint32 = 0
if paused {
    flags |= 1  // What flag is this?
}
if autodestroy {
    flags |= 2  // And this?
}

// ✅ CORRECT - use constants
var flags uint32 = 0
if paused {
    flags |= uint32(golibvirt.DomainStartPaused)
}
if autodestroy {
    flags |= uint32(golibvirt.DomainStartAutodestroy)
}
```

### Common Constants

**Domain States:**
- `golibvirt.DomainRunning` - Domain is running
- `golibvirt.DomainShutoff` - Domain is shut off
- `golibvirt.DomainPaused` - Domain is paused
- `golibvirt.DomainCrashed` - Domain has crashed

**Domain Start Flags:**
- `golibvirt.DomainStartPaused` - Start domain in paused state
- `golibvirt.DomainStartAutodestroy` - Destroy domain on client disconnect
- `golibvirt.DomainStartBypassCache` - Bypass file system cache
- `golibvirt.DomainStartForceBoot` - Force boot, even if saved state exists
- `golibvirt.DomainStartValidate` - Validate the XML before starting
- `golibvirt.DomainStartResetNvram` - Reset NVRAM on boot

### Type Casting

The `go-libvirt` library uses various integer types. Always cast to the appropriate type:

```go
// DomainGetState returns int32
state, _, err := client.Libvirt().DomainGetState(domain, 0)

// Cast to uint32 for comparison with constants
if uint32(state) == uint32(golibvirt.DomainRunning) {
    // ...
}
```

### Where to Find Constants

Check the `go-libvirt` package documentation or source code:
- https://pkg.go.dev/github.com/digitalocean/go-libvirt
- Look for const declarations matching the libvirt C API enums
