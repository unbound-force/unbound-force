---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-01T01:12:03Z
identity: spec-authoring-20261101T011203-jay-flowers
tier: draft
---

Closed-surface spec requirements need a bidirectional traceability audit at authoring time: every MUST clause needs a scenario and a task verification row, and every bound the schema or implementation enforces needs an FR home. Found on the model-matrix-delegation spec review (iteration 4): MM-FR-001's `agents` MUST clauses carried no scenarios while the Zod schema enforced undocumented bounds (agent-pin count 1-32, agent_timeout_ms cap 1_800_000, classifier.llm.request_timeout_ms cap 180_000) — the spec surface and the enforced surface diverged in both directions despite MM-FR-008's 'runtime validation MUST enforce the same surface as the schema' parity clause, and even though the spec had stated the analogous profile-count and cache bounds. Detection signal: task test rows enumerating bounds that appear in no requirement text. Fix pattern: when an FR declares a closed surface, diff the schema's constraint list against the FR text in both directions before writing the task list, and point every task bound row at the FR clause it verifies.
