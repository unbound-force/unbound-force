---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:57:00Z
identity: content-pipeline-20261107T195700-jay-flowers
tier: draft
---

Unshipped prescriptions persist across council rounds when the prescribed action lives outside the repository: on PR #957, the #964/#382 documentation-issue scope extension was prescribed by learnings at 18:40, re-verified unshipped at 19:05 in the final fix commit's own learning batch, and remained unshipped at council round 3 — because fix sessions close code findings in commits while the prescription targets issue-tracker state no fix commit can touch, and dispatched child sessions are prohibited from editing issues. Rule: re-run the unshipped-prescription audit (for each stored learning naming a concrete remedy — a test name, doc path, or issue edit — grep the fix commits for the remedy's artifact) on every council round, not just the round that stored the learning; verify issue-tracker prescriptions read-only via gh issue view before reporting; and when a prescription has already survived one advisory-level round unaddressed, report it at the severity the prescribing learning set and escalate the verdict so the author applies it before merge instead of deferring it a third time.
