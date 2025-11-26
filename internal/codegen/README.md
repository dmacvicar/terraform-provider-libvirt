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
