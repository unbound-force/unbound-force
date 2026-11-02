---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:14:09Z
identity: issue-triage-20261102T021409-jay-flowers
tier: draft
---

In the unbound-force meta repo, instruction text that hard-codes ordinal constitution-principle counts (e.g., "the four org principles (I, II, III, IV)" in .opencode/agents/constitution-check.md) goes stale on every constitution amendment, and because live .opencode/ files are drift-enforced byte-identical with internal/scaffold/assets twins, the fix cannot ship on a feature branch — it requires a chore/uf.init-sync-style scaffold-sync branch updating both copies. When triaging documentation issues that surface such gaps, recommend splitting the scaffold-sync work item into its own issue so each work stream maps to exactly one PR and branch type, and so docs-only work is not blocked behind a scaffold-sync window.
