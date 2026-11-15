---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T18:13:43Z
identity: determinism-20261115T181343-jay-flowers
tier: draft
as_of: "unbound-force@88deba6f4dd4024b88f4306e4f717b44f0dc37d7"
---

When a PR modifies instruction files, the DE-004 change-record requirement is cheapest to satisfy by pasting the instruction-lint delta block into the PR description: `.uf/scripts/instruction-lint --json --base <base> --changed` emits a `delta` object with per-file lines/tokens/loadTokens base→head counts plus totals — the paste-ready, authoritative numbers (no manual counting; the reviewing Determinist re-derives the same block from the linter, so the record and the review agree by construction). Verified on unbound-force PR #1100: the prior council's DE-004 MEDIUM (filed as #1114) was remediated exactly this way — the PR body's "+6 lines, +78 tokens, +78 loadTokens" row for .opencode/commands/uf.run-pipeline-plan.md matched the linter's delta output verbatim.
