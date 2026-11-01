---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-01T01:12:06Z
identity: spec-authoring-20261101T011206-jay-flowers
tier: draft
---

When a validation schema enforces numeric bounds (entry counts, timeout caps, length limits), anchor each bound in the requirement text before the task that enumerates boundary test rows is authored — otherwise the requirement, the schema, and the test enumeration diverge silently. Demonstrated on model-matrix-delegation (iteration 4): MM-FR-001 stated no agents count bound, the landed Zod schema enforced an upper bound only (<= 32, so `agents: {}` parsed), task 2.5's enumeration said "0 and 33 pins rejected, 32 accepted" (a 1-32 bound), and the tracked starter matrix required empty `agents` to load — making the task row unimplementable without breaking the starter matrix. The same change left agent_timeout_ms (1_800_000) and request_timeout_ms (180_000) caps in the schema and test rows but out of MM-FR-005/006, while sibling bounds (profiles 1-32, cache 128, 50 MiB response cap) were properly spec-anchored. Symptom: a boundary-test task whose enumerated row fails against the landed schema, or a closed contract that rejects documents the requirement text accepts. Fix pattern: when a bound is enforced in code, record it in the FR (MP-006) in the same edit, and diff the task's enumerated boundary rows against both the requirement and the schema before executing the task.
