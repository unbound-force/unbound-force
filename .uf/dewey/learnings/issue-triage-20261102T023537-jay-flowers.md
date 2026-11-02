---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:35:37Z
identity: issue-triage-20261102T023537-jay-flowers
tier: draft
---

Stale figures in a long-lived tracking issue's body do not invalidate the issue when later comments or downstream authoritative artifacts correct them — comments are the living update channel, and the body remains the technical reference. unbound-force/unbound-force#714 (opencode v2 adoption) cites @opencode-ai/plugin 1.4.10 and 16 agents/32 commands; the comment thread and umbrella #797's Verified Facts section both record the verified pin as 1.19.0 (matching AGENTS.md Technology Stack), and current AGENTS.md documents 17 active agents and 33 command files. Triage verdicts must verify current state against AGENTS.md and the tree before judging validity; body staleness alone is not an INVALID signal. Related org pattern: large cross-repo migrations route through an umbrella issue (#797) with per-phase tracking issues (#714 plugin port core, #800 CI v2 leg, #801 Phase 2 scaffold dual-emission) plus a Speckit strategic spec (#799) as the artifact of record — issues track execution, the spec holds requirements.
