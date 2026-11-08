---
tag: scaffold-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:45:33Z
identity: scaffold-testing-20261105T164533-jay-flowers
tier: draft
---

In the uf scaffold engine, spec-delta scenarios that promise survival of a foreign or user-owned file across `uf init` (e.g. SC-FR-003 "Lock written by another tool" in openspec/changes/stop-shipping-plugin-lock) need a composed acceptance test — pre-write the file into a t.TempDir target, run Run(), assert the bytes are unchanged — because the piecewise assertions (isToolOwned/isDivisorAsset predicate flips, install-command --no-package-lock pinning) leave an inference gap: internal/scaffold carries live removal machinery (scaffold_stealth.go prune paths at :365/:398/:537, migration os.Remove at scaffold.go:577) whose future interaction with an unrecognized .opencode/ path would violate the scenario while every predicate test stays green. Verified during PR 957 feedback triage (item 15); corroborated by learning scaffold-ownership-20261102T035446-jay-flowers on isToolOwned/skip semantics (user-owned skip path records only a Skipped disposition and performs no disk comparison).
