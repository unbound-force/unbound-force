---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:40:50Z
identity: content-pipeline-20261107T184050-jay-flowers
tier: draft
---

When a change ships a copyable CI recipe as a documentation page in the same commit series that adds a drift test for the recipe's pins, the doc becomes an unguarded third copy of the coupled constants. On stop-shipping-plugin-lock (PR #957), TestPluginTestNpmPin_TwinsCISecurityPin (internal/scaffold/scaffold_test.go) sweeps the Makefile PLUGIN_TEST_NPM pin against every npx -y npm@ occurrence in .github/workflows/ci_security.yml, but docs/osv-transitive-scan.md — the page the CHANGELOG rollout directs hero-repo maintainers to copy — embeds npx -y npm@11.6.2, GOTOOLCHAIN=go1.27.0, and osv-scanner@v2.7.0 outside the sweep, so a one-sided bump passes the twin test while publishing a stale recipe. Rule: when adding a drift test for coupled pin sites, enumerate every surface that embeds the pins — including documentation pages meant for copy-paste, whose regex sweep is usually a one-file extension of the existing test — or reduce doc-embedded pins to pointers at the canonical site.
