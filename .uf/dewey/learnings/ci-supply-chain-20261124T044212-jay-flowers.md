---
tag: ci-supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-24T04:42:12Z
identity: ci-supply-chain-20261124T044212-jay-flowers
tier: draft
as_of: "unbound-force@90e1e32de33ae656223b0bc8ff4187eb4ada30e0"
---

When Dependabot bumps a SHA-pinned reusable workflow, verify no sibling references to the same upstream remain at the old SHA (version skew). In unbound-force/unbound-force, .github/workflows/fullsend.yaml holds the only fullsend-ai/fullsend/reusable-dispatch.yml reference, so a single-file bump leaves no skew. Sibling repos differ: unbound-force/replicator keeps .github/workflows/prioritize.yml and fullsend.yaml pinned in lockstep (see replicator PR #161), so a one-file Dependabot PR there would signal partial upgrade. A correct bump keeps the full 40-char immutable SHA and updates the trailing version comment to the new tag (CI-001 + CI-003).
