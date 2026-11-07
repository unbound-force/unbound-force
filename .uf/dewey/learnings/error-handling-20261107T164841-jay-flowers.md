---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:48:41Z
identity: error-handling-20261107T164841-jay-flowers
tier: draft
---

Remediation strings are shipped recovery paths. When a change introduces a runtime-constraint split — npm 10.9.9's Arborist crashes the lockless install for devDeps-carrying manifests while npm 11 works — the working remedy must be propagated into every consumer-facing failure surface in the same change: the Go remediation constants (plugin_activation.go), the doctor InstallHints (internal/doctor/checks.go), and the CHANGELOG rollout guidance. In PR 957 (stop-shipping-plugin-lock) the remedy (npx -y npm@11.6.2) lived only on maintainer surfaces (Makefile PLUGIN_TEST_NPM, ci_security.yml osv_plugin_transitive, AGENTS.md) that uf init never deploys, so a scaffolded repo with preserved devDependencies was told to reinstall npm 10 and rerun the identical crashing command while uf doctor blessed npm 10.9.x healthy (the gate accepts npm 10 through 11). Audit rule: whenever a runtime floor or install strategy changes, grep the shipped remediation surface for the remedy, and amend any spec scenario that pins remediation wording (SC-FR-004 offline-install-failure) in the same change so code and spec travel together. A remedy that lives only in maintainer tooling is unreachable from the shipped artifact.
