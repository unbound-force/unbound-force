---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:00:08Z
identity: issue-triage-20261102T030008-jay-flowers
tier: draft
---

Triage pattern for issues proposing prose-to-tool extraction (Constitution Principle VI / DE-001, e.g. unbound-force issues #665/#671/#672/#673 under epic #659): verify three seams before marking VALID. (1) Source of truth — "mappings live as code constants sourced from the manifest" and "the tool reads the manifest at runtime" are contradictory designs with different failure modes: constants need a drift-detection test against the canonical source; runtime reads need error-path handling for missing, malformed, or schema-drifted files. Flag the contradiction for the clarify phase rather than blocking validity. (2) Input normalization — a free-text lookup parameter needs a specified contract (case, whitespace, empty string, length bounds) even when a fail-safe fallback bounds blast radius; a test criterion of "all known values + fallback" cannot pin matching semantics without it. (3) Savings attribution — token-savings estimates that count removing the original prose are unrealized unless an acceptance criterion tracks the removal; adding the tool alone adds prompt surface, not savings.
