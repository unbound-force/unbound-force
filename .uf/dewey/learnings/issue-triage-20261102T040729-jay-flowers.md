---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:07:29Z
identity: issue-triage-20261102T040729-jay-flowers
tier: draft
---

Triage in fast-moving repos must diff an issue's ask against current repository state, not the state at filing time. Issue #636 (sibling downstream-impact review, filed 2026-09-30) requested a declarative .uf/sibling-repos.yaml contract plus a downstream-impact review persona; by triage time the contract, the acquire_sibling_evidence tool (origin-verified, commit-pinned, untrusted-delimited), and review-prompt evidence injection had all landed, leaving only persona-ownership and per-sibling routing-tag deltas open. Component-by-component verification converts a binary valid/invalid call into a landed-vs-remaining delta report the maintainer can act on — preventing both stale re-work of shipped capability (zero-waste violation) and premature closure of unshipped scope.
