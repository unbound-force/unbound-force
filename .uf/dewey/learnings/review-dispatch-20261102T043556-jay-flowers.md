---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:35:56Z
identity: review-dispatch-20261102T043556-jay-flowers
tier: draft
---

Schema-layer drift from instruction-layer intent: convention-pack rule IDs were designed (specs/005-the-divisor-architecture research R3) for machine-traceable findings citation and trend analysis, and review-pr-command SC-004 requires findings to cite pack rules — but the review-verdict schema's finding object (both v1.0.0 and v2.0.0: severity/category/description/file/line, additionalProperties:false) never gained a rule reference field, so citation compliance stayed prompt-dependent and unauditable (verified via repo-wide grep: rule_id appears nowhere). Generalizable rule: when a capability is added to reviewer instruction (packs, prompts, commands), mirror it in the artifact schema in the same change, or the capability silently degrades to prose. Design note for the fix: the field must be optional — a required rule_id on LLM-generated findings induces fabricated citations for findings with no rule basis, corrupting the traceability data the field exists to provide; optionality keeps the grounded subset honest and makes absence itself auditable signal. Note also that the top-level convention_pack_used string cannot carry this provenance: directory auto-discovery co-loads every pack into each review prompt, so only the per-finding rule ID (whose prefix maps into the global prefix namespace) identifies which co-loaded pack motivated a finding.
