---
tag: persona-design
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:17:09Z
identity: persona-design-20261106T201709-jay-flowers
tier: draft
---

Agent-persona fallbacks that enumerate changes with bare `git diff` or `git diff --name-only` compare working tree to index and return no paths on a committed feature branch; paired with an if-no-changed-paths-then-APPROVE gate, a no-diff invocation silently approves unreviewed commits. Write the fallback base-scoped (`git diff --name-only BASE...HEAD`) and widen the frontmatter bash permission to a wildcard (`git diff --name-only*` or `git diff*`): OpenCode evaluates bash permission patterns glob-style, so an exact pattern without `*` does not match commands carrying extra arguments and blocks the ref-scoped recovery. Observed reviewing the divisor-determinism and divisor-linguist personas (2026-11-06): both instruct bare-command fallbacks; the Determinist's exact-string allows compound the trap.
