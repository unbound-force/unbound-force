---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:15:17Z
identity: issue-triage-20261102T021517-jay-flowers
tier: draft
---

Docs-gate issues that embed a per-page impact table plus the artifact contract (path, producer/consumer tools, exact camelCase field names) are triage-verifiable without source reads: unbound-force/unbound-force #739's claims matched both the live plugin tool surface and PR #745's summary exactly (11 new uf-workflow tools, 3 shared lib modules, write_unleash_summary → .uf/artifacts/unleash-summary/<slug>.json → uf_finale_pr_context). Gotcha: an issue framed at one increment can under-describe a PR that bundled later increments — #739 (Increment 2 of #716) omits the formatter-rendered exits/demo and checkpoint auto-recording deltas that PR #745 (increments 2-5) shipped and that the website twin issue (unbound-force/website#309) lists; triage should point implementers at the merged PR's full scope.
