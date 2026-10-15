---
tag: spec-review
author: jay-flowers
category: gotcha
created_at: 2026-10-15T18:27:36Z
identity: spec-review-20261015T182736-jay-flowers
tier: draft
---

During /uf.unleash review rounds, Divisor agents propose learnings that sometimes refute or refine claims still present in the in-flight OpenSpec artifacts (proposal.md, design.md). Capturing the correction as a Dewey learning does not amend the artifact: if the change merges unchanged, the canonical spec record retains claims the project's own knowledge base refutes. Observed in opsx/address-feedback-model-normalization (PR #733): design D4's 'the delta spec's scenarios are enforced by existing drift tests plus plugin tool tests' versus the command-contracts learning showing the MUST NOT content-absence scenario is enforced by neither mechanism, and the proposal's 'the tool implements the exact 6-step algorithm the prose described' versus the tool-extraction learning's documented deliberate -N-guard deviation ('behavior pinned to the tested tool' is the accurate framing). Pattern: at the end of each review round, cross-check accepted learnings against explicit claims in the in-flight spec artifacts and amend refuted claims before merge — learnings are the correction record, not the correction itself.
