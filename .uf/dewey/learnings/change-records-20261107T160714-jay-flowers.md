---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:07:14Z
identity: change-records-20261107T160714-jay-flowers
tier: draft
---

When one PR carries multiple logical changes (feature + security pin fix) that touch the same instruction file, per-entry DE-004 line-delta claims must be explicitly scoped and must reconcile with the PR's actual numstat for that file. On unbound-force PR #958, the instruction-lint CHANGELOG entry claimed "AGENTS.md +7/-2" while the real AGENTS.md diff was +13/-4: the +6/-2 security-stack-pin documentation was described prose-only in the companion Fixed entry ("stack pin updated in lockstep") with no delta, so the change-record arithmetic could not be reconciled against the diff. Rule of thumb: when a file is touched by more than one entry in the same PR, each entry states its own scoped delta and the entries' figures sum to the file's numstat, or one entry explicitly claims the total.
