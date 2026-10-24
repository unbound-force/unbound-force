---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-10-23T20:44:12Z
identity: determinism-20261023T204412-jay-flowers
tier: draft
---

Fixing an underspecified command step by prescribing its shell choreography in prose trades an improvisation defect for a Principle VI / DE-001 violation: the determinism pack governs new and modified instruction, so a gap-fixing bash block (mkdir/chmod/gh redirect plus size and existence checks) is CRITICAL material even inside a command whose older prose is grandfathered for opportunistic migration. When a command documents a tool-consumed input — e.g. diff_path for dispatch_review_runs — without a producing step, the compliant fix extends the tool layer: fetch_pr_review_context already saves PR diffs and returns the path, capture_review_diff performs bounded local capture, so a per-PR cache variant is a small extension with existing test infrastructure. After extraction, keep one or two tool output examples near the decision point per DE-002. Precedent: the unbound-force deterministic-instruction migration proceeds by such increments (address-feedback-model-normalization was Increment 2), and a prior input-documentation increment (address-feedback-diff-path-input) left exactly this kind of producer gap that tempts the prose fix.
