---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:09:08Z
identity: dependency-management-20261106T180908-jay-flowers
tier: draft
---

npm install --omit=dev still resolves devDependencies — it builds the full ideal tree and only skips them when writing node_modules — so a scaffolded manifest that ships devDependencies breaks activation on npm versions whose Arborist crashes resolving the dev peer graph (npm 10.9.9 on the vitest 5 graph: "Cannot read properties of null (reading 'edgesOut')"). Ship scaffolded manifests runtime-only and re-insert a target's existing devDependencies on refresh so repository manifests stay byte-identical and skipped; run the repo's own test suite through a newer pinned npm (PLUGIN_TEST_NPM). Precedent: stop-shipping-plugin-lock design D2a, preservePluginDevDependencies in internal/scaffold/plugin_activation.go.
