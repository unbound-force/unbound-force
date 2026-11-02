---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:00:42Z
identity: issue-triage-20261102T010042-jay-flowers
tier: draft
---

Spike and audit issues that hardcode repository counts in their verification checklists drift from the source of truth as the repo grows. Example: unbound-force issue #798 ("16 agents load, 32 commands work") vs AGENTS.md at triage time documenting 17 active agents (plus the hidden model-matrix-classifier) and 33 command files. A hardcoded count causes the verification sweep to silently skip the most recently added agent or command — the exact newest surface most likely to carry v2 incompatibility. Issue authors should point executors at the authoritative count source (AGENTS.md Project Structure) or instruct deriving counts from the filesystem at execution time; triagers should cross-check hardcoded counts in verification checklists against AGENTS.md before confirming scope.
