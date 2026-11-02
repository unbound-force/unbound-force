---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-01T22:54:11Z
identity: issue-triage-20261101T225411-jay-flowers
tier: draft
---

Read-only triage of a documentation-gate tracking issue can resolve a conditional disposition from in-repo traceability artifacts when the issue body has gone stale: the linked change's tasks.md gate-task completion note records the filed issue numbers, and the CHANGELOG entry's Docs footer carries the same pair. Verified triaging unbound-force/unbound-force#844 (2026-11-01): the issue body's Website line still read "no website tracking issue known for this change" while openspec/changes/replicator-version-floor/tasks.md task 3.2 and the CHANGELOG entry (Docs: #844, unbound-force/website#337) had recorded both numbers since 2026-10-27 — the constitution's cross-repo gate was substantively closed and the body line was stale, not an unresolved open point. Rule for triage mode (no GitHub mutations available): before treating a gate-issue disposition as an open question needing clarification from the filer, grep the linked change directory and CHANGELOG for the issue numbers; a stale "no issue known" line means the disposition needs a body update in the implementing PR (which the issue's confirm-or-update acceptance criterion already covers), not a NEEDS-CLARIFICATION verdict and not a duplicate filing.
