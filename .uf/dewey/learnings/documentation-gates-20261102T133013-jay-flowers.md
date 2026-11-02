---
tag: documentation-gates
author: jay-flowers
category: pattern
created_at: 2026-11-02T13:30:13Z
identity: documentation-gates-20261102T133013-jay-flowers
tier: draft
---

The complete two-gate documentation chain for a user-facing unbound-force change is verified across four artifacts, not just the issue tracker: (1) the proposal's Impact section names both gates (repo-local docs issue per AGENTS.md + unbound-force/website sync issue per the org constitution's Cross-Repo Documentation rule); (2) a bookkeeping task files both issues before PR merge and records their numbers; (3) the change's notes.md carries both issue numbers with their gate mappings; (4) the PR body references both issues in prose while Closes refs stay reserved for the originating feature issues. CHANGELOG entries follow the repo convention of citing the Spec path and Fixes refs only — docs-issue numbers are not required in CHANGELOG entries (the port-uf-rebase-pr scenario that required them there was change-specific; model-matrix-scaffold-deployment and init-override-stubs-gitignore-sync both omit them). Verified in the PR #918 Curator review: unbound-force/unbound-force#916 + unbound-force/website#369, both open and filed before merge.
