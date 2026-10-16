---
tag: fix-tool-args-schema-mismatch
author: jay-flowers
category: gotcha
created_at: 2026-10-16T16:13:22Z
identity: fix-tool-args-schema-mismatch-20261016T161322-jay-flowers
tier: draft
---

The review-dispatch plugin cannot finalize a second dispatch within one opencode session: dispatch_agent_run only persists session-metadata.json on the first call for a parent session (isFirstCall guard), so when /uf.unleash runs its spec-review dispatch followed by its code-review dispatch in the same session, the second dispatch's session_metadata (with the new head SHA) is silently dropped. consolidate_dispatch then rebuilds its payload from the stale first-dispatch metadata and deterministically rejects the true provenance with "artifact commit must equal the immutable reviewed head SHA" — an identical mechanical retry fails identically because the rejection is deterministic, not transient. The remedy is to restart opencode (clearing the in-memory sessionCorrelationMap) and re-run: all committed implementation work is preserved, the dispatch re-executes at modest cost (~$0.28), and the fresh process persists metadata correctly on its first dispatch. A related transient variant: a stale plugin-server module served an 8-reviewer manifest and failed plan_review_dispatch with "discovered agent divisor-herald has no reviewer-manifest entry" even though the on-disk reviewer-capabilities.yaml was verifiably correct — an identical-input re-invocation succeeded, confirming in-process staleness rather than a config defect. When plugin-driven gates fail inexplicably, verify disk config first, then suspect process state.
