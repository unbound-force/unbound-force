---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:56:07Z
identity: issue-triage-20261102T035607-jay-flowers
tier: draft
---

Triage of issue #652 (uf.review-pr lacks progress tracking) revealed a stale-exemplar hazard: the issue cites uf.unleash.md's in-file fenced checkbox checklist (updated in-place with the Edit tool) as the pattern to replicate, but uf.unleash.md has since been refactored — the in-file checklist was replaced by checkpoint-tool auto-checkpointing with get_progress/parse_tasks re-anchoring, and TodoWrite was demoted to a RECOMMENDED display aid subordinated to durable checkpoint state. Issues older than a command-family refactor can describe remediation patterns that no longer exist. When triaging feature requests that cite a sibling command's design, verify the cited pattern still exists in the current file before endorsing the requested fix shape, and redirect implementers to the current canonical pattern (for progress tracking: uf.review-council.md's get_progress-on-startup + per-step checkpoint + clear_progress-at-completion + TodoWrite display aid). Additionally, command-file fixes in unbound-force must update the scaffold mirror under internal/scaffold/assets/opencode/commands/ — drift-detection tests enforce byte parity.
