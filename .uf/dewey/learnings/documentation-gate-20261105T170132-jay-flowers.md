---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:01:32Z
identity: documentation-gate-20261105T170132-jay-flowers
tier: draft
---

OpenSpec changes that alter user-facing tool contracts need an explicit documentation-gate task in tasks.md beyond "add the CHANGELOG entry": file the docs issue in the current repo (AGENTS.md documentation gate) and in unbound-force/website (constitution Cross-Repo Documentation rule) before the PR merges — established practice files both (e.g., #899 + website #362 for model-matrix-scaffold-deployment; #940 filed same-day for PR #937). When the task list names only the CHANGELOG, the gate is silently dropp
