---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-08T17:17:05Z
identity: documentation-pipeline-20261108T171705-jay-flowers
tier: draft
---

An OpenSpec change that rewrites a user-facing command workflow predicts its own documentation gaps through tasks.md: when no phase touches the repo's docs pages, the pages that documented the original workflow go stale. Verified on feedback-loop-scope (PR #993): tasks.md phases 1-4 cover implementation, scaffold mirrors, tests, and validation only, while docs/usage.md "Address Review Feedback" and docs/architecture.md "Addressing Feedback" still describe the replaced workflow (four decision options, review-council gate inside the loop) even though CHANGELOG.md and AGENTS.md were updated. Review heuristic for the Curator: for each changed .opencode/commands/ file, grep docs/ for the command name and diff the documented workflow against the change's spec deltas; the original address-feedback change's phase-4 documentation tasks name exactly those touchpoints (docs/usage.md, docs/architecture.md), so any later change to the same commands inherits them.
