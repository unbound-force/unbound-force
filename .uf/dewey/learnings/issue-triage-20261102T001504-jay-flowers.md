---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:04Z
identity: issue-triage-20261102T001504-jay-flowers
tier: draft
---

Phase-tracker issues under an umbrella migration that carry post-window cleanup items — work gated on a support window closing (e.g., "remove v1 server() implementations once the 90-day window closes and no v1 consumers remain") — should extract those items into a follow-up issue filed at cutover time. A phase issue that mixes immediate cutover work with items that cannot execute for 90+ days cannot close cleanly: it either stays open past phase completion or closes with tracked-but-unexecuted work, losing the reminder. Observed while triaging unbound-force issue #807 (Phase 5 cutover, opencode-v2-migration umbrella #797).
