---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:41:00Z
identity: error-handling-20261107T184100-jay-flowers
tier: draft
---

Duplicated consumer-facing constant seams need executable parity guards regardless of what the constants gate. Precedent: the stop-shipping-plugin-lock feedback-fix PR added TestPluginTestNpmPin_TwinsCISecurityPin for the Makefile PLUGIN_TEST_NPM vs ci_security.yml npx-pin twins (pipeline-gating constants) while also updating the four remediation constants (internal/scaffold pluginRuntimeRemediation/pluginInstallRemediation vs internal/doctor reviewPluginRuntimeRemediation/reviewPluginInstallRemediation — consumer-recovery-path constants) in manual lockstep with no guard: grep found no test referencing the constants in either package, only InstallHint presence checks. Comment-coupled twins have no failing signal on a one-sided bump (the repo's own feedback-triage precedent), and drift on remediation strings re-ships inconsistent recovery guidance — the support-dead-end class the lockstep exists to prevent. Rule: when a change edits one member of a duplicated constant pair, guard the pair in the same change — either a source-regex parity test in the findProjectRoot/TestPluginTestNpmPin_TwinsCISecurityPin shape (constants are unexported in different internal packages, so cross-package import is not available) or extraction to a shared internal constants package; a documented design commitment records intent, not enforcement.
