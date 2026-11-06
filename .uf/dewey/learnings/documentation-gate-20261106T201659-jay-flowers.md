---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:16:59Z
identity: documentation-gate-20261106T201659-jay-flowers
tier: draft
---

When a feature branch stacks on an unmerged lower PR, the documentation-gate issue filed for the lower PR's scope does not cover the upper increment. Each user-facing increment needs its own docs-gate issue (or an explicit scope extension of the existing one) before merge — verified on unbound-force PR #959, where issue #970 and website #383 covered only PR 958's linguist/linter scope while the stacked Determinist-persona increment shipped with no coverage. Reviewers checking the documentation gate should match the issue's body scope against the full branch-vs-base diff, not the PR title or the issue title alone.
