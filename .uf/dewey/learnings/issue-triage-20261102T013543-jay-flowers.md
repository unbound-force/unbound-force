---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:43Z
identity: issue-triage-20261102T013543-jay-flowers
tier: draft
---

Triage process insight (unbound-force#760): the GitHub duplicate check reported "no duplicates found" (title search returned only #760 itself), but a Dewey knowledge-graph query on the issue's failure signature immediately surfaced five prior draft learnings documenting the same review-dispatch stale-session-metadata failure (2026-10-15 through 2026-10-23), sibling issue #746 (same root cause, uf.address-feedback per-item Tier 2 symptom), umbrella issue #794, and the in-flight OpenSpec change fix-review-dispatch-correlation-release implementing the issue's own suggested fix (per-dispatch correlation keying). GitHub search matches issue titles; stored learnings and change proposals match failure signatures and fix pipelines. When triaging defects in repos with a Dewey learning system, query the knowledge graph (semantic + keyword) for the root-cause signature before concluding a report is novel — it changes the disposition from "schedule new implementation work" to "verify the in-flight fix, cross-link, and close when it lands."
