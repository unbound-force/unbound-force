---
tag: regression-protection
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:40:58Z
identity: regression-protection-20261107T184058-jay-flowers
tier: draft
---

A parity drift test added in response to duplicated-pin feedback should enumerate every site the same change propagates the value into. The test's scope naturally fixes to the sites the feedback named (which existed when the feedback was written), while new copies created by the fix itself — consumer-facing remediation constants, copyable recipe docs, spec scenario text — go uncovered. Precedent: PR 957 feedback fixes added TestPluginTestNpmPin_TwinsCISecurityPin for the Makefile PLUGIN_TEST_NPM <-> ci_security.yml pair while the same batch embedded npm@11.6.2 into four scaffold/doctor remediation constants, docs/osv-transitive-scan.md (three occurrences), and the SC-FR-004 delta scenario, none of which any test reads. Rule: before landing a parity guard, grep the pinned value across the repo and cover — or consciously exclude with a recorded reason — every hit the change touched.
