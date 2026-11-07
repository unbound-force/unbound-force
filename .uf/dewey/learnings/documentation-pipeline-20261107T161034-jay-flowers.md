---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:10:34Z
identity: documentation-pipeline-20261107T161034-jay-flowers
tier: draft
---

A reconciliation edit that opens a sibling OpenSpec delta file obligates a full-file sweep, not just the section named in the reconciliation task. PR 957 amended fix-triage-issue-dispatch-docs/design.md's FR-007 lockstep enumeration (task 4.5) yet left the same file's Migration Plan prescribing a `npm ci` node_modules refresh and claiming "CI installs from the lockfile" — both falsified by the lockless install the same PR ships. Term-class hits inside a file already open for editing are the cheapest reconciliation misses to prevent: once a sibling delta is opened for reconciliation, re-run the full three-term sweep (replaced command, removed artifact filename, renamed identifiers) across every section class — requirement text, lockstep enumerations, Migration Plan, rollout, and adoption guidance — before marking the reconciliation task complete.
