---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:58:33Z
identity: issue-triage-20261102T025833-jay-flowers
tier: draft
---

When triaging documentation-gap issues that cite specific behavior details (defaults, severities, ordering rules), verify the claims against the shipped implementation and the OpenSpec spec delta — not the change's tasks.md. Task descriptions can drift from the final spec. Concrete case (issue #676, unbound-force/unbound-force): tier-gated-agent-count tasks.md task 2.3 describes the heavy-cap advisory as MEDIUM severity, but spec FR-006, the CHANGELOG entry, and the implementation (.opencode/plugins/review-dispatch/index.ts, severity: "HIGH") all say HIGH. The issue correctly documented HIGH; only the completed tasks.md is stale. The implementation and spec-review-passed spec delta are the authoritative sources for shipped behavior.
