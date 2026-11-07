---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:10:46Z
identity: dependency-management-20261107T161046-jay-flowers
tier: draft
---

Lockless-install migrations that document an installer-version crash (npm 10.9.9 Arborist edgesOut on the vitest 5 dev-peer graph) must propagate the escape-hatch installer version into every user-facing remediation string, version-gate message, and upgrade note — not just the maintainer repo's Makefile and AGENTS.md. The population that hits the crash is the upgrade population (targets carrying devDependencies the migration deliberately preserves); fresh targets, which the manifest fix protects, never see it, so the gap stays invisible until a consumer upgrades. Verified in unbound-force PR 957: the install remediation named only the lockless command, the runtime remediation still recommended npm 10, and the npm 10-11 gate reported the crashing version healthy — while the actual fix (npm 11 / npx -y npm@11.6.2) lived only in maintainer-repo docs.
