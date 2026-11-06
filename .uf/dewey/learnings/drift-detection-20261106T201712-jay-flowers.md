---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:17:12Z
identity: drift-detection-20261106T201712-jay-flowers
tier: draft
---

When a repository extends a project-owned scaffold asset that was previously guarded by a byte-drift test (e.g. .uf/reviewer-capabilities.yaml exempted from TestEmbeddedAssets_MatchSource), the compensating tests must cover three directions, not two: (1) the embedded canonical asset validated semantically against the schema or a hardcoded known-entries list, (2) the repository copy validated to preserve canonical content unchanged and in canonical order, and (3) the repository's own extensions proven PRESENT, not merely well-formed when present. Direction (3) is the one that gets missed: a one-directional extension test ("every extra entry must have an agent definition") passes with zero extras, so the extension the change exists to ship — e.g. registering a new repo-local reviewer persona — can silently disappear in a later merge and no test fails. The cheapest lock is the inverse of the runtime consumer's error: the dispatch path already fails on discovered agents lacking manifest entries, so a test asserting every .opencode/agents/divisor-*.md file has an entry in the repository manifest pins enrollment for all current and future repo-local personas. Verified on feat/divisor-determinism (PR #959): the mirror exemption gained canonical-side and repository-side guards, but enrollment of divisor-determinism/divisor-linguist remained unpinned until flagged in council review.
