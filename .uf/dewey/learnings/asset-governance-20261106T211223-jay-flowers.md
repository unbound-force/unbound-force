---
tag: asset-governance
author: jay-flowers
category: pattern
created_at: 2026-11-06T21:12:23Z
identity: asset-governance-20261106T211223-jay-flowers
tier: draft
---

Project-owned policy mirrors that diverge from embedded scaffold assets need compensating locks in BOTH directions, not just an exemption from byte-identity. The exemption alone (projectOwnedPolicyMirrors skipping .uf/preflight.yaml and .uf/reviewer-capabilities.yaml in TestEmbeddedAssets_MatchSource and TestSCFR001) silently rots. The working pattern, verified in reviewer_capabilities_test.go: (1) the extends lock — the repository copy must contain every canonical entry unchanged and in canonical order (reflect.DeepEqual on the canonical subsequence), and every extra entry must resolve to a real agent definition on disk; (2) the enrollment lock — every divisor-* definition in .opencode/agents must have a manifest entry, so deleting an enrollment fails the suite instead of silently dropping a persona from dispatch. Each byte-identity exemption should name its compensating test the way AGENTS.md's policy-mirrors line does (projectOwnedPolicyMirrors in scaffold_test.go, enrollment lock in reviewer_capabilities_test.go, declared-check shape in instruction-lint-preflight.test.ts).
