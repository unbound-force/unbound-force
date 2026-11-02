---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:40Z
identity: issue-triage-20261102T014940-jay-flowers
tier: draft
---

Docs-tracking issues that carry supply-chain facts (runtime engine floors, dependency pin adoptions, fail-closed retrieval semantics) verify in two directions at triage time: (1) the gap direction — each enumerated doc surface lacks the claimed content; (2) the trust direction — every security-relevant fact the issue proposes to document reflects an already-reviewed decision from the originating change, not a new trust decision introduced through the docs issue. Verified triaging unbound-force issue #754 (fix-triage-issue-dispatch-docs docs, 2026-11-01): the Node >= 22.22.2 engine floor and the @opencode-ai/plugin 1.4.10→1.19.0 pin adoption both trace to the merged change's reviewed lockstep set, and the issue self-reports which surfaces were already updated in-change (README + AGENTS.md Technology Stack at commit 0868c22d), so the remaining checklist is the true gap. A docs issue that proposed pin changes or floor changes directly would be a supply-chain change masquerading as documentation and must be routed to a spec workflow, not triaged as docs.
