---
tag: opencode-v2-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:30:16Z
identity: opencode-v2-migration-20261102T003016-jay-flowers
tier: draft
---

Runtime-gated dual-emission scaffolds — tools that emit different asset shapes (v1-shaped vs v2-native) based on the detected runtime version — must carry drift tests for every emitted shape, not just the canonical one. A scaffold that emits two shapes but drift-tests only one silently diverges on the untested shape, and the divergence surfaces only on machines running the uncovered runtime. Replicator's agent-kit dual-emission under the OpenCode v2 migration (unbound-force issue #804) mandates drift tests for both shapes; the same requirement should travel with every dual-emission scaffold adopted during the migration.
