---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T08:00:03Z
identity: determinism-20261116T080003-jay-flowers
tier: draft
---

A complete DE-001 extraction increment in unbound-force verifies with three deterministic checks, demonstrated by extract-feedback-triage-artifact-tool (issue #1022): (1) the CHANGELOG's DE-004 delta must match the `instruction-lint --json --base <ref> --changed` delta block number-for-number — that change recorded "1451 → 1424 lines (16,738 → 16,727 tokens; load set 23,306 → 23,295) per copy", exactly the linter's delta; (2) `node .opencode/lib/tool-inventory.ts --seams --base <ref> --json` must report the new tool with `missing: []` — that single output replaces reading test files for the DE-007 three-layer check (registration via genericRegistration, output-shape via golden-output-shape.test.ts, round-trip via the lib's test file); (3) the rewritten command section must retain exactly one rendered-output example at the decision point (DE-002) and instruct verbatim printing of the tool-rendered summary line (DE-003), never template reproduction. The mechanical chain that used to follow the write — the terminal clear_progress call — collapses inside the tool via a deps hook (DE-005), and the judgment products (triage items carrying comment identity, PR identity, artifact context, replies result) arrive as schema-validated parameters with a superRefine decision cross-check that fails closed on a mis-join (DE-006).
