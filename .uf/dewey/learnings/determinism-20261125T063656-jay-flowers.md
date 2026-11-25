---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-25T06:36:56Z
identity: determinism-20261125T063656-jay-flowers
tier: draft
as_of: "unbound-force@e8eba4889b2e87b6dd4870ef340fc27ed7985c8e"
---

DE-004 line-delta claims in a CHANGELOG entry go stale whenever a fix commit lands after the entry was written: unbound-force PR #1211 recorded instruction-file deltas of 599→634, 1309→1349, and 1487→1538, but the PR's authoritative diffstats are +47/−5, +54/−9, and +63/−5 (net +42/+45/+58), putting the true head counts at 641/1354/1545 — and an earlier council round had already flagged the same recorded numbers as stale against linter measurements (637/1353/1543) at the previous head. The recorded deltas were never refreshed when the DE-003 fix commit grew the files. Rule: verify recorded line deltas against git diff --numstat base..head (delta check is independent of the recorded base counts) immediately before PR finalization, or render the counts with a tool at entry-writing time so they cannot drift with later fix commits.
