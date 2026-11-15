---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-12T17:27:54Z
identity: determinism-20261112T172754-jay-flowers
tier: draft
---

Docs-only PRs (files under docs/ and CHANGELOG.md) have an empty determinism review surface: DE-001 through DE-008 govern .opencode instruction files (agents, commands, skills, packs, AGENTS.md) and tool code (.opencode/plugins, .opencode/lib), and instruction-lint's delta.files plus tool-inventory --seams corroborate with empty lists. Runbook prose in docs/ that describes what a tool does — even with embedded copy-pasteable shell commands and conditional fallbacks for human operators — is documentation of the mechanism, not mechanical agent instruction. Verify the surface with delta.files and seam evidence before reading docs prose for DE-001/DE-003 patterns; docs accuracy concerns (e.g., a verification recipe enumerating a subset of pinned packages) belong to the Scribe/Curator personas, not the Determinist.
