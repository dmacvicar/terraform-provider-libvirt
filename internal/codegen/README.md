# Code Generator for terraform-provider-libvirt

This directory contains the code generation system that automatically produces Terraform Plugin Framework schemas, Go models, and XML conversion code from libvirtxml structs.

## Table of Contents

- [Overview](#overview)
- [Special Patterns](#special-patterns)
- [Architecture](#architecture)
- [Generation Flow](#generation-flow)
- [Components](#components)
- [Usage](#usage)
- [Generated Code Structure](#generated-code-structure)
- [Extending the Generator](#extending-the-generator)
- [Troubleshooting](#troubleshooting)

## Current Status (2025-01-12)

**✅ Working:** 324 structs generated, all compile. Value+unit flattening works (`capacity` + `capacity_unit`).
**⏸️ Next:** Migrate pool resource to use 100% generated code and test.
**📋 See:** `/TODO.md` for detailed status and next steps.

## Overview

The code generator eliminates manual boilerplate by automatically creating:

1. **Terraform Models** - Structs with `tfsdk` tags for Plugin Framework
2. **Schema Definitions** - Plugin Framework schema with proper types and optionality
3. **Conversion Functions** - Bidirectional XML ↔ Model conversion with user intent preservation

## Mapping Specification

Every generated field follows a deterministic mapping so Terraform HCL mirrors the libvirt XML exactly. This is the canonical spec that all docs, examples, and tests must follow:

1. **Elements → Nested Attributes**: Each XML element becomes a nested attribute object with `schema.SingleNestedAttribute`/`schema.ListNestedAttribute`. Attributes on those XML elements become sibling attributes in that nested object.
2. **Repeated elements → Lists**: Multiple occurrences of the same XML element become Terraform lists of nested objects. Simple lists (`[]string`, `[]int`, etc.) map to `schema.ListAttribute`.
3. **Attributes → Scalars**: XML attributes stay as scalar attributes in the containing object (strings, bools, integers). Optional attributes use `Optional: true` and respect the user-intent-preservation logic described in AGENTS.md.
4. **Value-with-unit patterns**: Elements with chardata + attributes flatten into sibling attributes:
   - Value only → `memory = 512`
   - Value + one attribute → `max_memory = 2048`, `max_memory_slots = 16`
   - Value + 2+ attrs → stay nested (e.g., `<vcpu>` becomes `vcpu = { value = 4, placement = "static", ... }`)
   The flattening behavior is covered by `TestMemoryValueWithUnitRoundtrip` in `internal/generated/conversion_test.go`.
5. **Boolean-to-presence**: Elements that use presence semantics (e.g., `<acpi/>`) map to Terraform booleans. `true` produces an empty XML element, `false/null` omits it. Tested by `TestBooleanPresenceRoundtrip`.
6. **Type-dependent sources / variants**: Union types stay as nested attributes with mutually exclusive fields (e.g., disk `source = { file = "...", block = "...", volume = { ... } }`). Tests should assert that unknown/omitted variants stay untouched.
7. **Optional + Computed semantics**: Optional fields only populate on reads if the user set them in the plan. Computed-only fields always pull from XML. This is validated by `TestPreserveUserIntent`.

Whenever you add a new pattern, update this section and the associated tests so the mapping stays self-documenting.

## Special Patterns

The generator detects and handles several patterns found in libvirtxml structs:

### 1. Chardata with Attributes (Implemented - Always Flattened)

**Pattern:** Value INSIDE element (chardata), ALL other fields are attributes → FLATTEN

**Rule:** If a struct has chardata and all other fields are XML attributes (no nested elements),
flatten to separate top-level fields: `{element}` for the value, `{element}_{attr}` for each attribute.

**XML Examples:**
```xml
<memory unit='KiB'>524288</memory>                    → memory, memory_unit
<vcpu placement='static' cpuset='1-4' current='2'>4</vcpu>  → vcpu, vcpu_placement, vcpu_cpuset, vcpu_current
```

**Code example:**
```go
// libvirtxml
type DomainMemory struct {
    Value uint   `xml:",chardata"`
    Unit  string `xml:"unit,attr"`
}

// Generated: ALL attributes flattened
memory      = 524288  // chardata value
memory_unit = "KiB"   // unit attribute
```

```go
// libvirtxml
type DomainMaxMemory struct {
    Value uint   `xml:",chardata"`
    Unit  string `xml:"unit,attr"`
    Slots uint   `xml:"slots,attr,omitempty"`
}

// Generated: ALL attributes flattened
maximum_memory       = 2048
maximum_memory_unit  = "KiB"
maximum_memory_slots = 16
```

```go
// libvirtxml
type DomainVCPU struct {
    Value     uint   `xml:",chardata"`
    Placement string `xml:"placement,attr,omitempty"`
    CPUSet    string `xml:"cpuset,attr,omitempty"`
    Current   uint   `xml:"current,attr,omitempty"`
}

// Generated: ALL attributes flattened
vcpu            = 4
vcpu_placement  = "static"
vcpu_cpuset     = "0-3"
vcpu_current    = 2
```

### 2. Type-Dependent Source

Nested objects where exactly one variant must be set.

```go
// libvirtxml
type DomainDiskSource struct {
    File   *DomainDiskSourceFile   `xml:"file"`
    Block  *DomainDiskSourceBlock  `xml:"block"`
    Volume *DomainDiskSourceVolume `xml:"volume"`
}

// Generated model
type DomainDiskSourceModel struct {
    File   types.String `tfsdk:"file"`    // Standalone
    Block  types.String `tfsdk:"block"`   // Standalone
    Pool   types.String `tfsdk:"pool"`    // With volume
    Volume types.String `tfsdk:"volume"`  // With pool
}

// Usage (pick ONE variant)
source = { file = "/path/to/disk.qcow2" }
source = { block = "/dev/sda1" }
source = { pool = "default", volume = "vol1" }
```

### 3. Boolean to Presence

Boolean in TF becomes presence/absence in XML.

```go
// libvirtxml
type DomainFeatureList struct {
    ACPI *DomainFeature `xml:"acpi"`  // Empty struct
    PAE  *DomainFeature `xml:"pae"`
}
type DomainFeature struct{}

// Generated
features = {
  acpi = true   // Creates <acpi/> element
  pae  = false  // Omits element
}
```

### 4. Preserve User Intent

Optional fields only populated if user specified them, preventing drift from libvirt defaults.

```go
// Model to XML - only set if user provided
if !model.OnPoweroff.IsNull() && !model.OnPoweroff.IsUnknown() {
    domain.OnPoweroff = model.OnPoweroff.ValueString()
}

// XML to Model - preserve from plan
if plan != nil && !plan.OnPoweroff.IsNull() {
    model.OnPoweroff = types.StringValue(domain.OnPoweroff)
} else {
    model.OnPoweroff = types.StringNull()
}
```

### 5. Variant Objects

Parent contains mutually exclusive nested objects.

```go
// libvirtxml
type DomainGraphics struct {
    VNC   *DomainGraphicsVNC   `xml:"vnc"`
    Spice *DomainGraphicsSpice `xml:"spice"`
}

// Generated - only one should be set
graphics = {
  vnc = { port = 5900 }
}
```

### 6. String to Yes/No

Boolean in TF becomes "yes"/"no" string in XML.

```go
// libvirtxml
type DomainLoader struct {
    Readonly string `xml:"readonly,attr"`  // "yes" or "no"
}

// Generated
loader_readonly = true  // Converts to "yes"
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Code Generation System                       │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  │
                    ┌─────────────▼─────────────┐
                    │    main.go (Orchestrator) │
                    └─────────────┬─────────────┘
                                  │
                ┌─────────────────┼─────────────────┐
                │                 │                 │
       ┌────────▼────────┐ ┌─────▼──────┐ ┌────────▼────────┐
       │  LibvirtXML     │ │    RNG     │ │   Generation    │
       │   Reflector     │ │   Parser   │ │     Config      │
       │                 │ │  (Future)  │ │   (Future)      │
       └────────┬────────┘ └─────┬──────┘ └────────┬────────┘
                │                │                  │
                └────────────────┼──────────────────┘
                                 │
                          ┌──────▼──────┐
                          │ Intermediate │
                          │Representation│
                          │    (IR)      │
                          └──────┬──────┘
                                 │
                ┌────────────────┼────────────────┐
                │                │                │
       ┌────────▼────────┐ ┌────▼─────┐ ┌────────▼────────┐
       │     Model       │ │  Schema  │ │    Convert      │
       │   Generator     │ │Generator │ │   Generator     │
       └────────┬────────┘ └────┬─────┘ └────────┬────────┘
                │                │                │
                │                │                │
                └────────────────┼────────────────┘
                                 │
                          ┌──────▼──────┐
                          │  Go Code    │
                          │ (*.gen.go)  │
                          └─────────────┘
```

## Generation Flow

### High-Level Flow

```
libvirtxml Struct ──┐
                    │
  RNG Schema ───────┼──► Analysis Phase ──► IR ──► Generation Phase ──► Generated Code
                    │
  Config File ──────┘
```

### Detailed Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│ 1. INPUT PHASE                                                        │
└──────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │  libvirtxml.DomainTimer   │
                    │  {                        │
                    │    Name string            │
                    │    Track string           │
                    │    CatchUp *Catchup       │
                    │  }                        │
                    └─────────────┬─────────────┘
                                  │
┌──────────────────────────────────────────────────────────────────────┐
│ 2. REFLECTION PHASE                                                   │
└──────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │  LibvirtXMLReflector      │
                    │  - Analyze struct fields  │
                    │  - Parse XML tags         │
                    │  - Detect nesting         │
                    │  - Convert names          │
                    └─────────────┬─────────────┘
                                  │
┌──────────────────────────────────────────────────────────────────────┐
│ 3. IR GENERATION                                                      │
└──────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │  StructIR                 │
                    │  {                        │
                    │    Name: "DomainTimer"    │
                    │    Fields: [              │
                    │      {GoName: "Name",     │
                    │       TFName: "name",     │
                    │       Required: true}     │
                    │    ]                      │
                    │  }                        │
                    └─────────────┬─────────────┘
                                  │
┌──────────────────────────────────────────────────────────────────────┐
│ 4. CODE GENERATION                                                    │
└──────────────────────────────────────────────────────────────────────┘
                                  │
                ┌─────────────────┼─────────────────┐
                │                 │                 │
       ┌────────▼────────┐ ┌─────▼──────┐ ┌────────▼────────┐
       │ model.go.tmpl   │ │schema.go   │ │convert.go.tmpl  │
       │                 │ │  .tmpl     │ │                 │
       │ type Model      │ │ func       │ │ func FromXML    │
       │ struct {...}    │ │ Schema()   │ │ func ToXML      │
       └────────┬────────┘ └────┬───────┘ └────────┬────────┘
                │                │                  │
                └────────────────┼──────────────────┘
                                 │
                          ┌──────▼──────┐
                          │   gofmt     │
                          │ (formatting)│
                          └──────┬──────┘
                                 │
                          ┌──────▼──────┐
                          │ .gen.go     │
                          │   files     │
                          └─────────────┘
```

## Components

### 1. Parser (`parser/`)

#### LibvirtXML Reflector (`parser/libvirtxml.go`)

Analyzes libvirtxml structs using Go reflection:

```
Input: reflect.Type(libvirtxml.DomainTimer)
                    │
                    ├─► Read struct fields
                    ├─► Parse XML tags
                    ├─► Detect types (string, uint, nested)
                    ├─► Identify optional/required
                    └─► Recursively process nested structs
                    │
Output: StructIR with complete field metadata
```

**Capabilities:**
- Parses `xml:"name,attr"` and `xml:"element"` tags
- Detects `omitempty` → optional field
- Handles pointer types → optional field
- Recursively processes nested structs
- Converts CamelCase → snake_case for Terraform

**Future: RNG Parser**
- Parse RelaxNG schema files
- Extract validation patterns
- Identify required vs optional more accurately
- Extract documentation strings

### 2. Generator (`generator/`)

#### Intermediate Representation (`generator/ir.go`)

The IR is the bridge between input (libvirtxml) and output (generated code):

```go
StructIR {
    Name: "DomainTimer"
    Fields: [
        FieldIR {
            GoName: "Name"        // Go field name
            XMLName: "name"       // XML attribute/element
            TFName: "name"        // Terraform attribute (snake_case)
            GoType: "string"      // Original Go type
            TFType: "types.String" // Terraform type
            IsRequired: true
            IsOptional: false
        },
        FieldIR {
            GoName: "CatchUp"
            IsNested: true
            NestedStruct: StructIR {...}  // Recursive
        }
    ]
}
```

#### Model Generator (`generator/model.go`)

Generates Terraform model structs:

```
Input: []StructIR
           │
Template: model.go.tmpl
           │
Output:
    type DomainTimerModel struct {
        Name types.String `tfsdk:"name"`
        ...
    }
```

**Features:**
- Generates one model per struct
- Includes nested models
- Adds tfsdk tags
- Preserves field documentation

#### Schema Generator (`generator/schema.go`)

Generates Plugin Framework schemas:

```
Input: []StructIR
           │
Template: schema.go.tmpl
           │
Output:
    func DomainTimerSchema() schema.Attribute {
        return schema.SingleNestedAttribute{
            Attributes: map[string]schema.Attribute{
                "name": schema.StringAttribute{
                    Required: true,
                },
            },
        }
    }
```

**Features:**
- Determines Required/Optional from IR
- Handles nested attributes recursively
- Supports all schema types (String, Int64, Bool, etc.)
- Adds descriptions (future: from RNG)

#### Conversion Generator (`generator/convert.go`)

Generates bidirectional conversion functions:

```
Input: []StructIR
           │
Template: convert.go.tmpl
           │
Output:
    func FromXML(xml *libvirtxml.T, plan *Model) (*Model, error)
    func ToXML(model *Model) (*libvirtxml.T, error)
```

**Key Pattern: Preserve User Intent**

```go
// Optional field - only populate if user specified it
if plan != nil && !plan.Field.IsNull() && !plan.Field.IsUnknown() {
    if xml.Field != "" {
        model.Field = types.StringValue(xml.Field)
    }
} else {
    model.Field = types.StringNull()
}
```

This prevents Terraform from detecting drift when libvirt applies defaults that the user didn't explicitly specify.

### 3. Templates (`templates/`)

Go text/template files that define the structure of generated code:

- **`model.go.tmpl`** - Model struct template
- **`schema.go.tmpl`** - Schema definition template
- **`convert.go.tmpl`** - Conversion function template

Templates have access to the full IR and can:
- Iterate over fields
- Conditionally generate code
- Call helper functions
- Recursively handle nesting

### 4. Main (`main.go`)

The orchestrator that ties everything together:

```
1. Create reflector
2. Analyze target struct(s)
3. Collect all structs (including nested)
4. Generate models
5. Generate schemas
6. Generate conversions
7. Write formatted Go files
```

## Usage

### Running the Generator

```bash
# From project root
go run ./internal/codegen

# Output
Generated: internal/generated/domain_timer_model.gen.go
Generated: internal/generated/domain_timer_schema.gen.go
Generated: internal/generated/domain_timer_convert.gen.go
Code generation completed successfully!
```

### Generated Output

```
internal/generated/
├── domain_timer_model.gen.go     # Model structs
├── domain_timer_schema.gen.go    # Schema functions
└── domain_timer_convert.gen.go   # Conversion functions
```

### Verifying Generated Code

```bash
# Build to check compilation
go build ./internal/generated/...

# Run tests
go test ./internal/generated/... -v
```

## Generated Code Structure

### Example: DomainTimer

**Input (libvirtxml):**
```go
type DomainTimer struct {
    Name       string              `xml:"name,attr"`
    Track      string              `xml:"track,attr,omitempty"`
    TickPolicy string              `xml:"tickpolicy,attr,omitempty"`
    CatchUp    *DomainTimerCatchUp `xml:"catchup"`
    Frequency  uint64              `xml:"frequency,attr,omitempty"`
}

type DomainTimerCatchUp struct {
    Threshold uint `xml:"threshold,attr,omitempty"`
    Slew      uint `xml:"slew,attr,omitempty"`
    Limit     uint `xml:"limit,attr,omitempty"`
}
```

**Generated Model:**
```go
type DomainTimerModel struct {
    Name       types.String `tfsdk:"name"`
    Track      types.String `tfsdk:"track"`
    TickPolicy types.String `tfsdk:"tick_policy"`
    CatchUp    types.Object `tfsdk:"catch_up"`
    Frequency  types.Int64  `tfsdk:"frequency"`
}

type DomainTimerCatchUpModel struct {
    Threshold types.Int64 `tfsdk:"threshold"`
    Slew      types.Int64 `tfsdk:"slew"`
    Limit     types.Int64 `tfsdk:"limit"`
}
```

**Generated Schema:**
```go
// For nested components - returns schema.Attribute
func DomainTimerSchema() schema.Attribute {
    return schema.SingleNestedAttribute{
        Optional: true,
        Attributes: map[string]schema.Attribute{
            "name": schema.StringAttribute{
                Required: true,
            },
            "track": schema.StringAttribute{
                Optional: true,
            },
            "catch_up": schema.SingleNestedAttribute{
                Optional: true,
                Attributes: map[string]schema.Attribute{
                    "threshold": schema.Int64Attribute{
                        Optional: true,
                    },
                    // ...
                },
            },
        },
    }
}

// For top-level resources - returns complete schema.Schema with helper
func DomainResourceSchema(overrides map[string]schema.Attribute) schema.Schema {
    attrs := map[string]schema.Attribute{
        "name": schema.StringAttribute{
            Required: true,
        },
        "memory": schema.Int64Attribute{
            Required: true,
        },
        "clock": DomainClockSchema(),
        "devices": DomainDevicesSchema(),
        // ... all libvirtxml fields
    }

    // Merge resource-specific overrides (like computed ID)
    for k, v := range overrides {
        attrs[k] = v
    }

    return schema.Schema{
        Description: "Manages a libvirt domain (virtual machine)",
        Attributes:  attrs,
    }
}
```

**Generated Conversions:**
```go
func DomainTimerFromXML(ctx context.Context, xml *libvirtxml.DomainTimer,
                        plan *DomainTimerModel) (*DomainTimerModel, error) {
    model := &DomainTimerModel{}

    // Required field - always populate
    model.Name = types.StringValue(xml.Name)

    // Optional field - preserve user intent
    if plan != nil && !plan.Track.IsNull() && !plan.Track.IsUnknown() {
        if xml.Track != "" {
            model.Track = types.StringValue(xml.Track)
        }
    } else {
        model.Track = types.StringNull()
    }

    return model, nil
}

func DomainTimerToXML(ctx context.Context, model *DomainTimerModel)
                     (*libvirtxml.DomainTimer, error) {
    xml := &libvirtxml.DomainTimer{}

    xml.Name = model.Name.ValueString()

    if !model.Track.IsNull() {
        xml.Track = model.Track.ValueString()
    }

    return xml, nil
}
```

## Resource vs Component Code Generation

### Understanding the Hierarchy

The generator produces code for two types of structures:

```
┌─────────────────────────────────────────────────────────────┐
│                    Terraform Resources                       │
│                    (Top-Level Entities)                      │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
   ┌────▼────┐          ┌─────▼─────┐        ┌─────▼─────┐
   │ Domain  │          │  Volume   │        │  Network  │
   │Resource │          │ Resource  │        │ Resource  │
   └────┬────┘          └───────────┘        └───────────┘
        │
        │ Contains (nested components)
        │
        ├─► Timer (DomainTimer)
        ├─► Disk (DomainDisk)
        ├─► Interface (DomainInterface)
        └─► Graphics (DomainGraphics)
```

**Top-Level Resources** (`libvirt_domain`, `libvirt_volume`, etc.):
- Map to libvirt objects you can create/destroy
- Have CRUD operations (Create, Read, Update, Delete)
- Correspond to `libvirtxml.Domain`, `libvirtxml.StorageVolume`, etc.
- Located in `internal/resources/`

**Nested Components** (`timer`, `disk`, `interface`, etc.):
- Parts of a top-level resource
- No independent lifecycle
- Correspond to `libvirtxml.DomainTimer`, `libvirtxml.DomainDisk`, etc.
- Their generated code is used by parent resources

### Integrating Generated Code with Resources

Generated code doesn't automatically become a Terraform resource. Here's the integration pattern:

#### Step 1: Generate Code for Top-Level Struct

```go
// In internal/codegen/main.go
domainType := reflect.TypeOf(libvirtxml.Domain{})
domainIR, err := reflector.ReflectStruct(domainType)
// This generates:
// - DomainModel
// - DomainSchema()
// - DomainFromXML()
// - DomainToXML()
```

#### Step 2: Create Resource Implementation

Use the generated helper method that returns a complete schema:

```go
// internal/resources/domain.go
package resources

import (
    "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
)

type DomainResource struct {
    client *libvirtclient.Client
}

func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    // Use generated helper - it returns complete schema.Schema
    resp.Schema = generated.DomainResourceSchema(
        // Add/override resource-specific attributes
        map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "Domain UUID",
                Computed:    true,
            },
        },
    )
}
```

**What the generator provides:**

For every struct, two schema functions are generated:

1. **Nested component schema** - Returns `schema.Attribute` for embedding:
   ```go
   func DomainTimerSchema() schema.Attribute
   ```

2. **Resource schema helper** - Returns `schema.Schema` with merge support:
   ```go
   func DomainTimerResourceSchema(overrides map[string]schema.Attribute) schema.Schema {
       attrs := map[string]schema.Attribute{
           "name": schema.StringAttribute{Required: true},
           "track": schema.StringAttribute{Optional: true},
           // ... all libvirtxml fields
       }

       // Merge in resource-specific overrides
       for k, v := range overrides {
           attrs[k] = v
       }

       return schema.Schema{
           Attributes: attrs,
       }
   }
   ```

**Benefits:**
- ✅ **Implemented** - Available now in Phase 1
- One-line schema setup in resources
- Generated code handles all libvirtxml fields
- Easy to add/override resource-specific attributes
- Clean separation: generated handles XML mapping, resource adds Terraform specifics

#### Step 3: Use Generated Conversions in CRUD

```go
// Create operation
func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan DomainResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

    // Convert Terraform model to libvirtxml using generated code
    domainXML := &libvirtxml.Domain{
        Name: plan.Name.ValueString(),
    }

    // Use generated conversion for complex nested components
    if !plan.Timer.IsNull() {
        var timerModel generated.DomainTimerModel
        plan.Timer.As(ctx, &timerModel, basetypes.ObjectAsOptions{})

        timer, err := generated.DomainTimerToXML(ctx, &timerModel)
        if err != nil {
            resp.Diagnostics.AddError("Conversion Error", err.Error())
            return
        }
        domainXML.Clock.Timers = append(domainXML.Clock.Timers, timer)
    }

    // Create in libvirt
    xml, _ := domainXML.Marshal()
    domain, err := r.client.Libvirt().DomainDefineXML(string(xml))
    // ...
}

// Read operation
func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state DomainResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    // Get from libvirt
    xmlStr, err := r.client.Libvirt().DomainGetXMLDesc(domain, 0)

    var domainXML libvirtxml.Domain
    xml.Unmarshal([]byte(xmlStr), &domainXML)

    // Convert libvirtxml to Terraform model using generated code
    if domainXML.Clock != nil && len(domainXML.Clock.Timers) > 0 {
        // Get plan for preserve user intent pattern
        var plan DomainResourceModel
        req.State.Get(ctx, &plan)

        timerModel, err := generated.DomainTimerFromXML(
            ctx,
            domainXML.Clock.Timers[0],
            &plan.TimerModel, // Preserve user intent
        )
        // Convert to types.Object and assign to state.Timer
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

### Resource Model vs Generated Model

**Resource Model** (manual, in `internal/resources/`):
```go
type DomainResourceModel struct {
    ID     types.String `tfsdk:"id"`      // Computed, not in libvirtxml
    Name   types.String `tfsdk:"name"`    // Top-level field
    Memory types.Int64  `tfsdk:"memory"`  // Top-level field
    Timer  types.Object `tfsdk:"timer"`   // Uses generated.DomainTimerModel
}
```

**Generated Model** (auto-generated, in `internal/generated/`):
```go
type DomainTimerModel struct {
    Name       types.String `tfsdk:"name"`
    Track      types.String `tfsdk:"track"`
    TickPolicy types.String `tfsdk:"tick_policy"`
}
```

### Decision: When to Generate

**Generate for:**
- ✅ Complex nested structures (DomainTimer, DomainDisk, DomainInterface)
- ✅ Structures with many optional fields
- ✅ Structures that appear in multiple resources
- ✅ Eventually, entire top-level resources (Phase 4)

**Don't generate yet:**
- ❌ Simple 2-3 field structs (manual is clearer)
- ❌ Resource CRUD logic (always manual)
- ❌ Computed-only fields (id, path)
- ❌ Special cases requiring custom logic

### Example: Full Domain Resource

```
libvirt_domain Resource
├── internal/resources/domain.go (manual)
│   ├── DomainResource struct
│   ├── Schema() - composes generated schemas
│   ├── Create() - uses generated ToXML()
│   ├── Read() - uses generated FromXML()
│   ├── Update() - uses generated conversions
│   └── Delete() - manual libvirt API calls
│
└── Uses generated code from internal/generated/
    ├── domain_timer_model.gen.go
    ├── domain_timer_schema.gen.go
    ├── domain_timer_convert.gen.go
    ├── domain_disk_model.gen.go
    ├── domain_disk_schema.gen.go
    └── domain_disk_convert.gen.go
```

## Extending the Generator

### Adding a New Nested Component

1. **Modify main.go to analyze the struct:**

```go
// Add to run() function
diskType := reflect.TypeOf(libvirtxml.DomainDisk{})
diskIR, err := reflector.ReflectStruct(diskType)
if err != nil {
    return fmt.Errorf("reflecting DomainDisk: %w", err)
}
```

2. **Run the generator:**

```bash
go run ./internal/codegen
```

3. **Verify output:**

```bash
go build ./internal/generated/...
go test ./internal/generated/... -v
```

### Customizing Templates

Templates are in `templates/` and use Go's `text/template` syntax:

```go
{{- range .Structs }}
type {{ .Name }}Model struct {
    {{- range .Fields }}
    {{ .GoName }} {{ .TFType }} `tfsdk:"{{ .TFName }}"`
    {{- end }}
}
{{- end }}
```

**Available template data:**

- `.Structs` - Array of StructIR
- `.Name` - Struct name
- `.Fields` - Array of FieldIR
- `.GoName`, `.TFName`, `.XMLName` - Field names
- `.GoType`, `.TFType` - Type information
- `.IsRequired`, `.IsOptional`, `.IsNested` - Field properties

### Adding Custom Logic

For special cases that don't fit templates:

1. **Add hook detection in reflector:**

```go
// In parser/libvirtxml.go
if field.Name == "Memory" {
    fieldIR.CustomHook = "flatten_unit"
}
```

2. **Handle in template:**

```go
{{- if .CustomHook }}
    // Custom handling for {{ .GoName }}
    // Hook: {{ .CustomHook }}
{{- else }}
    // Standard handling
{{- end }}
```

3. **Or generate placeholder and fill manually:**

```go
// TODO: Custom conversion for Memory with unit flattening
// See internal/codegen/hooks/memory.go
```

## Troubleshooting

### Generated Code Won't Compile

**Check imports:**
```bash
# Common issue: wrong module path
# Fix in template or generated file
```

**Check type conversions:**
```bash
# uint vs uint64 mismatch
# Update template to handle both
```

**Check nested structs:**
```bash
# Missing nested model
# Ensure collectAllStructs() is working
```

### Fields Not Generated

**Check XML tags:**
```go
// This won't be generated (no XML tag)
Field string

// This will be generated
Field string `xml:"field,attr"`
```

**Check reflection:**
```go
// Unexported fields are skipped
field string  // Won't be generated

// Must be exported
Field string  // Will be generated
```

### Type Mapping Issues

Current type mappings in `goTypeToTFType()`:

| Go Type | Terraform Type |
|---------|---------------|
| string | types.String |
| int, int64, uint, uint64 | types.Int64 |
| bool | types.Bool |
| float32, float64 | types.Float64 |
| *Struct | types.Object (nested) |

To add new types, update `parser/libvirtxml.go`.

### Template Syntax Errors

```bash
# Error in template
template: model:10: unexpected "}" in operand

# Fix: Check template syntax
# Common issues:
# - Missing {{- or -}}
# - Unclosed {{ }}
# - Wrong field name in .Field access
```

**Debug template:**
```go
// Add debug output in generator
fmt.Printf("IR: %+v\n", structs)
```

## Next Steps

See `../GEN.md` for:
- Overall architecture and design
- Phase 2-4 implementation plans
- Hook and exclusion system
- RNG schema integration
- Configuration system

## Development Workflow

1. **Modify generator/templates**
2. **Run generator:** `go run ./internal/codegen`
3. **Verify:** `go build ./internal/generated/...`
4. **Test:** `go test ./internal/generated/... -v`
5. **Commit both generator and generated code**

## Key Principles

1. **Generate early, generate often** - Don't hand-write what can be generated
2. **Templates over code** - Easier to modify and understand
3. **Test generated code** - Ensure it works in practice
4. **Document exceptions** - When manual code is needed, explain why
5. **Preserve user intent** - Always consider Terraform state management

---

For questions or issues, see the main project documentation in `../../GEN.md`.
