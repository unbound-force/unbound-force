---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:33:31Z
identity: content-pipeline-20261107T173331-jay-flowers
tier: draft
---

Repo-local CI compensating controls do not travel with scaffold deployment: `uf init` deploys plugin manifests and sources but not GitHub workflow jobs, so hero repos that follow a rollout step (e.g., dropping the tracked `.opencode/package-lock.json`) silently lose any CI-only control such as the `osv_plugin_transitive` throwaway-lock scan. Rollout guidance that only says "review the scanner-loss posture first" leaves maintainers with a caution but no remedy. The fix is a copyable recipe — the full job YAML including gotcha comments (the npm pin that is the twin of Makefile PLUGIN_TEST_NPM, and the GOTOOLCHAIN=auto override needed because setup-go v7 pins GOTOOLCHAIN=local while osv-scanner v2.7.0 requires go >= 1.27.0) — placed in durable documentation linked from the CHANGELOG, not YAML inlined in the changelog entry itself, which is an append-only record that drifts from the canonical workflow file. Before filing a new docs issue for such a gap, check the open docs-labeled issues: the PR's documentation-gate issue may already cover the posture and need only a scope extension with the recipe.
