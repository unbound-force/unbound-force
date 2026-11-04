---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:35:36Z
identity: determinism-20261104T153536-jay-flowers
tier: draft
---

DE-001 residuals over MCP-bound tools are governed by the AGENTS.md determinism gate's "where an executable tool is feasible" qualifier. When command prose must orchestrate per-item calls to an MCP-server tool — dewey_store_learning once per lesson proposal across the five lesson-storing commands (uf.review-pr, uf.review-council, uf.triage-issue, uf.address-feedback, uf.unleash Steps 4 and 6) — the review-dispatch plugin cannot collapse the loop into an executable tool, so the prose loop is a legitimately accepted residual, not a DE-001 violation (verified on PR 949 feedback item 12: the architect reviewer rated it LOW, stated extraction is infeasible while Dewey owns the storage tool, and named retirement via a future batch-store tool on the Dewey side). The retirement pattern is an intent-level batch tool on the tool-owning MCP-server side, and feedback-time acceptance should be paired with a durable tracking artifact — the change's documentation-gate issue or a cross-repo follow-up on the tool-owning repository — because the originating issue closes with the fixing PR and the review thread becomes the only remaining record of the retirement path.
