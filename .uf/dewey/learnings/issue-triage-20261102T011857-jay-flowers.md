---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:18:57Z
identity: issue-triage-20261102T011857-jay-flowers
tier: draft
---

Triage of follow-up refactor issues must re-verify counts against HEAD — they go stale fast in active areas. unbound-force #769 (filed 2026-10-19) cited a 5,992-line review-dispatch index.ts and seven locally-defined withBun test-helper copies; by 2026-11-01 the plugin measured 8,634 lines, a shared .opencode/test/helpers/bun-runtime.ts existed with two of the seven original suites migrated (plus five newer suites importing it), six local withBun copies remained, and a post-filing suite (read-diff-chunk.test.ts) had accreted its own withBun and scratchProject copies. Pattern: state current-state numbers in the triage assessment, and split operationally independent parts (test-only work vs. production + scaffold-twin work) — shared provenance survives the split via issue links.
