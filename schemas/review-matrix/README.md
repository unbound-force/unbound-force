# Review Matrix Schema

The review matrix is the canonical policy input for explicit-first
Divisor review dispatch. Version 2 defines model profiles, mode
defaults, advisor allowlists, ordered explicit runs, opt-in risk
augmentation, and execution limits.

The schema validates structural contracts only. The dispatch planner
resolves profile references, applies variant overrides, selects host
fallback, and enforces cross-field planning rules.

## Files

- `v2.schema.json` defines the closed version 2 policy contract.
- `samples/sample-review-matrix.json` covers ordered runs and a
  run-level variant override.
- `samples/valid-host-fallback-absence.json` proves that advisor and
  run entries are optional, preserving planner-level host fallback.
- `samples/invalid-*.json` exercise rejection boundaries.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 2 | 2026-10-01 | Initial explicit-first policy contract |
