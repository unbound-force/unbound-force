---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-16T02:24:08Z
identity: review-dispatch-20261116T022408-jay-flowers
tier: draft
as_of: "unbound-force@1677fd12d4ed8969d5ac87110e35ca030b9326b1"
evidence:
  - "dewey:.uf/artifacts/dispatch/a504842e-57da-41a4-926a-884c37671964.json"
  - "dewey:.uf/artifacts/dispatch/fb460712-fb0f-49e7-bf32-4362f5c4eba8.json"
  - "dewey:.uf/artifacts/dispatch/0d1746c4-73c3-4457-85fd-087febc073d6.json"
  - "session:ses_current-uf-pipeline-1048"
---

Recovery pattern for storm-aborted delegated review runs (recorded 2026-11-16 on branch opsx/heredoc-ban-escaped-and-ansic, epic #1060, issue #1048): when a delegated /uf.unleash run aborts with a context_thrash storm (the known false-abort where reviewer sub-sessions on tight-context models compact; unfixed #1125/#1117), the stranded review dispatch is fully recoverable from the orchestrator session: (1) redispatch_failed_runs re-runs cancelled/failed reviewers with their original persisted prompts; (2) consolidate_dispatch accepts the explicit correlation_id for cross-session recovery; (3) relay_spec_review_gate computes the pause state from progress state + dispatch artifact; (4) checkpoint spec_review_decisions records the apply/override/pause adjudication of a REQUEST CHANGES iteration (a checkpoint with apply decisions returns pending_approving_iteration — expected control flow; the next iteration must consolidate to APPROVE); (5) the orchestrator then drives remaining mechanical steps directly — re-dispatch with fresh correlation for the next iteration, implement verification, code review — because orchestrator-direct dispatch_review_runs calls are NOT under the delegated-child storm detector. Consent gates (commit/push/PR) still route through pause-relay questions or the delegated /uf.finale child.
