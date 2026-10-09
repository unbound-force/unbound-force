---
tag: review-council-multi-model-fanout
author: jay-flowers
category: gotcha
created_at: 2026-10-02T16:48:50Z
identity: review-council-multi-model-fanout-20261002T164850-jay-flowers-2
tier: draft
---

Branch opsx/review-council-multi-model-fanout (2026-10-02): Auto-discovered OpenCode plugins under `.opencode/plugins/` (invoke-agent, review-dispatch) are NOT loaded into a session that was started before those plugins existed. The dispatch tools (plan_review_dispatch, invoke_agent, finalize_review_dispatch, acquire_sibling_evidence, prepare_lesson_learning) were absent from both the delegate and the orchestrator toolset, requiring a session reload to register them. When they are absent, the review council falls back to an embodied multi-persona review which produces sound substantive findings but cannot emit canonical review-dispatch artifacts or a native finalize_review_dispatch verdict. A session restart after scaffolding new plugins is required before the shared dispatch protocol can run for real.
