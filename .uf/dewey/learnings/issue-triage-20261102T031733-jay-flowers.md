---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:17:33Z
identity: issue-triage-20261102T031733-jay-flowers
tier: draft
---

When triaging Constitution Principle VI prose→tool extraction issues, verify the proposed tool's input contract against the outputs of the gather-tools already in the pipeline before treating the issue's sketched signature as complete. Issue #667 (generate_pr_body) lists template detection as a tool responsibility while omitting templateSections from its args — but uf_finale_pr_context already returns templateSections, unleashSummary (whatWasBuilt/howToVerify), diffStat, commitLog, and originatingIssue, and the uf.finale.md section-6c prose also consumes Known Issues and How-to-Test/Demo content the sketch omits. The spec phase must reconcile to avoid duplicating template detection across two tools (DRY) and to cover the full body contract. Corollary: line ranges cited in month-old issues drift (issue #667 cited uf.finale.md 392-454; after later landings the prose sits at 232-254 of a 399-line file) — verify against the current working tree and pin spans with text anchors at spec time (corroborates learning spec-authoring-20261019T223602).
