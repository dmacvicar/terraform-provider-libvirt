# TODO

- [x] Consolidate README and codegen README so they are not redundant; keep only the essentials in each
- [x] Keep the XML <-> HCL mapping rules centralized in `docs/schema-mapping.md` and make sure both READMEs point to it (no block syntax)
- [x] Document the codegen documentation tools (docindex/docgen) and link to them from the main README
- [x] Remove or update references to obsolete tools (e.g., doclint) and dead pipelines
- [ ] Review/adjust the generated docs YAML once the new guidance is in place to confirm correctness
- [x] Improve the docgen prompt to surface valid values/examples using RNG context and the IR paths (plan drafted in `internal/codegen/README.md`)
- [x] Keep the codegen README focused on tricky cases and exceptions rather than repeating the main README
- [ ] Upgrade docgen/docindex prompts and inputs:
  - Enrich FieldContext with optional/required/computed flags, presence/string-to-bool/flattening info, valid values/patterns, and union notes (one-of branches).
  - Include docindex URLs as `reference` instead of forcing empty refs; only leave blank when nothing is known.
  - Prompt for constraints and short examples when meaningful; avoid examples for trivial booleans.
  - Add guardrails against invented defaults/side effects; prefer “user-specified” when unsure.
  - Build batches skipping already-documented paths to avoid overwriting manual entries.
  - (Later) feed RNG-derived optionality/enums/patterns into FieldContext once the RNG parser is wired.
