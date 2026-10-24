---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-10-23T20:44:14Z
identity: review-dispatch-20261023T204414-jay-flowers
tier: draft
---

When a command-file change adds a new documented command-to-tool call site, update the golden-output-shape coverage ledger in the same change: either add a golden entry or an explicit cross-reference to the test file that pins the site's input variant. The suite's charter maps every documented site to executable verification (DE-007); a new site with no ledger note is a false-negative audit gap even when the tool is thoroughly tested elsewhere. Watch-for signal: the change already edits the coverage comment block for line-label recomputation, so the marginal cost of the inventory note is one line. Concrete case: quick/address-feedback-diff-path-producer added section 1.4 (uf.address-feedback.md:172, fetch_pr_review_context with pr_number) — the first documented call site for that tool, functionally pinned in fetch-pr-review-context.test.ts — without a ledger cross-reference, leaving 'Coverage: 43 of the 47 documented sites' silent about the new site.
