---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:18:34Z
identity: issue-triage-20261102T011834-jay-flowers
tier: draft
---

When triaging a bug issue, search the learning store for the issue's subsystem and compare learning timestamps against the issue creation date before recommending implementation work. Retrospective learnings captured after the issue was filed often document the implemented fix: triaging unbound-force #771 (filed 2026-10-20T00:02) surfaced learning review-dispatch-20261020T170534-jay-flowers, captured roughly 17 hours later, describing the outcome.persisted three-gate verification added to redispatch_failed_runs to close exactly the gap the issue reports. A post-filing learning that corroborates the defect and describes its remedy flips the triage next-step from "schedule implementation" to "verify the fix landed in current code, then close as completed", preventing duplicate work against an already-remedied defect.
