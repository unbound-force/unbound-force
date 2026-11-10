---
tag: instruction-lint
author: jay-flowers
category: pattern
created_at: 2026-11-09T21:26:40Z
identity: instruction-lint-20261109T212640-jay-flowers
tier: draft
---

Instruction-lint findings live in installed .opencode/ files, but some installed text is generated from verbatim fenced blocks inside uf.init.md: Step 5 creates the four UF-custom speckit commands, Step 6 injects/replaces their ## Guardrails sections from canonical blocks (replacing the entire section whenever a correctness marker is absent), and Step 8 injects opsx-propose guardrails. A wording fix applied only to the installed copy desyncs it from the embedded generator — fresh uf init deployments re-inject the old wording, and the lint is blind to the stale text because it sits in a fenced code block outside IL-023 scope. When fixing lint findings in speckit.* or opsx.* files, grep the pre-fix phrasing repo-wide (canonical .opencode/, internal/scaffold/assets/ mirrors, and uf.init.md generator blocks) and fix every copy in the same commit; identical before/after git blob hashes in the diff verify canonical-vs-mirror byte identity for the direct mirrors.
