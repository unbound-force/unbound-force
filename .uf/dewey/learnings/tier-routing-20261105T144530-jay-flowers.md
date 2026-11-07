---
tag: tier-routing
author: jay-flowers
category: gotcha
created_at: 2026-11-05T14:45:30Z
identity: tier-routing-20261105T144530-jay-flowers
tier: draft
---

Tier-routing risk floors must be evaluated against a changeset's final shape, not its initial description. Bundled quick-tier fixes that grow to include a file migration/deletion path (#933's migrateSkillDir in unbound-force PR #948) or a plugin-tool input-contract break (#947's branch-optional-to-required on git_push/git_push_force_with_lease) trip the tier-routing Phase 4 floors ('Data-loss paths: migrations, deletions, destructive operations'; 'Breaking a public API contract') that forbid the quick tier 
