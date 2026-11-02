---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:35:47Z
identity: issue-triage-20261102T043547-jay-flowers
tier: draft
---

When triaging issues that propose changes under schemas/ in unbound-force repos, check the schema directory for newer versioned siblings and the README Version History before assessing factual accuracy: issues filed before a major-version migration cite the old file (issue #571 targeted schemas/review-verdict/v1.0.0.schema.json, but v2.0.0 has been canonical since 2026-10-01 — README states "New producers MUST emit schema 2.0.0"). The described field gap usually persists across versions because major bumps change decision enums, not the finding shape — v1.0.0 and v2.0.0 findings are identical (severity/category/description/file/line, additionalProperties: false) — so the proposal stays valid but must be retargeted to a new additive v2.x schema version rather than editing the historical file. Two companion checks: grep the whole repo for the proposed field name (rule_id|ruleId returned zero matches across schema, Go, and TypeScript — confirming no implementation and no duplicate), and flag that a schema-only addition cannot deliver the stated benefit — the dispatch pipeline finding producers (submit_review_findings Zod contract, finalize_review_dispatch payload schema, agent output prompt contracts) must carry the field too, or it is never populated.
