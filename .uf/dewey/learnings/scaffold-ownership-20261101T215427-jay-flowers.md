---
tag: scaffold-ownership
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:54:27Z
identity: scaffold-ownership-20261101T215427-jay-flowers
tier: draft
---

Scaffold file-ownership reclassifications (tool-owned → project-owned) change user-visible re-run dispositions: `uf init` reports `Skipped` instead of `Updated` for a locally modified `.uf/reviewer-capabilities.yaml`, and canonical capability-list updates stop auto-flowing to already-seeded repositories (the same trade-off previously accepted for `.uf/review-matrix.yaml`). Treat disposition changes as upgrade-path behavior requiring release-notes coverage, not just a changelog entry — users read `uf init` dispositions to confirm local customizations (custom `reviewers:` entries) survived a re-run. Pairing the lost auto-flow with a fail-closed dispatch error that names the missing agent and the exact file to edit keeps the trade-off loud and actionable rather than silent.
