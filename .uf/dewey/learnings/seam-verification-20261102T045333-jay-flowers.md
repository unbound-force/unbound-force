---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:53:33Z
identity: seam-verification-20261102T045333-jay-flowers
tier: draft
---

When refreshing a documented-site inventory in a golden test (e.g., .opencode/test/golden-output-shape.test.ts), reconcile every count in the file — the top-level coverage claim, the per-plugin enumeration, and the describe-block title — against what the suite actually exercises, not just the lines naming changed sites. Partial refreshes leave pre-existing omissions surviving beneath a task completion that claims the inventory "stays exhaustive": in review-pr-consolidate-dispatch, the review-dispatch describe block exercises 15 sites (including capture_review_diff at uf.unleash.md:620, tested at line 1386) while the header enumerates 14 and the top-level claim says 44 of 48, an internally inconsistent state no failing test surfaces because the describe title (15) and the header (14) never get cross-checked.
