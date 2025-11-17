# TODO

## Completed

- [x] Update acceptance tests for new disk schema (2025-10-18)

## Code Generation Status (2025-11-16)

**✅ DONE:**
- Value+unit flattening: `capacity` + `capacity_unit` (not nested objects)
- Boolean-to-presence, string-to-bool patterns
- 324 structs generated (Domain, Network, StoragePool, StorageVolume) - all compile
- Cycle detection, anonymous field handling
- Pool resource using `generated.StoragePoolFullSchema()` with overrides ✅
- Pool acceptance tests pass ✅

**Current focus areas (2025-01-18):**

1. **Domain resource migration**:
   - Schema/model now fully generated; finish migrating remaining acceptance tests and docs to the nested syntax (interfaces, filesystems, RNG, nvram, etc.).
   - RNG acceptance test now reflects the generated schema directly (`model = "virtio"`). Keep this example aligned with docs.
   - Confirm wait_for_ip overrides keep working after generator changes.

2. **Generated schema QA + docs**:
   - Keep `internal/codegen/README.md`, AGENTS.md, and examples aligned with the XML↔HCL mapping spec (flattened values, boolean-to-presence, mutually exclusive lists).
   - Add/maintain unit tests in `internal/generated/conversion_test.go` for tricky generator patterns (flattened values with units, boolean presence, nested list truncation, RNG backends).
   - Track schema mismatches with RelaxNG for future enhancements.

3. **Volume resource migration (queued)**:
   - Embed `generated.StorageVolumeModel`.
   - Add provider-specific fields (`pool`, `create` block).
   - Replace manual conversions with generated ones; keep special create/upload logic separate.

4. **RNG parser (future)**:
   - Parse RelaxNG to auto-detect `<choice>`/mutual exclusivity, better optionality, and validations (WWN patterns, etc.).
   - Will enable automatically generated validators/documentation.

### Documentation generation spec

- Maintain a central doc registry (`internal/codegen/docs/*.yaml`) keyed by libvirt XML paths (e.g., `domain.devices.disks.source.volume.pool`).
- Extend codegen IR to attach `Description`, `MarkdownDescription`, and `DeprecationMessage` fields sourced from the registry; fallback to generic text when no entry exists.
- Update schema templates to emit those fields for every attribute/nested object so descriptions stay in sync with generated code.
- Provide tooling (`go run ./internal/codegen/cmd/doclint`) to report missing docs and keep coverage visible.

## Pool Resource Migration Complete (2025-01-12)

✅ **Successfully migrated pool resource to 100% generated code**

**Results:**
- Pool resource: 628 → 428 lines (32% reduction, -200 lines)
- Eliminated 372 lines of manual conversion code, added only 31
- All acceptance tests pass
- Now uses generated model, schema, AND conversions

**What was built:**
1. Universal field patterns hardcoded in generator (uuid→Computed, name→RequiresReplace, etc.)
2. Resource-specific patterns (StoragePool capacity→Computed)
3. Nested plan extraction and preservation through entire conversion chain
4. User intent preservation for all optional fields (including nested)

**Key insights:**
- Generator extracts nested plans from parent and passes to nested conversions
- Unit fields (_unit suffix) only populated if user specified them
- Read() must pass current state as plan to preserve user intent
- Pattern: Embed generated model, add resource-specific fields (ID, etc.)

## Volume Resource Migration (Next)

**Complexity discovered:**
Volume has more provider-specific logic than pool:
- `pool` field (not in libvirtxml - specifies which pool to create in)
- `create` block (upload content from URL on creation)
- Current schema is flat (format/path/permissions at top) but should be nested in target

**Approach:**
1. Follow pool pattern: embed generated.StorageVolumeModel
2. Add provider-specific fields: ID, Pool, Create
3. Update schema to use generated with overrides for ID/Pool/Create
4. Replace manual conversions with generated.StorageVolumeToXML/FromXML
5. Handle special Create logic (URL upload) separately after volume creation

**Estimated effort:** Similar to pool (~2-3 hours) due to Create block complexity.

## Notes

- Domain-level disk `<backingStore>` input is intentionally not supported; configure copy-on-write via the `libvirt_volume.backing_store` block until libvirt exposes the `backingStoreInput` feature to guests.
