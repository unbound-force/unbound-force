---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:49:58Z
identity: determinism-20261110T194958-jay-flowers
tier: draft
---

When a quick-tier security fix's compliant shape exceeds the tier's scope, escalate the tier rather than landing an interim prose choreography. On the triage-issue heredoc removal (#1011/PR #1024), removing the heredoc safely requires splitting the shell block into a mktemp -> file-write -> gh api sequence; the compliant terminal shape is an intent-level tool (DE-005), which is spec-tier work (new plugin tool, DE-007 seam tests). The determinism pack provides no interim-mitigation carve-out for DE-005 (unlike DE-008's delegation prose, which explicitly names prose binding as interim), so landing the three-call choreography as the final shape leaves a CRITICAL that no amount of iteration resolves — the correct move is MP-009 tier escalation or splitting the tool extraction into a tracked spec change before merge, not iterating the prose fix.
