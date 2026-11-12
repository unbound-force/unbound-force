---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:00:30Z
identity: dependency-management-20261107T180030-jay-flowers
tier: draft
---

A lockless `npm install --omit=dev` still resolves devDependencies while building the ideal tree — it only skips writing them to node_modules — and npm 10.9.9's Arborist crashes ("Cannot read properties of null (reading 'edgesOut')") building the lockless vitest 5 peer graph. Structural fix (unbound-force stop-shipping-plugin-lock design D2a): scaffolded manifests ship no devDependencies, uf init preserves a target's existing devDependencies on refresh (preservePluginDevDependencies re-inserts them sorted before overrides), uf doctor warns on npm 10 when the target manifest declares devDependencies, and repository-side installs pin npm 11 (PLUGIN_TEST_NPM=npx -y npm@11.6.2) because the repo's own dev manifest needs the full graph.
