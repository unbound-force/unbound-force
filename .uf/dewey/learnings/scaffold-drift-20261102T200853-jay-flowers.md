---
tag: scaffold-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-02T20:08:53Z
identity: scaffold-drift-20261102T200853-jay-flowers
tier: draft
---

New scaffold template pairs (canonical repo-root file + internal/scaffold/assets mirror) must be added to task42CanonicalAssets in internal/scaffold/scaffold_test.go, enforced by TestSCFR001_CanonicalAssetSurfaceMatchesMirrors via bytes.Equal — the uf/model-matrix.yaml precedent. expectedAssetPaths only pins deployment existence, and a rerun-preservation test's substring check does not enforce a byte-identity MUST. The preflight-project-checks change created the uf/preflight.yaml pair without the drift entry, leaving its spec MUST unenforced (found 2026-11). Caveat: when the meta repo customizes its own deployed copy (e.g., declares real checks in .uf/preflight.yaml), byte-parity goes stale by design and the spec MUST must be re-scoped at that point.
