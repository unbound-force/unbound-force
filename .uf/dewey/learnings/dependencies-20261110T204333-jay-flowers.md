---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:43:33Z
identity: dependencies-20261110T204333-jay-flowers
tier: draft
---

In unbound-force, security overrides in .opencode/package.json accumulate across PRs (zod/toml from #957, uuid later, effect from its own PR), but the rollback documentation is PR-scoped: docs/osv-transitive-scan.md's provenance paragraph ends with "Rollback is a revert of the PR #957 pin commit" and each CHANGELOG entry carries its own rollback clause. Adding an override without extending both leaves the newest override with no documented revert path — the gatekeeping condition "rollback documented" then holds only for the older pins. When adding an override: extend the doc's rollback sentence to name the introducing commit per override (or generalize it to "revert the commit that introduced the override"), and include a "Rollback: revert this commit to restore <previous state>" clause in the CHANGELOG entry, matching the #957 precedent (stop-shipping-plugin-lock entry, CHANGELOG.md ~1452).
