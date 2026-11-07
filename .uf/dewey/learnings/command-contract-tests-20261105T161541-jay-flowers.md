---
tag: command-contract-tests
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:15:41Z
identity: command-contract-tests-20261105T161541-jay-flowers
tier: draft
---

Command-content contract coverage is per-command, not per-tool. When a plugin-tool contract migration rewrites command consumers, validate each migrated command file against the command-contract suite's covered set before assuming content pins exist: internal/scaffold/command_contract_test.go covers a fixed set (uf.review-council, uf.triage-issue, uf.fix-ci, uf.address-feedback, speckit.testreview, uf.review-pr, uf.unleash), so commands outside it — uf.finale.md and uf.rebase-pr.md in PR #948 — ship zero content pins even when the tool layer carries exemplary tests and another command's test holds a presence-only bare-tool-name clause for the same tools. Grep the tool name across both test trees separately (TS plugin suite vs Go scaffold suite) to separate tool-layer coverage from consumer-layer coverage: a bare tool-name Contains assertion passes any parameter drift because the name string survives the change.
