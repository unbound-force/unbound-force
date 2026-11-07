---
tag: address-feedback
author: jay-flowers
category: decision
created_at: 2026-11-07T17:06:22Z
identity: address-feedback-20261107T170622-jay-flowers
tier: draft
---

Assessing bundled maintainer feedback (PR 957 item 4: "add a drift test ... and/or record dist.integrity for npm@11.6.2") — split the bundle by remedy, not by finding. The drift-test half was data-driven and accepted: grep verified no test references PLUGIN_TEST_NPM (Makefile:36) or the osv_plugin_transitive pin (ci_security.yml:75); the live precedent TestPluginPackageInputs_ApprovedMaintenanceProvenance (internal/scaffold/scaffold_test.go:1611) already reads multiple repo-root files via findProjectRoot and cross-checks pinned values, so the fix is one test in that shape; and no spec amendment is needed because SC-FR-003's provenance clauses cover only the four manifest pins, not the npm CLI runner. The dist.integrity half was delegated to author discretion inside accepted-risk class A7: npx resolves name+version at runtime, so a recorded sha512 is audit documentation, not verification, and every pin bump must re-resolve it or the record goes stale and misleads auditors (adoption guidance if taken: record via `npm view npm@11.6.2 dist.integrity` beside the pin together with the resolution command and the re-resolution-on-bump obligation). A single-source alternative — routing the CI step through a make target that uses the PLUGIN_TEST_NPM variable — eliminates the duplication outright but restructures the workflow; the drift test is the lower-friction, in-pattern remedy the repo's learnings prescribe.
