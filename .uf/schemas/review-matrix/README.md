# Review Matrix Schema

The review matrix is the canonical policy input for explicit-first
Divisor review dispatch. Version 3 defines model profiles (with
optional `model` fields), mode defaults, advisor allowlists, ordered
explicit runs, opt-in risk augmentation, and execution limits.

The schema validates structural contracts only. The dispatch planner
resolves profile references, applies variant overrides, selects host
fallback, and enforces cross-field planning rules.

## Files

- `v3.schema.json` defines the current version 3 policy contract.
  Profile `model` fields are optional — omit to use host model
  resolution. All other structural rules are unchanged from v2.
- `v2.schema.json` defines the closed version 2 policy contract
  (model required). Preserved for legacy validation.
- `samples/sample-review-matrix.json` covers ordered runs and a
  run-level variant override.
- `samples/valid-host-fallback-absence.json` proves that advisor and
  run entries are optional, preserving planner-level host fallback.
- `samples/invalid-*.json` exercise rejection boundaries.

## Override File

Per-developer model and variant customization is supported via
`.uf/review-matrix.override.yaml`. This file is gitignored and
allows individual developers to override `model` and `variant`
fields for specific profile tiers without modifying the shared
base matrix.

**Constraints (whitelist):**
- Only `profiles.<tier>.[model, variant]` paths are permitted.
- Any other key (e.g., `defaults`, `always`, `runs`) causes a
  validation failure, preventing silent governance weakening.
- The override is merged via nested-path merge — absent keys
  fall through to the base matrix.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 3 | 2026-10-04 | Profile `model` field made optional; host model fallback when omitted. `.uf/review-matrix.override.yaml` support for per-developer model/variant customization. |
| 2 | 2026-10-01 | Initial explicit-first policy contract |
