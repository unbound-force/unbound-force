---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:40:44Z
identity: supply-chain-20261107T184044-jay-flowers
tier: draft
---

Copyable CI recipes in documentation are a third pin-twin surface. When a repo publishes a copyable job recipe (a docs page mirroring a canonical CI job) whose security properties depend on pinned versions — scanner version, package-manager runner, toolchain floor — the pin-parity drift test must cover the doc surface too, not just the Makefile-to-workflow pair. A doc that instructs copyists to "keep your copy aligned" while itself having no alignment enforcement reproduces exactly the one-sided-bump gap the twin test was created to close: a future pin bump updates the workflow, the doc silently serves a stale recipe, and copying repos inherit a stale scanner (narrowed advisory coverage) or a stale crash-workaround pin. Precedent: stop-shipping-plugin-lock follow-up (PR #957) — TestPluginTestNpmPin_TwinsCISecurityPin guards Makefile PLUGIN_TEST_NPM against ci_security.yml while docs/osv-transitive-scan.md carries the same npm/osv-scanner/GOTOOLCHAIN pins unguarded; the fix is extending the same findProjectRoot-shaped test to parse the doc's pins.
