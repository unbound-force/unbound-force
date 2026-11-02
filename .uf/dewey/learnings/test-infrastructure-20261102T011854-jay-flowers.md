---
tag: test-infrastructure
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:18:54Z
identity: test-infrastructure-20261102T011854-jay-flowers
tier: draft
---

CS-004 shared-helper extraction does not stop duplication from regenerating without a structural guard. In unbound-force's .opencode/test/, helpers/bun-runtime.ts was extracted citing default.md CS-004, yet only 2 of the 7 suites named in issue #769 migrated to it (dispatch-status, dispatch-review-runs) while 5 still carry local withBun definitions — and read-diff-chunk.test.ts, absent from the issue's seven-suite inventory, now carries a fresh local copy beside the shared helper. Later shared modules (aggregate-dispatch-fixtures.ts, gh-spawn-mock.ts) repeat the same "older copies migrate opportunistically" posture. Durable closure pairs the one-time migration with a guard — e.g., a test asserting no `function withBun` definition outside helpers/, or a suite template that imports the helper — so the shared module stays the path of least resistance for new suites.
