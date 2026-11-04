---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:41:48Z
identity: review-dispatch-20261104T144148-jay-flowers
tier: draft
---

Capped-collection regression tests must pin the retention policy, not just cardinality. When a spec FR mandates WHICH N entries survive a cap (e.g., first-512 in collection order), `toHaveLength(N)` plus warning-count assertions cannot distinguish first-N from last-N (`slice(-N)`) or offset-N (`slice(k)`) retention — all keep length N and identical overflow warnings. Pin the policy with boundary content spot checks on the first and last kept entries (`arr[0]` carries the earliest fixture marker, `arr[N-1]` the Nth), on both the retrieval result and the persisted artifact. Give fixtures unique per-index markers so the spot checks are unambiguous. Precedent: .opencode/test/consolidate-dispatch.test.ts run_findings_overflow and consolidated_findings_overflow pin keep-first via `findings[511].description === "Finding 511"` plus first-dropped-absent checks; the RD-FR-054 proposals-overflow test initially omitted them (PR 949 review feedback, thread PRRT_kwDORVoNbc6zVoG8).
