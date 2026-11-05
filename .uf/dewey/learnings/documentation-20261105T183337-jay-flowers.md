---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:33:37Z
identity: documentation-20261105T183337-jay-flowers
tier: draft
---

Documentation tables that state parameter ranges drift from enforcement when bounds live in split validator clauses: a zod `.positive().max(N)` chain enforces no floor, so a docs row like `60,000 -- 14,400,000 ms` overstates the accepted range, while an adjacent row backed by `.min(60_000)` is accurate — making the table look uniformly enforced. When reviewing parameter-bound documentation, diff every documented bound against its validator clause; kind-dependent bounds split across a schema max and a superR (From divisor-guard review of PR #950.)
