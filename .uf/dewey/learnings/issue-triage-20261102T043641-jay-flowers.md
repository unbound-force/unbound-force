---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:36:41Z
identity: issue-triage-20261102T043641-jay-flowers
tier: draft
---

Issue-triage duplicate scans that cover only open issues have a blind spot for in-flight pull requests. Before finalizing a VALID verdict on an enhancement issue, search the knowledge graph for PRs referencing the issue number ("Closes #N" / "Fixes #N"): an existing implementing PR corroborates that the gap was real and actionable, but changes triage disposition — flag the PR in the reasoning with an index-snapshot caveat (merge/close state is not determinable from indexed content) instead of recommending fresh spec work. A merged implementing PR auto-closes the issue via its closing keyword; a closed-unmerged PR leaves the issue as the live tracking artifact.
