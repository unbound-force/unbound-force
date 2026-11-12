---
tag: gatekeeping
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:11:50Z
identity: gatekeeping-20261110T211150-jay-flowers
tier: draft
---

A pinned-dependency security override that ships without its rollback documentation fails one conjunctive condition of the Gatekeeping Integrity exception even when every other condition holds (documented advisory, verified upstream git head, npm integrity hash, gates enforced, OSV scan green) — PR #1031's initial push hit exactly this and needed the docs follow-up commit b909c9b3 before its feedback run. Ship the rollback sentence in the same commit as the override, on both surfaces the #957 precedent names (the docs runbook's provenance paragraph and the CHANGELOG entry), and prefer the generalized phrasing "a revert of the pin commit for the override in question" so the paragraph does not need per-override re-edits. A rollback sentence that names only earlier PRs cannot restore the new override — reverting #957 would reintroduce the zod/toml advisories, not remove the effect override.
