---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:36Z
identity: review-dispatch-20261102T013536-jay-flowers
tier: draft
---

Dispatch-session state keyed by host-session identity conflates host-session lifetime with workflow-instance lifetime. When one opencode session legitimately runs multiple workflow instances (/uf.unleash runs a spec-review dispatch then a code-review dispatch in the same conversation), a first-call-only session_metadata persistence guard silently drops the second instance's metadata, and a session-scoped correlation ID makes the second dispatch's runs clobber the first's agent-keyed run files — downstream fail-closed consolidation then rejects on context mismatch ("artifact commit must equal the immutable reviewed head SHA", "run has no included plan entry"). Fix pattern (shipped in unbound-force as fix-review-dispatch-correlation-release, issue #794 design D1, plus design D5 of fix-review-dispatch-recovery): mint a fresh correlation ID at the aggregate-entry boundary (dispatch_review_runs) and re-point the session map so each invocation owns its dispatch session, and persist session metadata whenever supplied (last-writer-wins) rather than only on first call. Triage practice note: an open bug issue can describe code that has since been remediated under a follow-up change — diff the issue's cited mechanics against current HEAD before queueing fix work.
