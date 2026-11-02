---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:17:39Z
identity: issue-triage-20261102T031739-jay-flowers
tier: draft
---

Triage supersession check for prose→tool extraction issues: an issue filed mid-epic can be architecturally superseded by a later-landed spec change, not merely line-drifted. unbound-force#667 (generate_pr_body, filed 2026-10-05 citing uf.finale.md lines 392-454 of the then-~1,089-line file) predates the landed finale-token-reduction change (Phase 3 of the same #659 epic), which rewrote uf.finale.md to 399 lines, already shipped template detection inside uf_finale_pr_context (.opencode/plugins/uf-finale/index.ts parses PULL_REQUEST_TEMPLATE.md headings into templateSections), moved finale composite tools into a new uf-finale plugin (uf-workflow keeps shared primitives), and explicitly preserved "PR body authoring" as an LLM decision point. Beyond the #675 content-diff rule (learning/issue-triage-20261102T025951-jay-flowers), four supersession checks drive the verdict: (1) grep the proposed tool name and each stated responsibility against current plugin code — an already-shipped responsibility silently halves the issue's scope; (2) reconcile the issue's placement directive ("add to uf-workflow plugin") against the post-change plugin organization; (3) check whether a landed change's proposal explicitly preserved the behavior the issue proposes to extract — reversing a documented design decision requires the issue to engage it; (4) diff the issue's arg schema and section list against the current prose inputs — omitted inputs (unleashSummary, review-council Known Issues findings, templateSections, behind-base warning) mean tests written from the issue as filed pass green while dropping live PR-body behaviors.
