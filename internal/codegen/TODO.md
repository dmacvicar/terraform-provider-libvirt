# Code Generator TODO

## Status: Resources Migrated ✅

**Phase 1**: Complete ✅ - Code generator fully functional
**Phase 2**: Complete ✅ - Nested object conversions and list/slice handling implemented
**Phase 3**: In Progress 🚧 - Migrating resources to use generated code

The code generator is **fully functional** and generates code for all libvirt resources.

### Resources Using Generated Code ✅
- **StoragePool** - 100% generated code, all tests passing
- **StorageVolume** - 100% generated code, all tests passing (2025-01-15)
- **Network** - 100% generated code, all tests passing (2025-01-15)
- **Domain** - Partial (using generated conversion helpers)

## What's Implemented

### Generator Infrastructure ✅
- **LibvirtXML Reflector** - Analyzes Go structs using reflection
- **Intermediate Representation (IR)** - Bridges libvirtxml and generated code
- **Model Generator** - Generates Terraform model structs with tfsdk tags
- **Schema Generator** - Generates Plugin Framework schemas
- **Conversion Generator** - Generates XML ↔ Model conversions with user intent preservation

### Resources Generated ✅
- **Domain** (163 structs) - Full VM configuration
- **Network** - Virtual networks
- **StoragePool** - Storage pools
- **StorageVolume** - Storage volumes

**Total:** 163 unique structs, 594 fields, ~15,600 lines of generated code

### Key Features ✅
- **Deduplication** - Shared structs (like `Name`, `StorageEncryption`) generated once
- **Resource schema helpers** - `DomainSchema()` for one-line resource setup
- **Nested schema calls** - `DomainDiskSchema()` instead of recursive inline expansion
- **Preserve user intent** - Optional fields only populated if user specified them
- **Pointer handling** - Correctly handles `*int`, `*uint`, `*string`, etc.
- **Anonymous struct skipping** - Ignores unnamed structs
- **XMLName field skipping** - Ignores encoding/xml metadata fields
- **Nested object conversions** ✅ - Full `types.Object` ↔ nested struct conversions (258 structs)
- **List/slice handling** ✅ - Full `types.List` support for both simple and nested lists
  - Simple lists (`[]string`, `[]int`) → `schema.ListAttribute` with `types.ListValueFrom()`
  - Nested lists (`[]*DomainDisk`) → `schema.ListNestedAttribute` with element-by-element conversion
  - Proper null handling and error messages

### Performance Optimizations ✅
- **Non-recursive schemas** - Call functions instead of inline expansion
- Memory efficient - Was using 836MB and hanging, now completes in seconds
- Template execution optimized for deep nesting

## Generated Files

```
internal/generated/
├── all_model.gen.go      (1,253 lines) - All Terraform models
├── all_schema.gen.go     (3,092 lines) - All schemas
└── all_convert.gen.go    (11,276 lines) - All conversions
```

## Usage

### Generate Code
```bash
go run ./internal/codegen
```

### Use in Resources
```go
// In a resource file
func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    schemaDef := generated.DomainSchema(nil)
    schemaDef.Attributes["id"] = schema.StringAttribute{Computed: true}
    resp.Schema = schemaDef
}
```

## Known Issues / Limitations

1. **Circular references** - Structs with circular references (like `DomainDiskSource` ↔ `DomainDiskDataStore`) cause stack overflow in `AttributeTypes()` generation. Solution in progress: flag-based cycle detection (see "Automatic pattern detection" below)
2. **Documentation+tests lag** - Generator now handles flattened value/unit patterns and boolean-to-presence, but we still need better documentation/tests to keep contributors aligned with the mapping spec.
3. **Large generated files** - All code in 3 big files (~1.7MB total). Future: generate separate files per struct
4. **No RNG schema integration** - Using reflection only, not parsing RNG for validation patterns
5. **No documentation extraction** - Descriptions are empty (need RNG parsing)

## Next Steps (Phase 2+)

### Immediate (to make generated code usable)
- [x] **Implement nested object conversions** ✅ (2025-11-10)
  - ✅ Generated `AttributeTypes()` helper functions for all structs
  - ✅ Implemented XML → types.Object conversion using `types.ObjectValueFrom()`
  - ✅ Implemented types.Object → XML conversion using `object.As()`
  - ✅ Proper null handling and error messages
  - ✅ Generated code compiles successfully

- [x] **Implement list/slice handling** ✅ (2025-11-10)
  - ✅ Detect slice fields in parser with IsList flag
  - ✅ Store ElementGoType for element type information
  - ✅ Generate `types.List` in models
  - ✅ Handle `schema.ListAttribute` for simple types
  - ✅ Handle `schema.ListNestedAttribute` for nested structs
  - ✅ Implement XML → types.List conversion using `types.ListValueFrom()`
  - ✅ Implement types.List → XML conversion using `list.ElementsAs()`
  - ✅ Generated code compiles and passes linting

- [x] **Write integration tests** ✅ (2025-11-10)
  - ✅ Test generated schemas compile
  - ✅ Test XML roundtrip conversions
  - ✅ Test complex nested structures
  - ✅ Test list conversions work correctly
  - 9/9 tests passing

- [ ] **Migrate remaining resources to generated code** ⬅️ **CONTINUE HERE**
  - **Completed**: StoragePool, StorageVolume, Network (2025-01-15)
  - **Next**: Domain resource (largest, most complex - may need careful planning)
  - **Later**: CloudInit, Combustion, Ignition (provider-specific helpers, lower priority)
  - **Learnings from migrations**:
    - Field read semantics documented in AGENTS.md work well
    - Computed fields must always be populated
    - Optional fields only populate if user specified (preserves user intent)
    - Resource code mutates schema attributes after calling `generated.FooSchema()` (id, autostart, path, etc.)
    - Generated conversion handles most cases, manual tweaks needed for computed fields
    - Snake_case conversion issue: "IPs" → "i_ps" (codegen bug to fix later)

- [ ] **Automatic pattern detection**
  - **Goal**: Automatically detect and handle special schema patterns without manual exclusion lists

  #### Pattern 1: Value with Unit (Memory fields)
  - **Status**: ✅ Detection wired into templates (2025-01-10). Value/unit pairs now flatten to sibling attributes (e.g., `memory`, `memory_unit`, `memory_slots`).
  - **Follow-ups**: keep expanding coverage (vcpu placement, cpuset) and make sure docs/tests always reflect the flattening rules.

  #### Pattern 2: Circular References
  - **Problem**: `DomainDiskSource` ↔ `DomainDiskDataStore` causes stack overflow
  - **Solution**: Simple flag-based cycle detection
  - **Implementation**:
    ```go
    type StructIR struct {
        IsBeingAnalyzed bool  // true during analysis, false when done
    }

    // In ReflectStruct:
    if existing := r.processedTypes[typeName]; existing != nil {
        if existing.IsBeingAnalyzed {
            // Cycle detected! Break it.
            fieldIR.IsCycle = true
            return fieldIR, nil
        }
        return existing, nil
    }

    ir.IsBeingAnalyzed = true
    // ... analyze fields ...
    ir.IsBeingAnalyzed = false
    ```
  - **Status**: Stack-based approach written but NOT tested
  - **TODO**: Replace stack with simple flag approach (simpler, no stack overhead)
  - **TODO**: Update templates to skip fields marked with `IsCycle = true`
  - **Files**: `parser/cycle_detection.go` (remove stack, use flag instead)

- [ ] **Generate separate files per struct** ⬅️ **FUTURE IMPROVEMENT**
  - **Current**: One big `all_model.gen.go` (97KB), `all_schema.gen.go` (409KB), `all_convert.gen.go` (1.2MB)
  - **Desired**: Separate file per struct/schema
    ```
    internal/generated/
    ├── domain/
    │   ├── domain_model.gen.go
    │   ├── domain_schema.gen.go
    │   ├── domain_convert.gen.go
    │   ├── domain_disk_model.gen.go
    │   ├── domain_disk_schema.gen.go
    │   └── ...
    ├── network/
    │   └── ...
    └── storage/
        └── ...
    ```
  - **Benefits**:
    - Easier to navigate
    - Better for IDEs (smaller files)
    - Can generate recursively from Domain downwards
    - Parallel generation possible
  - **TODO**: Refactor `main.go` to generate per-struct files
  - **TODO**: Organize by top-level resource (domain/, network/, storage/)

### Short Term
- [ ] **RNG schema parsing**
  - Parse `/usr/share/libvirt/schemas/*.rng`
  - Extract documentation strings
  - Extract validation patterns
  - Identify required vs optional more accurately

- [ ] **Configuration system**
  - `codegen.yaml` for exclusions
  - Field-level customization
  - Hook system for special cases

- [ ] **Better type mapping**
  - Handle more Go types (float, byte, etc.)
  - Custom type mappings
  - Enums from RNG choices

### Medium Term (Phase 3)
- [ ] **Generate tests automatically**
  - Roundtrip tests for each struct
  - Schema validation tests
  - User intent preservation tests

- [ ] **Documentation generation**
  - Extract from RNG annotations
  - Generate Terraform Registry docs
  - Link to libvirt docs

- [ ] **Resource integration**
  - Create actual resource implementations using generated code
  - Replace manual domain/network/volume code
  - Migration guide

### Long Term (Phase 4)
- [ ] **Full code generation**
  - Generate entire resources (not just schemas/models)
  - Generate CRUD operations
  - Minimal manual code

- [ ] **Advanced features**
  - Validators from RNG patterns
  - Plan modifiers
  - State upgraders

## Design Decisions

### Why call schema functions instead of inlining?
- **Performance** - Inlining caused exponential expansion with Domain's deep nesting
- **Readability** - Generated code is much smaller and clearer
- **Maintainability** - Each schema defined once and reused

### Why standalone schema functions instead of `ResourceSchema`?
- **Naming collision** - Resource structs already end with `Resource`, so `DomainResourceSchema()` would be awkward.
- **Clarity** - `DomainSchema()` returns the generated schema, while `DomainSchemaAttribute()` is used for embedding nested objects. This matches the Plugin Framework naming pattern.

### Why skip anonymous structs?
- **No name** - Can't generate a named type
- **Rare** - Very few in libvirtxml
- **Workaround** - Usually can ignore them without impact

## File Structure

```
internal/codegen/
├── README.md              - Architecture and usage guide
├── TODO.md               - This file
├── main.go               - CLI entry point and orchestration
├── generator/
│   ├── ir.go             - Intermediate representation
│   ├── model.go          - Model generator
│   ├── schema.go         - Schema generator
│   └── convert.go        - Conversion generator
├── parser/
│   └── libvirtxml.go     - Reflection-based struct analyzer
└── templates/
    ├── model.go.tmpl     - Model code template
    ├── schema.go.tmpl    - Schema code template
    └── convert.go.tmpl   - Conversion code template
```

## Testing

Current status: Generated code compiles but has TODOs for nested objects and lists.

To test:
```bash
# Build generated code
go build ./internal/generated/...

# Eventually will have tests
go test ./internal/generated/... -v
```

## Recent Work Sessions

### 2025-01-15: Volume Resource Migration ✅
**Migrated volume resource to use 100% generated code**

**Changes:**
- Documented field read semantics in AGENTS.md:
  - Computed (not Optional): Always read from API
  - Optional: Only read if user specified
  - Key insight: Optional flag means "only populate if user cares"

- Fixed volume resource schema:
  - Mark `capacity` as Optional+Computed (computed from upload size)
  - Mark `target.path` as Computed (libvirt always generates it)
  - Add top-level `path` field for interpolation convenience
  - Override generated schema attributes as needed

- Fixed volume Create to not modify plan model
  - Use separate xmlModel for conversion
  - Avoid "inconsistent result" errors

- Fixed readVolume to populate computed fields correctly
  - Always populate `target.path` and top-level `path`
  - Properly handle nested object extraction

- Fixed test syntax:
  - Updated volume format to nested structure
  - Fixed backing_store format syntax
  - Fixed diskWWN test

**Results:**
- All acceptance tests passing (40 pass, 2 skip)
- Volume resource using 100% generated code
- Clean separation: generated handles XML, resource adds TF specifics

### 2025-01-15: Network Resource Migration ✅
**Migrated network resource to use 100% generated code**

**Changes:**
- Replaced manual XML conversion with generated code
- Updated schema to use full libvirt Network structure (breaking changes):
  * `mode` string → `forward.mode` (full forward object)
  * `ips` → `i_ps` (snake_case conversion issue)
  * `bridge` string → full bridge object

- Applied field read semantics:
  * bridge is Optional - only populate if user specified (preserving user intent)
  * bridge.name is Optional+Computed within bridge object
  * autostart is Computed - always populate

- Updated tests to use new schema syntax
- Handled "inconsistent result" error by following Optional field semantics
- Schema override for bridge.name to mark it Computed

**Results:**
- All acceptance tests passing (42 pass, 2 skip)
- Network resource using 100% generated code
- 351 lines deleted, clean implementation

**Key learning**: Optional fields with computed sub-fields need careful handling to avoid
plan consistency errors. Solution: only populate Optional fields if user specified them.

### 2025-11-10 (Session 2): List/Slice Handling ✅
**Implemented full list/slice conversion support**

**Changes:**
- Modified `internal/codegen/parser/libvirtxml.go`:
  - Added slice detection logic in `analyzeField()`
  - Set `IsList` flag for slice types
  - Extract element type (both simple and nested)
  - Store `ElementGoType` field for conversion use

- Modified `internal/codegen/generator/ir.go`:
  - Added `ElementGoType` field to `FieldIR`
  - Updated `ElementType()` method to handle both simple and nested list element types

- Modified `internal/codegen/templates/schema.go.tmpl`:
  - Added `ListNestedAttribute` handling for lists of structs
  - Added `ListAttribute` handling for lists of simple types
  - Updated `attrType` template to handle list types correctly
  - Proper element type determination for both cases

- Modified `internal/codegen/templates/convert.go.tmpl`:
  - Implemented `fromXMLField` list handling with element-by-element conversion
  - Implemented `toXMLField` list handling with `list.ElementsAs()`
  - Added `types.ListValueFrom()` for XML → Model conversion
  - Proper null handling and error messages

**Results:**
- Generated code compiles cleanly (362 structs)
- All linting passes (0 issues)
- Supports both simple lists (`[]string`) and nested lists (`[]*DomainDisk`)
- Phase 2 complete!

**Next:** Write integration tests for roundtrip conversions

### 2025-11-10 (Session 1): Nested Object Conversions ✅
**Implemented full nested object conversion support**

**Changes:**
- Modified `internal/codegen/templates/schema.go.tmpl`:
  - Added `AttributeTypes()` function generation for all structs
  - Added `attrType` template helper for recursive type mapping
  - Added imports for `attr` and `types` packages

- Modified `internal/codegen/templates/convert.go.tmpl`:
  - Implemented `fromXMLField` nested object handling with `types.ObjectValueFrom()`
  - Implemented `toXMLField` nested object handling with `object.As()`
  - Added `fmt` and `basetypes` imports for error formatting
  - Proper null/unknown value handling
  - Descriptive error messages on conversion failure

**Results:**
- Generated code compiles cleanly
- All 258 nested object TODOs resolved
- 516 total conversion operations (258 each direction)
- Project builds successfully

## Recent Commits (on `codegen` branch)

1. `feat: add code generation system (Phase 1 prototype)` - Initial generator
2. `docs: add comprehensive codegen architecture and documentation plan`
3. `docs: clarify resource vs component code generation`
4. `docs: move code generation overview to README, remove GEN.md`
5. `feat: implement resource schema helper pattern` - `Schema()` helper
6. `feat: extend generator to all libvirt resources` - Network, StoragePool, StorageVolume
7. `fix: resolve schema generation performance and naming issues` - Domain generation working

## Quick Reference

### How to add a new resource type
1. Add to `resources` slice in `main.go`:
   ```go
   {"secret", reflect.TypeOf(libvirtxml.Secret{})},
   ```
2. Run `go run ./internal/codegen`

### How generated code is organized
- All structs from all resources go into single `all_*.gen.go` files
- Deduplication ensures shared structs only generated once
- Only top-level resources get `Schema()` helper; nested structs only expose `SchemaAttribute()`

### Template syntax quick reference
- `{{- range .Structs }}` - Loop over all structs
- `{{ .Name }}` - Struct name
- `{{ .Fields }}` - Struct fields
- `{{ template "fieldSchema" . }}` - Call nested template
- `{{- if .IsNested }}` - Conditional
