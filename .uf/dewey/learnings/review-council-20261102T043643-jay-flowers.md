---
tag: review-council
author: jay-flowers
category: context
created_at: 2026-11-02T04:36:43Z
identity: review-council-20261102T043643-jay-flowers
tier: draft
---

In the unbound-force pipeline, originating-issue state terminates at PR creation, not at review time. /uf.unleash tier routing records issue_number in branch-scoped checkpoint state and OpenSpec changes may carry originating_issue in .openspec.yaml, but /uf.finale consumes both solely to emit `Closes #N` in the PR body. The review-context skill's Protocol 2 (Issue Linking) is explicitly conditional — its consumer table marks /uf.review-council as "Protocol 2 conditional (explicit PR number only)" and the protocol self-skips for local reviews without a PR body, which is exactly the pre-PR /uf.review-council path inside /uf.unleash Step 6 — and even when it fires, gh fetch failures continue without blocking (advisory, not a gate). divisor-guard's intent-fidelity checks are grounded in spec artifacts ("the original spec's stated goals and acceptance criteria"), not in the originating GitHub issue. Net effect: the pre-PR review council has no machine-readable path from a branch back to the originating issue's acceptance criteria, so quality verdicts ship without intent validation. Issue #563 proposes closing this gap with an intent-satisfaction check plus required issue-ID threading.
