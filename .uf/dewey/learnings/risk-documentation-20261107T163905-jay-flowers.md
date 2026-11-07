---
tag: risk-documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:39:05Z
identity: risk-documentation-20261107T163905-jay-flowers
tier: draft
---

When a change preserves user state across a migration, re-run the risk assessment against the preserved-state population, not just fresh installs. In stop-shipping-plugin-lock (PR 957), the npm 10.9.9 Arborist crash was recorded as a fresh-scaffold risk — solved by shipping no devDependencies — while the same diff added preservePluginDevDependencies, guaranteeing that existing devDeps-carrying targets hit the crash on a runtime uf doctor accepts as healthy. The risk register, CHANGELOG rollout note, remediation strings, and docs follow-up issue all omitted the migration path. A preservation behavior silently doubles the affected population; every documentation surface scoped only to the fresh-install path drops it.
