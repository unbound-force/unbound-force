---
tag: uf-workflow-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:31:26Z
identity: uf-workflow-testing-20261103T213126-jay-flowers
tier: draft
---

Breaking plugin-tool contract migrations need consumer-side command-content pins, not just tool tests. Observed on PR #948 (push tools: `branch` made required, structured GitPushData, raw bash fallback removed from uf.rebase-pr.md): the tool itself shipped exemplary tests (scripted pushExec harness + real-git lease regression tests replaying the #947 incident), and Layers 1-2 of the DE-007 seam were verified (zod-args schema exposure test, output-shape assertions) — but the three rewritten command consumers
