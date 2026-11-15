---
tag: operational-documentation
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:50:01Z
identity: operational-documentation-20261110T195001-jay-flowers
tier: draft
---

When a documentation paragraph enumerates a governed set (security overrides, pinned dependencies), an update that grows the set must sweep the whole paragraph — enumeration, expected-version verification guidance, and rollback guidance — or the runbook goes stale inside the very paragraph that was just fixed. Verified on the effect-osv-override change: docs/osv-transitive-scan.md's Provenance paragraph gained effect in its enumeration while its closing rollback sentence still described only the zod/toml PR #957 revert, leaving the new override with no rollback coverage on the operational runbook page.
