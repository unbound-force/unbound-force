## Why

CI `Build and Test` has been failing on `main` (and all feature
branches) due to three pre-existing test failures. These are not
caused by any feature branch — they exist on `main` and block all
PRs:

1. **Scaffold embedded asset drift**: The canonical command file
   `.opencode/commands/uf.review-pr.md` and the review matrix
   `.uf/review-matrix.yaml` were updated, but their scaffold
   copies at `internal/scaffold/assets/` were not synced. Two
   tests catch this: `TestEmbeddedAssets_MatchSource` and
   `TestSCFR001_CanonicalAssetSurfaceMatchesMirrors`.

2. **Review matrix v3 schema test expectations**: The
   `TestReviewMatrixSchema_CanonicalPolicyValidates` test expects
   specific model strings per profile (`opencode-go/qwen3.8-flash`,
   etc.), but the canonical review-matrix.yaml now uses
   `model: {}` (empty model map). The v3 schema made `model`
   optional, and the canonical file reflects that.

## What Changes

- Sync the two drifted scaffold assets to their canonical sources
- Update the CI schema test to handle optionally-empty model
  fields on review matrix profiles

## Capabilities

### Modified Capabilities

- **CI pipeline**: `Build and Test` job passes cleanly on all
  branches

## Impact

- `internal/scaffold/assets/opencode/commands/uf.review-pr.md` —
  copied from canonical
- `internal/scaffold/assets/uf/review-matrix.yaml` — copied from
  canonical
- `internal/schemas/ci_test.go` — `TestReviewMatrixSchema_CanonicalPolicyValidates`
  updated to accept nil/empty model fields

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change fixes CI infrastructure — no artifact format or
communication contract changes. Heroes continue to collaborate
through existing artifact envelopes.

### II. Composability First

**Assessment**: PASS

CI fixes improve reliability for all heroes. No new dependencies
or coupling introduced.

### III. Observable Quality

**Assessment**: PASS

Fixes restore the CI quality signal. A failing `Build and Test`
is noise that masks real regressions. This change makes the CI
signal meaningful again.

### IV. Testability

**Assessment**: PASS

The fix updates test expectations to match the actual schema
contract (model is optional in v3) and syncs scaffold assets with
their canonical sources — both preserving existing test coverage.

### V. Security by Default

**Assessment**: N/A

No security-sensitive changes. No new dependencies, inputs, or
permission changes.
