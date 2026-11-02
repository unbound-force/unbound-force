---
tag: review-verdict
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:35:50Z
identity: review-verdict-20261102T043550-jay-flowers
tier: draft
---

Issue #571 triage (Architect persona): the review-verdict schema registry carries TWO versions — schemas/review-verdict/v1.0.0.schema.json and v2.0.0.schema.json — and v2.0.0 is the canonical active one (finalize_review_dispatch emits version-2 decision data; v2 adds INCONCLUSIVE/UNAVAILABLE council verdicts, maxItems bounds, and a shared $defs/finding). Proposals that change the finding shape must target v2.0.0's $defs/finding, where one edit covers both persona_verdicts[].findings and unresolved_findings (v1.0.0 instead duplicates the finding object in two places). Both schema versions set additionalProperties:false on findings, and the same shape is enforced in plugin-side finding contracts (submit_review_findings Zod validation, dispatch finalize payload), so an optional rule_id field must be added at every finding definition in the pipeline or child reviewers cannot emit it end-to-end. Verified supporting fact: every convention-pack rule prefix in the ecosystem is exactly 2 uppercase letters + 3 digits (CS, AP, SC, TC, DR, CP, CI, VB, TD, BA, PR, FA, FT, MP, DE, RV, TA, CR), so the pattern ^[A-Z]{2,3}-[0-9]{3}$ proposed in issue #571 accepts all existing rule IDs.
