---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:05:20Z
identity: dependency-management-20261110T200520-jay-flowers
tier: draft
---

Deterministic enforcement is feasible for the OSV runbook's confirm-recipe staleness that prior learning supply-chain-overrides-20261110T190811 identified (the `npm ls` command in docs/osv-transitive-scan.md went stale after the effect and uuid overrides were added; doc surfaces have no drift test). Repo precedent already cross-checks pin surfaces this way: TestPluginTestNpmPin_TwinsCISecurityPin parses the Makefile PLUGIN_TEST_NPM pin and every `npx -y npm@` occurrence in ci_security.yml and fails on drift. The same pattern extends to the runbook — a scaffold test can parse the fenced `npm ls` command in docs/osv-transitive-scan.md and assert its package list equals the keys of .opencode/package.json's `overrides` block (currently effect, toml, uuid, zod), so adding a security override without extending the confirm recipe fails the build instead of silently shipping an unverifiable manual check.
