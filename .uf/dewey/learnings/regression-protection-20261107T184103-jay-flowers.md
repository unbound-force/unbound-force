---
tag: regression-protection
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:41:03Z
identity: regression-protection-20261107T184103-jay-flowers
tier: draft
---

A user-facing guidance fix shipped as corrected string constants needs its content asserted on every surface it ships to, plus a parity guard when the constants are duplicated across packages. Precedent: PR #957 feedback fixes corrected four remediation constants (scaffold pluginRuntimeRemediation/pluginInstallRemediation and their byte-identical doctor twins) to carry the npx -y npm@11.6.2 escape hatch, but no test asserted the new content or the D2 twin lockstep — the pre-existing failure-path test pinned only the install-command fragment, which the old crashing-runtime guidance also contained, so reverting the fix kept the suite green. The same PR added TestPluginTestNpmPin_TwinsCISecurityPin as the drift mechanism for another comment-coupled constant pair, leaving the remediation seam as the unguarded twin. Rule: when a fix's deliverable IS the string (remediation text, install hints, rollout guidance), the regression test asserts the corrected fragment on each shipping surface and cross-checks duplicated copies (source-scraping parity test or a shared internal package); a spec scenario that pins remediation wording (e.g. SC-FR-004's escape-hatch clause) gets an acceptance test named after the scenario, like its sibling clauses.
