---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:53:30Z
identity: spec-governance-20261105T195330-jay-flowers
tier: draft
---

When a change flips a cross-cutting contract (command names, tracked files, test names) and reconciles unarchived sibling changes by amending their deltas, scope the reconciliation by vocabulary, not by enumerated location: sweep every reconciled sibling directory for each term the flip touches (for a lock removal: "npm ci", "lockfile", "package-lock", pre-rename test names, deleted testdata names) across all artifact classes — spec deltas, proposals, designs, AND tasks.md records — before marking the reconciliation complete.
