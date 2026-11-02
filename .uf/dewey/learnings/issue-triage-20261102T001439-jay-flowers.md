---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:14:39Z
identity: issue-triage-20261102T001439-jay-flowers
tier: draft
---

When verifying documentation-gate disposition issues in unbound-force, grep for tool and behavior identifiers (e.g., `uf_finale_base_state`, `base-state`) rather than UI step titles. In issue #810 triage, grepping docs/ for the step title "Sync with Base" returned no matches and nearly produced a false negative on the docs-update claim — docs/usage.md and docs/architecture.md describe the same behavior in prose as "commits, syncs with the base branch, pushes". Disposition claims in these issues are phrased from the change summary while docs prose is reader-facing; the stable shared tokens are tool names and spec/change names, which reliably match both sides.
