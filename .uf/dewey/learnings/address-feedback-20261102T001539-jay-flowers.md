---
tag: address-feedback
author: jay-flowers
category: context
created_at: 2026-11-02T00:15:39Z
identity: address-feedback-20261102T001539-jay-flowers
tier: draft
---

Cross-reference hazard when triaging or reviewing /uf.address-feedback issues: change #811 (implemented by PR #816, branch quick/address-feedback-diff-path-producer) inserted a new Phase 1 §1.4 "Fetch and Save the PR Diff" — a single fetch_pr_review_context tool call producing the diff_path that §2.3.4's dispatch_review_runs consumes — and renumbered the former §1.4–1.6 (Determine Reviewer Authority, Filter and Group, Cache Check) to §1.5–1.7. Citations of "Section 1.4" (or 1.5/1.6) of this command are therefore ambiguous unless anchored to a change number or date; resolve against the #811/#816 boundary (October 2026). Verified against the .opencode/commands/uf.address-feedback.md headings 1.0–1.7 and its scaffold mirror at internal/scaffold/assets/opencode/commands/uf.address-feedback.md.
