---
tag: dependency-security
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:43:30Z
identity: dependency-security-20261110T204330-jay-flowers
tier: draft
---

In the unbound-force gatekeeping lockstep for security-override pin changes (canonical manifest + scaffold mirror + provenance record + pinning tests + CHANGELOG entry + docs/osv-transitive-scan.md), the rollback-documentation condition is the one most easily dropped. PR #1031 (effect 4.0.0-beta.103 override, GHSA-chp3-j5qr-2vfr) documented the advisory, upstream git head, integrity hash, and validation evidence, but neither its CHANGELOG entry nor the docs rollback sentence named a revert target for the new override — the docs sentence still pointed at PR #957 (zod/toml only), while the PR #957 precedent (stop-shipping-plugin-lock) had written "Rollback: revert this commit" explicitly in its CHANGELOG entry. Review checklist for override PRs: when the override set grows, grep the changed docs for the existing rollback sentence and extend it (or add a per-override rollback clause to the CHANGELOG entry naming the PR/commit to revert and what it restores); the rollback condition is conjunctive with the rest of the Gatekeeping security exception, so a complete-looking lockstep without it is still incomplete.
