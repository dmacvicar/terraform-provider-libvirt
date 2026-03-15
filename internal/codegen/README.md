# Code Generator for terraform-provider-libvirt

The generator keeps the provider in sync with libvirtxml by emitting the Terraform models, Plugin Framework schemas, and XML conversions automatically.

## What it produces

- Terraform models with `tfsdk` tags
- Schema definitions with correct optional/required/computed flags and descriptions from the docs registry
- Conversion helpers that preserve user intent on optional fields

## Mapping rules

The generator follows the global XML <-> HCL mapping rules in `docs/schema-mapping.md` (nested attributes only, flattening rules, union handling, presence booleans). New features must conform there; do not add Terraform blocks.

## Architecture

- Inputs: libvirtxml reflection, docs registry (`internal/codegen/docs/*.yaml` from docindex/docgen), small config hooks
- IR builder: normalizes struct metadata, tracks optionality, and carries doc strings
- Field policy layer: applies Terraform-specific semantics after reflection (for example top-level identities, immutability, and exact path overrides)
- Generators: templates render models/schemas/converters into `internal/generated/*.gen.go`
- Orchestration: `main.go` wires the pieces and runs gofmt

```
            libvirtxml structs               docindex/docgen (docs)
                    |                                |
                    v                                v
             +-------------+                +---------------+
             |  IR builder |<---------------| docs YAML     |
             |             |                | internal/     |
             +------+------+                | codegen/docs/ |
                    |                       +-------+-------+
                    |
                    v
          +---------------------+
          |  code generators    |
          | (models/schemas/    |
          |  converters)        |
          +----------+----------+
                     |
                     v
          +---------------------+
          | internal/generated/ |
          | *.gen.go            | ----------------> tfplugindocs
          +----------+----------+                      |
                     |                                 |
                     v                                 v
             Provider binaries                   docs/ markdown
```

## Usage

- Generate everything: `go run ./internal/codegen` (or `make generate`)
- Generated Go files land in `internal/generated/` (xxx_convert.gen.go, xxx_schema.gen.go, xxx_model.gen.go)

- Resources embed the generated models/schemas and call the conversions; add/override resource-specific fields (IDs, create helpers) manually

## Field Policy Design

The generator has two separate responsibilities:

1. Structural analysis: reflect libvirtxml structs into IR using facts from Go types and, over time, RNG metadata. This layer should answer questions like "is this a pointer?", "is this an XML attribute?", and "is this nested or repeated?".
2. Terraform policy: decide how those reflected fields behave in Terraform schemas and conversions. This layer owns decisions like `Computed`, `Required`, `RequiresReplace`, and "preserve user intent" behavior.

Keep those layers separate. The parser should not accumulate ad hoc Terraform exceptions based on struct names. If a field needs special Terraform behavior, prefer a policy rule after reflection rather than embedding one-off conditionals into the reflector.

### Policy rules

Policy should be applied in an ordered pass over `StructIR` / `FieldIR`:

- Generic scope-aware rules first
- Exact path overrides second
- Resource-specific fallbacks last

Examples:

- Top-level resource identity fields such as `id`, `uuid`, and `key` are provider-managed and should usually be `Computed`
- Top-level `name` and some top-level `type` fields are immutable inputs and should usually be `Required` plus `RequiresReplace`
- Nested `id` fields are not automatically provider-managed; many are part of libvirt configuration and should keep their reflected semantics unless an explicit override says otherwise
- Reported-only fields such as storage pool `capacity` / `allocation` / `available` should be handled by explicit policy rules, not inferred from field names alone

### Override strategy

When the default rules are not enough, use explicit overrides keyed by full Terraform or XML path rather than helper functions like `isUserManagedFooStruct`.

Good override targets:

- `storage_pool.capacity`
- `storage_pool.allocation`
- `storage_volume.physical`

Avoid:

- Struct-name allowlists for one field
- Field-name heuristics that ignore nesting scope
- Mixing reflection logic with provider semantics in the same function

This keeps the generator predictable, makes exceptions easy to audit, and prevents the parser from turning into a collection of special cases.

## Documentation tools

- `cmd/docindex`: scrape `/usr/share/doc/libvirt/html` into an index used for prompting. Run `go run ./internal/codegen/cmd/docindex --input /usr/share/doc/libvirt/html --output internal/codegen/docs/.index.json`.

- `cmd/docgen`: uses the index + IR paths to propose YAML docs in `internal/codegen/docs/*.yaml`. Run with `--generate` to call the OpenAI API (`OPENAI_API_KEY` required). Supports `--model`, `--start-batch`, and `--batch-size` for resumable runs. Output is appended directly to the YAML files alongside any existing manual entries.

## Prompt improvement plan (docgen)

1. Feed RNG validation context: include the RNG xpath for each field (available in the IR) so the prompt can extract allowed values/patterns from `/usr/share/libvirt/schemas`. Where patterns exist (e.g., WWN), request that the AI restate valid forms.
2. Ask for concrete examples: instruct the prompt to emit 1–2 short examples of valid values (when meaningful) and to omit examples when the field is self-evident (e.g., booleans).
3. Clarify optionality: include whether the RNG marks the field as optional/required and ask the model to call that out explicitly in the description.
4. Cross-link sources: keep libvirt reference URLs and the original XML path in the prompt so the response can cite the right section and stay aligned with schema naming.
5. Guardrails: require the model to avoid inventing defaults/side effects; everything must be phrased as "user-provided" unless RNG states otherwise. Reject or re-prompt when validation details are missing.

These improvements should make the generated YAML confident enough to surface valid values and constraints directly from upstream sources.
