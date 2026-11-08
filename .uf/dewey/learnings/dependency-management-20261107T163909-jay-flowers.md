---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:39:09Z
identity: dependency-management-20261107T163909-jay-flowers
tier: draft
---

When a change swaps an install or resolution mechanism (lock-based npm ci to lockless npm install --omit=dev --no-package-lock), audit every existing-consumer population against the new mechanism, not just fresh deployments: preserved manifest fields can reintroduce the exact failure the change fixed for fresh targets. In stop-shipping-plugin-lock (PR 957), the scaffolded manifest dropped devDependencies to avoid npm 10.9.9's Arborist edgesOut crash, yet the same change's SC-FR-003 preservation keeps the vitest 5 devDependencies pair in previously-scaffolded hero repos and the meta repo itself — so activation crashes on the runtime the doctor gate accepts (npm 10-11) and the remediation strings recommend, while the working escape (npx -y npm@11.6.2) is documented only in repo-internal surfaces (Makefile PLUGIN_TEST_NPM, ci_security.yml, AGENTS.md) that scaffold consumers never see. Remediation strings and risk registers are the consumer-visible contract: validate them against every accepted runtime x manifest-shape combination, and propagate escape hatches into every surface the affected population can read.
