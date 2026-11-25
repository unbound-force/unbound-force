---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-25T04:00:25Z
identity: test-quality-20261125T040025-jay-flowers
tier: draft
as_of: "unbound-force@15b2ed01d5d1d00d9cc6797040a2d55fcd01f02b"
---

When a plugin tool documents a closed set of failure codes via an exported const (e.g. REVIEW_HEAD_ROOT_ERROR_CODES in uf-workflow/index.ts), the emission-side tests must enumerate every code in the set — including the catch-all — not just the semantically distinct ones. The reliable shape: iterate the exported const and map each code to one triggering fixture asserting the full envelope (status error, code, retryable where the contract defines it, message content), mirroring the documentation-side parity loop that already iterates the same const. In PR #1211 (review-head-preflight), five of six codes got exact envelope assertions while the catch-all REVIEW_HEAD_ROOT_ERROR went unpinned because no "unexpected error" fixture existed — the missing-args test hit its emission site but asserted only status, and the outer failureFromError catch was never triggered. Coverage floors do not catch this: statement coverage of both emission sites was 100% while the emitted code value remained contract-unverified.
