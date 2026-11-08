---
tag: github-api
author: jay-flowers
category: gotcha
created_at: 2026-11-08T17:16:52Z
identity: github-api-20261108T171652-jay-flowers
tier: draft
---

GitHub's compare API (repos/{owner}/{repo}/compare/A...B) truncates the files array at 300 entries without failing. Conditions derived from compare data — line-count thresholds, per-file status checks, sensitive-path denylists — silently evaluate over an incomplete view and can pass fail-open when the true delta is larger (changes beyond the truncation boundary, including .github/ workflow files, are invisible). Any gate that branches on compare data must treat a failed or truncated response as the conservative branch, mirroring the fail-closed clause given to missing review data in the same evaluator.
