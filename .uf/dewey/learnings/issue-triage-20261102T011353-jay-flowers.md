---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:13:53Z
identity: issue-triage-20261102T011353-jay-flowers
tier: draft
---

Documentation-tracking issues filed per the unbound-force documentation gate triage cleanly when they carry full artifact traceability: each user-facing behavior change enumerated as a checklist item mapped to the change's spec FR numbers, plus citations of the spec directory, originating bug report, antecedent docs issue, and an explicit non-overlap statement for adjacent issues. Issue #778 (docs for fix-dispatch-verdict-derivation, fixes #764) demonstrates the pattern: its six checklist items map one-to-one onto the delta spec's FR-009–FR-012 scenarios, so triage verification reduces to artifact-existence checks (spec directory present, command-contract citations grounded, duplicate search clean) and the VALID/enhancement/objective verdict follows mechanically rather than from judgment.
