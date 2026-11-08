---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-08T21:33:07Z
identity: determinism-20261108T213307-jay-flowers
tier: draft
---

Retroactive change records that cite instruction-size deltas (DE-004 evidence) go stale when later fix-round commits grow the measured files. On PR #959 the CHANGELOG and proposal recorded divisor-determinism at 246 lines and divisor-linguist at 236->203/202 lines, but the round-2 bash-hardening commits subsequently grew both personas and the head state shipped at 263 and 262->230 lines — the numbers were captured mid-branch and never re-measured when the record was finalized in the last commit. Recipe: when a change record cites line/token deltas, re-run the measuring tool (`instruction-lint --base <ref>`) at the final head commit and paste its delta.totals and per-file rows verbatim; never carry numbers forward from an earlier review round. A recorded delta that does not reproduce at head is worse than no delta — it is affirmative false evidence in the record future audits read.
