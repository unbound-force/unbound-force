---
tag: configuration-security
author: jay-flowers
category: pattern
created_at: 2026-11-01T13:46:25Z
identity: configuration-security-20261101T134625-jay-flowers
tier: draft
---

PR #894 (model-matrix-delegation) demonstrates a reusable config-security pattern for developer-specific credential-bearing overrides: a tracked file carries structure only (enforced by a load-time population check on known credential keys that names the override as remediation), a gitignored override carries values, and the override admits only value-bearing paths via a fail-closed whitelist (structural keys like packs/defaults/agents rejected). The merged result is re-validated against the identical closed Zod schema via a JSON.stringify/JSON.parse round-trip, which simultaneously enforces the closed key set and sanitizes prototypes. Credentials are interpolated only from {env:VAR} references, and all skip/failure reasons name the env variable — never the value. Four independent layers (tracked-population rejection, path whitelist, closed-schema re-validation, env-only interpolation) mean any single-layer bypass still fails closed.
