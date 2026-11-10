---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-09T21:26:33Z
identity: determinism-20261109T212633-jay-flowers
tier: draft
---

Instruction-lint remediation PRs that only reword prose (IL-022/IL-023 vague-language fixes) change no procedures, so the determinism review reduces to converting IL-040/IL-041/IL-042 hints on changed lines plus DE-004 delta evidence. Gotcha: recording lint-count deltas in the PR body (e.g. "31 warnings → 16") does not satisfy DE-004 — the rule wants line/token deltas, which `instruction-lint --json --base <merge-base> --changed` emits as delta.files per-file rows plus delta.totals, ready to paste into the change record.
