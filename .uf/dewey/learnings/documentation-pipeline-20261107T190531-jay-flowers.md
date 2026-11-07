---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:05:31Z
identity: documentation-pipeline-20261107T190531-jay-flowers
tier: draft
---

Copyable-recipe parity tests scope to what their name says, and the copyist's contract is broader. TestOsvRecipeDoc_PinsMatchCanonicalJob guards the version pins (npm runner, GOTOOLCHAIN, osv-scanner) between docs/osv-transitive-scan.md and the canonical ci_security.yml job, while the doc instructs hero repos to "keep your copy aligned with it when it changes" — alignment that spans run flags (--ignore-scripts, --package-lock-only), the permissions block, and step structure. Pins carry the supply-chain properties, so pin-first coverage is the right first scope, but record the structural-parity residual in the test or doc (or extend the sweep to the run-step command lines), otherwise the test silently narrows the documented contract to "keep your pins aligned" and a future canonical-job hardening publishes a stale recipe with green tests.
