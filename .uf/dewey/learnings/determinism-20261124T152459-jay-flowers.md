---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-24T15:24:59Z
identity: determinism-20261124T152459-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
---

Spec-reviewing CI changes for determinism: a mechanical procedure placed in a GitHub Actions `run:` block is executed by the CI runner, not re-reasoned from agent prose, so DE-001's letter (commands/agents/skills) does not reach it — but when a design justifies keeping such a block untested and inline-duplicated by claiming its spec scenarios "pin the behavior," the tasks must actually execute those scenarios; acceptance criteria with no verification task pin nothing. Check especially the scenarios where a detector bug escapes the fail-safe direction: rename handling is the classic case — without --no-renames, a rename into the skip prefix reports only the new path and can skip CI on a code deletion, which a "fail-safe direction" claim about indeterminate states does not cover.
