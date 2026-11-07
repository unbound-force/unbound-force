---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:07:28Z
identity: error-handling-20261107T160728-jay-flowers
tier: draft
---

A lint-wrapper that pre-checks dependency presence before exec'ing node guards the missing-install class of could-not-run, but not the runtime-compatibility class: a Node older than the TypeScript type-stripping era fails at module load on a .ts entry (ERR_UNKNOWN_FILE_EXTENSION) and exits 1 — the same crash-as-findings masquerade the wrapper was built to prevent. engines-field floors only warn (EBADENGINE), and `requires:`-style pre-flight probes check executable existence, not version, so no entry point enforces the floor. Exit-code contracts for node-based lint tools need either a version-floor probe in the wrapper (parse `node --version`, compare against the declared engines range, exit 2 with the required range in the remediation message) or an explicit entry in the design's accepted-residuals list alongside corrupted installs and node-binary absence — the sub-floor-node case is easy to overlook because it sits between the two residuals usually considered.
