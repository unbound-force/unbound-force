---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T16:55:05Z
identity: determinism-20261110T165505-jay-flowers
tier: draft
---

Instruction-delta figures pasted into change records go stale when the instruction file is edited after measurement. On PR #1007 (feedback-reply-tools), the CHANGELOG and proposal recorded endpoints of 1510→1509 lines and ~17068→~17215 tokens for uf.address-feedback.md, but instruction-lint run at final head against the PR merge base reports 1452→1451 lines and 16419→16566 tokens — the deltas (−1 line, +147 tokens) still matched while the absolute endpoints had drifted by a uniform offset, meaning the measurement predated a later edit of the file. Practice: run `.uf/scripts/instruction-lint --json --base <PR base> --changed` at the final head commit, paste the per-file delta row verbatim (base/head/delta for lines and tokens), and name the base ref in the change record so the figures are reproducible as written; re-verify the sentence whenever the command file changes after the entry is drafted.
