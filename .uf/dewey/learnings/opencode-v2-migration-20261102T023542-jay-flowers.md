---
tag: opencode-v2-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:42Z
identity: opencode-v2-migration-20261102T023542-jay-flowers
tier: draft
---

When a distribution channel installs a platform's new major version unpinned (e.g., `uf setup` installing latest opencode, delivering v2 binaries to hero repos while v1 plugin implementations cannot run there) and the two majors cannot coexist on one machine (v1/v2 both use the `opencode` command), a forced-cutover migration is avoidable via dual-entry plugin packaging: one package exporting `{ ...Plugin.define({id, setup}), server() }` loads on v2 (reads define+setup) and on v1 >= 1.18.29 (reads server()), making rollout order irrelevant across the fleet. Verified during #714 triage (umbrella #797 verified facts): `@opencode-ai/plugin` is pinned at 1.19.0 in both `.opencode/package.json` and `internal/scaffold/assets/opencode/package.json`, satisfying the v1 floor, so the dual form works on v1 as-is. Operational prerequisites for the accompanying gatekept pin change: keep zod as the single source of truth for tool args and emit v2 JSON Schema via `z.toJSONSchema()`; capture `ctx.location.directory` at setup (v2 removes `directory` from tool context); document rollback as plain revert — `server()` keeps working, so reverting restores v1-only behavior without stranded state; and re-resolve the exact v2.0.x pin at work start rather than trusting the version quoted when the tracking issue was filed.
