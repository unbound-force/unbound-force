---
tag: documentation-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:24:35Z
identity: documentation-drift-20261104T152435-jay-flowers
tier: draft
---

A tool-output shape change must update four documentation surfaces in the same change or the command templates drift: the rendering code, the tool description, the CHANGELOG entry, and the command-template decision-point examples. PR 949's #946 fix (dispatch_review_runs non-ok summaries now ending with " — error: <message> (retryable: <bool>)") updated the first three but left all four dispatch command templates (.opencode/commands/uf.review-pr.md, uf.address-feedback.md, uf.review-council.md, uf.triage-issue.md) and their byte-identical internal/scaffold/assets/opencode/commands/ mirrors documenting only the success form — 8 files total. The drift-detection suite enforces canonical↔scaffold byte parity, not command-prose↔tool-behavior consistency, so the gap passes every gate; only the tool description (a separate channel) keeps executors informed. When changing any tool's caller-visible summary line, grep the documented output shape across both .opencode/commands/ and internal/scaffold/assets/opencode/commands/ and update every hit in lockstep.
