---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:56:28Z
identity: review-dispatch-20261104T155628-jay-flowers
tier: draft
---

Make reader-tool coverage enumeration mechanical: list every literal in the tool's input-schema enum (or every error code) and grep the test suite for each literal; the literals with zero happy-path matches are the coverage gaps, and the .opencode suite's 94 percent statements floor never surfaces them. On PR 949's RD-FR-057 suite (.opencode/test/read-dispatch-artifact.test.ts), this enumeration finds section "advisories" never requested and section "findings" requested only past-the-end (offset 5 of a 1-entry section), even though RD-FR-056 requires the triage consumer to read both sections through the tool. Third recurrence of scenario-derived test lists missing requirement-body-enumerated conditions (after learning/review-dispatch-20261104T143125-jay-flowers and the RD-FR-057 oversize-bound gap): judgment-based derivation keeps missing them, so the enumeration belongs in a mechanical pre-merge checklist step (grep each enum literal), extending learning/review-dispatch-20261031T025137-jay-flowers from error codes to all input-schema enum literals.
