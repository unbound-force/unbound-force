---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-09T20:27:30Z
identity: instruction-lint-20261109T202730-jay-flowers
tier: draft
---

When a Divisor review runs from a checkout whose working tree sits on the PR base (e.g. local main) rather than the PR head, the instruction linter's --base/--changed scope yields an empty delta.files even though the caller's diff names instruction files — the linter cannot lint prose that is not in the working tree. In that state Step 1 produces no evidence either way: review the changed instruction content from the manifest diff chunks, cross-check tool contracts from the plugin source, and record the scope limitation in the review summary instead of reporting or implying linter-scoped findings.
