---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-10T22:45:10Z
identity: determinism-20261110T224510-jay-flowers
tier: draft
---

UF command templates now carry two divergent remediation patterns for quoted-heredoc breakout. Pattern 1 (uf.address-feedback.md, #1007): route posting through intent-level plugin tools — Guardrail 8 mandates post_feedback_replies / file_feedback_issue, which own temp files, serialization, gh invocation, and cleanup, with composed text arriving as a validated parameter. Pattern 2 (uf.triage-issue.md 4.3/4.4, #1011/PR #1024): rewrite the choreography as numbered prose — bash mktemp/chmod/echo, host file-write to the printed path, bash gh api --input + rm — leaving the agent to thread the temp path via placeholder substitution and hand-JSON-escape untrusted text. Only pattern 1 satisfies both the security fix and determinism DE-005: the choreography is a judgment-free three-call chain whose every argument after the first call is determined by the first call's printed output, and the pre-change single-bash-block form was DE-005-exempt by the one-call calibration — the rewrite itself introduced the chain. Heredoc-removal changes should extract the posting step to an intent-level tool (post_issue_comment / create_child_issue shape, with registration/output-shape/round-trip seam tests per DE-007) instead of re-teaching the shell dance in prose; a deferral must be a tracked issue naming the extraction and its seam layers, not a design-note narration.
