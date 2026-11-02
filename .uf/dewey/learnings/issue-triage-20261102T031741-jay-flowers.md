---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:17:41Z
identity: issue-triage-20261102T031741-jay-flowers
tier: draft
---

Line-reference drift in prose→tool extraction issues has a benign flavor that complements the #675 gotcha (misstated branch point): sibling extractions under the same parent epic shrink the host command over time, so a cited range can exceed the current file length and hold entirely unrelated content — issue #667 cited uf.finale.md lines 392-454 for PR body assembly, but the file is now 399 lines with that range holding the Branch Safety block, while the assembly prose lives in section 6 (~lines 210-260: "Closes #<N>" at 236, "Template mapping" at 248). The issue remains VALID because its substantive description matches the relocated prose. Rule: the content diff, not the line numbers, is the validity determinant — stale references with accurate substance get a re-location note for the implementer; only misstated load-bearing details block. Two spec-phase corollaries from #667: (1) tool sketches under-cover inputs the prose gained after filing (unleashSummary preferences, the review-council Known Issues section, templateSections) — the acceptance criteria are the binding contract and full section parity must be enumerated at spec time; (2) when a sketch has the new tool re-reading a file an existing input-gathering tool already returns (generate_pr_body re-reading .github/PULL_REQUEST_TEMPLATE.md vs uf_finale_pr_context's templateSections), reconcile at spec time to avoid duplicated reads and two sources of truth for template detection.
