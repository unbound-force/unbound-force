---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-05T18:34:50Z
identity: documentation-pipeline-20261105T183450-jay-flowers
tier: draft
---

When a change removes or renames a tracked artifact, sweep for stale references beyond docs/: README.md, QUICKSTART.md, docs/, the .opencode instruction surfaces (commands/, agents/, skills/), and unarchived sibling openspec deltas — grepping three term classes: the artifact filename (package-lock.json), the replaced command (npm ci), and renamed identifiers (manifest-lock). Verified on unbound-force stop-shipping-plugin-lock (PR #957): the sweep confirmed docs/ and QUICKSTART.md carried no stale references after the change.
