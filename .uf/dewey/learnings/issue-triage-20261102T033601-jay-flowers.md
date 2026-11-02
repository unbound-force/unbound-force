---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:36:01Z
identity: issue-triage-20261102T033601-jay-flowers
tier: draft
---

Triage trap: architectural issues age faster than their filing date — verify against current repo state plus resolution provenance, not filing-time history. Issue #658 (create a review-orchestrator agent to replace cobalt-crush-dev in /uf.unleash review delegation, filed 2026-10-05) was factually accurate at filing — cobalt-crush-dev.md is exactly 270 lines and the child-driver delegation existed — but was overtaken by events: the unleash-direct-review-dispatch change (2026-10-14, commit 877d5e7, Increment 1 of issue #716) eliminated the child driver session entirely, delivering all of the issue's stated benefits via direct review-dispatch plugin calls from the parent session. Verifying an agent-architecture issue requires checking the live command files AND resolution provenance (Dewey learnings, openspec/changes directories, compiled pages): an issue whose claims verify against filing-time history can still be inactionable against current main, and implementing it would create orphaned agents (zero-waste violation). Prefer closure reasoning of the form "superseded by <change>, commit <sha>" over a bare INVALID when the report was correct at filing — it preserves the diagnostic credit and gives the reporter a traceable resolution. Also check whether postdating governance (e.g., Constitution Principle VI / determinism pack DE-005, which prohibit judgment-free driver agents) has made the proposed mechanism itself non-compliant since filing.
