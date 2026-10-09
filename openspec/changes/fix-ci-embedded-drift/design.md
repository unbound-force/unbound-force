## Context

Three pre-existing test failures on `main` block all PR CI
checks. Two are scaffold embedded-asset drift failures
(`TestEmbeddedAssets_MatchSource`, `TestSCFR001_CanonicalAssetSurfaceMatchesMirrors`),
and the third is a schema validation test
(`TestReviewMatrixSchema_CanonicalPolicyValidates`) whose
expectations no longer match the v3 schema contract.

## Goals / Non-Goals

### Goals

- Fix `TestEmbeddedAssets_MatchSource` by syncing drifted
  scaffold assets with their canonical sources
- Fix `TestSCFR001_CanonicalAssetSurfaceMatchesMirrors`
  (same root cause)
- Fix `TestReviewMatrixSchema_CanonicalPolicyValidates` to
  accept optionally-empty model fields on profiles

### Non-Goals

- Changing the review-matrix v3 schema itself
- Updating the review-matrix.yaml policy content (only
  syncing it)
- Refactoring the scaffold drift detection mechanism

## Decisions

### D1: Byte-copy canonical files to scaffold assets

The scaffold drift tests enforce byte-identical copies.
The fix is to copy the canonical files:

```bash
cp .opencode/commands/uf.review-pr.md internal/scaffold/assets/opencode/commands/uf.review-pr.md
cp .uf/review-matrix.yaml internal/scaffold/assets/uf/review-matrix.yaml
```

This is the established pattern used in prior changes (see
Dewey learnings and OpenSpec changes like
`fix-ask-user-question-tool-name` and
`council-review-sandbox`).

### D2: Relax test expectations for profile model fields

The v3 review-matrix schema made `model` optional on
profiles, allowing `model: {}` (empty, host-resolved). The
CI test at `ci_test.go:350` asserts specific model strings.
The fix updates the assertion to accept `nil` or empty map
values, preserving the structural validation while matching
the actual schema contract.

## Risks / Trade-offs

- **R1: Canonical drift recurrence**: Any future edit to
  `.opencode/commands/uf.review-pr.md` or
  `.uf/review-matrix.yaml` that omits the scaffold copy
  will reproduce these failures. Mitigated by the existing
  CI drift test — it catches mismatch immediately.
