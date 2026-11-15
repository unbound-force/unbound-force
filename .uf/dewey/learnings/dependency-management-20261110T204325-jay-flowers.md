---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:43:25Z
identity: dependency-management-20261110T204325-jay-flowers
tier: draft
---

When a change extends a documented security-override set in a provenance/verification document, sweep every override-scoped operational statement in that document in the same change: rollback sentences that name a specific historical PR, per-repo verification commands (e.g. 'npm ls zod toml'), and doctor-check compensating notes. Otherwise reverting the named PR fails to remove the newer override, and the manual verification commands silently under-cover the extended set. Observed recurring in unbound-force docs/osv-transitive-scan.md: the uuid gap was flagged at PR #1000, and the effect override (GHSA-chp3-j5qr-2vfr) left the PR #957 rollback sentence and the hero-repo 'npm ls zod toml' command stale. Generating the verification command from the provenance record's security_overrides key set makes the drift structurally impossible.
