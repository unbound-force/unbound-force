---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:38Z
identity: review-dispatch-20261102T013538-jay-flowers
tier: draft
---

Triage signal for unbound-force review-dispatch failures: when consolidate_dispatch fails closed with "artifact commit must equal the immutable reviewed head SHA" together with "run <id> has no included plan entry", suspect dispatch-session identity collision (one correlation ID per opencode conversation), not bad dispatch inputs. Symptom issues #760 (/uf.unleash spec-review-then-code-review) and #746 (/uf.address-feedback per-item Tier 2 finalization) shared root cause #794: sessionCorrelationMap minted one correlation per conversation and never released it, and session_metadata was persisted first-call-only, so a second dispatch in the same conversation had its metadata silently dropped. Fixes landed in two layers: design D5 of fix-review-dispatch-recovery made session-metadata persistence replace-with-latest, and design D1 of fix-review-dispatch-correlation-release gives every dispatch_review_runs invocation a freshly minted correlation. Recovery for an already-stranded dispatch: from a fresh opencode session, call consolidate_dispatch with the explicit correlation_id plus provenance — the finalizer's EEXIST collision path persists under a fresh artifact UUID without overwriting the first workflow's artifact.
