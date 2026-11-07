---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:38:59Z
identity: error-handling-20261107T163859-jay-flowers
tier: draft
---

When a migration replaces a pinned resolution artifact (a tracked npm lockfile) with runtime resolution (a lockless npm install), re-validate the accepted-runtime matrix for every consumer class that still carries the old constraint set — not just the class the migration re-shapes. In PR 957 (stop-shipping-plugin-lock), the fresh-scaffold path was fixed by shipping no devDependencies, but preservePluginDevDependencies keeps the vitest 5 pair in previously-scaffolded targets, and a lockless `npm install --omit=dev` still builds the devDependencies ideal tree — crashing npm 10.9.9 (Node 22's bundled npm) while the doctor gate accepts npm 10-11 and both remediation strings point back at npm 10 and the crashing command. Rule: a change that shifts a runtime requirement must, in the same change, update every remediation string, version gate, risk-register entry, and rollout note that names the old requirement; a fix validated on one consumer class is not a fix for the preserved-devDependencies class.
