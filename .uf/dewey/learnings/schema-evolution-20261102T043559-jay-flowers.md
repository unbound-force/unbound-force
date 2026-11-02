---
tag: schema-evolution
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:35:59Z
identity: schema-evolution-20261102T043559-jay-flowers
tier: draft
---

Issue #571 (unbound-force) surfaced the pattern that prompt-instructed metadata is not audit-grade until the artifact schema enforces it structurally: Divisor personas were prompt-instructed to cite convention-pack rule IDs (CS-003 etc.) in finding prose, but the review-verdict findings schema captured only severity/category/description/file/line, so rule traceability required natural-language parsing. Remediation pattern: add an optional, regex-constrained field (backward compatible, no required-list change). Operational checks before implementing such an addition: (1) verify which schema version is canonical in the live pipeline — the registry held both review-verdict v1.0.0 (cited by the issue) and v2.0.0 (canonical for the dispatch tooling), so the change must target the version the pipeline actually emits; (2) finding objects use additionalProperties: false, so validators pinned to the pre-change schema will reject new artifacts carrying the new field — the versioned-directory registry is the upgrade-path mechanism for that skew; (3) update the full producer chain (submission tool validators, agent prompt output contracts, consolidation logic), not just the published JSON Schema, or the field stays empty in practice.
