---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:58:39Z
identity: issue-triage-20261102T025839-jay-flowers
tier: draft
---

When triaging or scoping documentation issues for new review-matrix configuration fields, the affected surface includes the canonical JSON Schema registry (schemas/review-matrix/: schema JSON, boundary samples, README version-history table) alongside docs/ narrative pages — not just the plugin's zod schema and the CHANGELOG entry. Verified while triaging unbound-force/unbound-force#676 (tier_caps docs): the plugin zod LimitsSchema accepts limits.tier_caps and the CHANGELOG documents it, but v3.schema.json omits the field entirely and docs/configuration.md's limits summary still reads "persona/run/timeout/cost" — the registry drifted silently past a shipped change. Cross-checking the issue's claimed defaults and behavior constants against the implementation (DEFAULT_TIER_CAPS, TIER_CAP_PRIORITY, the Math.max(cap, alwaysRequiredCount) floor enforcement) confirmed every claim verbatim and surfaced the omitted registry surface; triage verification against implementation constants is the mechanism that catches scope omissions in docs issues.
