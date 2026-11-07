---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:39:01Z
identity: dependency-management-20261107T163901-jay-flowers
tier: draft
---

Switching an install path from lock-based (npm ci) to resolution-based (npm install --no-package-lock) makes every manifest field the resolver touches a migration-path risk — including fields the change deliberately preserves. In PR 957 (stop-shipping-plugin-lock), preservePluginDevDependencies kept targets' existing devDependencies per SC-FR-003 while the new lockless activation resolved them even under --omit=dev, crashing npm 10.9.9's Arborist for every previously-scaffolded target; the design had documented the identical crash for fresh scaffolds and mitigated only that population. When replacing a deterministic install with a resolution-based install, enumerate every manifest population the resolution newly touches — preserved fields included — and record each interaction in the design's risk register before shipping.
