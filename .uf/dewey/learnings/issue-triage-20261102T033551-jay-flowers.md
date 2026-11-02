---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:51Z
identity: issue-triage-20261102T033551-jay-flowers
tier: draft
---

Stale-premise verification is mandatory when triaging orchestration-agent issues in unbound-force: issue #658 (filed 2026-10-05) accurately described /uf.unleash delegating review-council work to cobalt-crush-dev (cobalt-crush-dev.md is exactly 270 persona lines) and proposed a lean review-orchestrator agent, but nine days later the unleash-direct-review-dispatch change (issue #716 Increment 1, landed 2026-10-14) removed the review-step child driver entirely — /uf.unleash now calls capture_review_diff → plan_review_dispatch → dispatch_review_runs → consolidate_dispatch directly, and its command text explicitly reserves cobalt-crush-dev for "content work, not review driving." Two checks before validating such an issue: (1) re-verify the delegation premise against current command text, since orchestration commands evolve weekly and line/step references go stale; (2) test the proposal against determinism pack DE-005 (Constitution Principle VI) — when the proposed new agent would drive a judgment-free deterministic tool sequence, the governance-preferred fix is intent-level tools with no driver at all, not a leaner driver. Corroborating empirical evidence: Dewey learning uf-unleash-20261006T004544-jay-flowers records that the intermediary approach consumed excessive context through the Task tool boundary and failed at dispatch protocol steps.
