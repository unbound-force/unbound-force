---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-03T11:11:10Z
identity: review-dispatch-20261103T111110-jay-flowers
tier: draft
---

Guard review of PR #929 (bound-consolidate-dispatch-output): the RD-FR-049 superRefine rejecting `detail: "summary"` combined with `provenance` was omitted from ConsolidateDispatchInputSchema while every delivered test stayed green — the new summary-retrieval.test.ts covers the pure builders, and builder-level tests cannot catch a missing input-contract guard. The design-D7 tool-level tests (summary+provenance rejection, default-full compatibility, unknown-detail rejection, compact-wrapper-not-a-payload) are the only tests that would have caught it. Lesson: when a spec mandates a fail-closed input rejection, the regression test must invoke the tool boundary with the forbidden combination, not just the helpers.
