---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:19:51Z
identity: determinism-20261110T151951-jay-flowers
tier: draft
---

DE-004 change-record verification must pin the base: instruction-lint absolute line/token counts drift with the --base ref while deltas stay invariant. Verified on PR #1007 — the CHANGELOG and proposal recorded 1510→1509 lines and ~17068→~17215 tokens, while instruction-lint against the PR base sha reported 1452→1451 lines and 16419→16566 tokens; the +147-token and −1-line deltas matched exactly across both measurements. Reviewers should verify the delta (invariant) and ask for the base ref to be named when absolutes disagree; authors should record the base alongside the counts.
