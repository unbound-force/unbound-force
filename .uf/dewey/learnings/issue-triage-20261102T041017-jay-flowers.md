---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:17Z
identity: issue-triage-20261102T041017-jay-flowers
tier: draft
---

Scaffold-sync chores ("re-run uf init to refresh scaffolded files") recur across the Unbound Force org as per-repo tracking issues (e.g., replicator issue #52, unbound-force issue #591 for the complytime org). The strongest precedent (replicator #52) defines file-level acceptance criteria — old unprefixed command files removed, new `uf.`-prefixed files present, skills/packs updated to latest scaffold versions, CI-equivalent checks pass — while weaker variants verify only with generic gates (peer review + CI green). Because `uf init` re-runs overwrite tool-owned files deterministically (openspec archive 2026-03-26-uf-init-command, "re-scaffold updates command" scenario), triage of these issues should push verification toward the file-level checklist pattern and should check whether cited known scaffold defect classes (org-specific content leakage, pre-rename command references, schema installed without consumer config alignment) have executable checks rather than conceptual test sketches in comments. Generic CI-green verification confirms the process ran, not that the scaffold landed correctly.
