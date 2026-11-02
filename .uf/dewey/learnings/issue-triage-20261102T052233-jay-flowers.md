---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:22:33Z
identity: issue-triage-20261102T052233-jay-flowers
tier: draft
---

Cross-repo compliance issues: when an audit issue is filed against a meta/governance repository but every named gap lives in sibling repositories, triage it as valid against the governing standard (constitution, convention pack) but recommend splitting into per-repo issues or transferring — the fixes require PRs in each affected repo, so a single centrally-filed issue cannot be closed by one PR. Pattern: unbound-force issue #560 documented constitutional pinning violations with file/line evidence, but all named targets (containerfile/ Containerfiles and workflows, zero-dot-force/vibe-check) live outside the filing repo.
