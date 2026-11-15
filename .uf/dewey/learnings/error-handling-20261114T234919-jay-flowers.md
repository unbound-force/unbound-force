---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-14T23:49:19Z
identity: error-handling-20261114T234919-jay-flowers
tier: draft
---

When a tool's output can be silently invalid (e.g., a dispatched reviewer child that exhausts its output budget while reasoning and returns empty final text), classify it as a failed, retryable result inside the producing tool rather than returning success with an advisory flag. Per-consumer filtering of a "success + warning" shape re-introduces mechanical filtering logic into every consumer and fails open the moment one consumer forgets to check. Pattern from unbound-force issue #1052 (openspec change review-dispatch-a-reviewer-with-empty-ou): the agent executor returns status failed with error code child_empty_response at the boundary, so dispatch consumers inherit run_counts.failed, workflow_verdict null, and exclusion from successfulPersonaVotes with zero consumer-side changes; a defense-in-depth recheck at dispatch finalization covers callers that bypass the executor. The explicitly rejected alternative — success plus a warning each consumer must remember to exclude — is the fail-open shape to avoid when designing status classification at tool boundaries.
