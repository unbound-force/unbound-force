---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:34:48Z
identity: ci-cd-20261105T183448-jay-flowers
tier: draft
---

A compensating vulnerability-scan job added to replace lock-based scanning should persist its findings (SARIF upload to the Security tab), not only gate via exit code. Job logs expire with the run, so the audit trail for the control that justified removing the lock disappears: gating proves the scan ran clean at CI time, but SARIF makes findings triageable and history queryable. Pair the scan step with an osv-scanner SARIF output format, security-events: write permission, and a SHA-pinned upload-sarif step.
