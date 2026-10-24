---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-10-23T20:44:08Z
identity: determinism-20261023T204408-jay-flowers
tier: draft
---

Review-artifact producer steps belong in the tool layer, not command markdown. The review-dispatch plugin already ships fetch_pr_review_context — a tested intent-level tool that fetches a PR diff via injection-safe Bun.spawn, persists it with unlink-then-O_EXCL create plus 0o600 file / 0o700 directory permissions, and returns diff_path/diff_bytes with structured error codes (timeout, rate_limited, api_error); dispatch_review_runs independently enforces the 2 MiB diff limit with a structured prompt_size_exceeded error naming the byte count. Adding shell choreography (mkdir/chmod/gh pr diff redirect) plus existence and size pre-checks as command prose therefore violates DE-001/DE-005 (CRITICAL per the AGENTS.md determinism gate) and duplicates tested tool behavior with weaker safety: a raw `>` redirect truncates the target file before gh failure is detected, leaving a zero-byte file that a missing-only resume check will pass. The compliant pattern is to extend the existing tool (e.g., a persistent per-PR destination option writing .uf/feedback/pr-<N>/pr.diff, which already passes validateDiffPath's project-root allowlist) with plugin-suite tests, keeping only policy — immutability of the saved path, report-and-STOP semantics — as prose per DE-006.
