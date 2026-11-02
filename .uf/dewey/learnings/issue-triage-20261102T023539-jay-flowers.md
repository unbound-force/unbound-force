---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:39Z
identity: issue-triage-20261102T023539-jay-flowers
tier: draft
---

Large tracking issues can carry body text that conflicts with repo governance — e.g., unbound-force #714's Decision 4 routed a cross-repo, >3-story migration around the spec workflow, contradicting AGENTS.md's Speckit-strategic tier rule. Before flagging a governance violation or scoring NEEDS-CLARIFICATION, read the comment thread: a later member comment superseding the decision (here, rerouting through Speckit spec #799 under umbrella #797) changes the issue's current state and resolves the conflict. Treat comments as part of the issue content — the body is not the whole record. The same applies to stale figures: #714's body cited @opencode-ai/plugin 1.4.10 while the follow-up comment verified the actual pin at 1.19.0.
