---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:01:07Z
identity: issue-triage-20261102T020107-jay-flowers
tier: draft
---

When triaging follow-up documentation issues, validate by substance rather than literal-quote match against current file state. Triage of unbound-force/unbound-force#747 (command-prose-pilot docs gate): the follow-up comment and the change's metrics.md both cite two AGENTS.md references as reading "already executed as /uf.unleash Step 8", but both named locations (AGENTS.md lines 213 and 252) now read "Step 6, Code Review" — the numeric portion was corrected at some point while the tracked name-based-referencing conversion ("the Code Review step of /uf.unleash") remains outstanding. Literal-quote validation would falsely conclude the fix was already applied or the issue stale; substance validation (the reference style at the named locations) shows the work remains actionable. Corollary: observed numeric drift in section-number cross-command references is itself evidence of the renumbering fragility that name-based referencing eliminates — the reference required a manual numeric correction within two weeks of the follow-up being recorded.
