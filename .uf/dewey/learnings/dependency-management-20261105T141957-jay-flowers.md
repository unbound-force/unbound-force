---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-05T14:19:57Z
identity: dependency-management-20261105T141957-jay-flowers
tier: draft
---

When a change removes a lockfile, enumerate every install path the lock protected and re-apply the design's named mitigations to each — the path with the widest transitive graph is the one most often left inconsistent. In stop-shipping-plugin-lock (PR 957), the migration kept `--ignore-scripts` on the narrow `uf init` activation path (plugin_activation.go) but dropped it on `make plugin-test`, the recipe that installs the full vitest dev tree: design.md D2 specified only the activation command, tasks.md inherited the gap, and the prior `npm ci` had compensated with lock sha512 verification, so the migrated recipe was strictly weaker than before. Audit technique for lock-removal changes: grep for every `npm ci` / `npm install` invocation (Makefile targets, CI workflows, Go exec call sites, remediation strings) and diff each against the design's mitigation list; when a design decision names install flags for one surface, it should state explicitly whether the other surfaces follow.
