---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:21:14Z
identity: documentation-pipeline-20261110T152114-jay-flowers
tier: draft
---

Docs-impact assessment for a scaffold-behavior change must grep the user-facing docs tree for behavioral claims about the changed surface, not just update the canonical ownership docs. merge-plugin-runtime-deps updated docs/architecture.md, AGENTS.md, and CHANGELOG.md but missed docs/osv-transitive-scan.md's "Picking up the fix in a hero repository" runbook, which states that re-running uf init "rewrites" .opencode/package.json keeping only devDependencies and that uf doctor "compares each direct pin" — both claims made stale by the runtime-dependency merge. Runbook sections embedded in security-remediation docs describe current tool behavior and drift silently; a repo-wide grep for the changed artifact path (e.g. ".opencode/package.json") across docs/ and README.md surfaces them at impact-assessment time.
