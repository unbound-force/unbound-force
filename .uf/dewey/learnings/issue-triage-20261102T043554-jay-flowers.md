---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:35:54Z
identity: issue-triage-20261102T043554-jay-flowers
tier: draft
---

Triaging unbound-force/unbound-force#571 (rule_id on review-verdict findings) surfaced three repeatable verification checks. (1) Read the cited schema directly: the finding objects in schemas/review-verdict/v1.0.0.schema.json carry exactly severity/category/description/file/line with `additionalProperties: false`, so the convention-pack traceability gap is structurally enforced (the schema actively rejects a rule_id today, not merely omits it), and the schema defines TWO parallel finding shapes — persona_verdicts[].findings and unresolved_findings — so any new finding field must be added to both; the top-level `convention_pack_used` string shows pack-level traceability already exists while rule-level does not, and the $id (internal/schemas/review-verdict-payload) indicates a Go-side generation source that must change in the same increment. (2) Verify code citations by content, not line number: the issue's scaffold.go:478 reference for alwaysDeployedPacks drifted to line 785 as the file evolved (severity-pack membership intact) — grep the symbol, never trust the line. (3) Cross-check architectural intent: specs/005-the-divisor-architecture/research R3 explicitly designed rule IDs so "a finding can cite CS-006" and to support trend analysis, and the review-pr-command SC-004 scenario expects findings to reference pack rules — both prompt-level expectations with no structural capture, making the proposed rule_id field the completion of documented intent rather than a new design. The deployed rule-prefix space is broader than the issue's CS/AP/SC/TC/DR enumeration (also CP, TA, VB, TD, BA, PR, FA, FT, CI, MP, DE, RV, CR — all two-letter), so the proposed pattern ^[A-Z]{2,3}-[0-9]{3}$ covers the full deployed space with headroom.
