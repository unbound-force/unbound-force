---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:17:35Z
identity: issue-triage-20261102T031735-jay-flowers
tier: draft
---

Triage checkpoint extension for prose-to-tool extraction issues filed before adjacent tooling lands (verified on unbound-force #667 generate_pr_body, filed 2026-10-05 under #659 Phase 2): the pre-VALID content diff must check responsibility overlap, not just prose survival. By triage time, /uf.finale had gained uf_finale_pr_context (returning templateSections, diffStat, originatingIssue, commitLog) and uf_finale_create_pr, and PR #737 added unleash-summary threading — so the issue's stated "template detection" responsibility now partially lives in the context tool, and its cited line range (uf.finale.md 392-454) has drifted with step renumbering (body assembly now sits in Step 6c). The target assembly prose (section mapping, Closes #N insertion, template heading mapping, no-headings fallback) is confirmed still present, so the issue stays VALID, but the implementing spec must reconcile the boundary: the new tool should consume structured inputs from the existing context tool rather than re-implement detection. Rule: extraction issues age in two dimensions — the cited prose moves, and neighboring intent-level tools absorb sub-responsibilities; triage both before VALID, and record boundary reconciliation as an implementation note rather than a clarification blocker when the core target prose is confirmed intact.
