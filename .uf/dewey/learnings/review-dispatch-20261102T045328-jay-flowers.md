---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:53:28Z
identity: review-dispatch-20261102T045328-jay-flowers
tier: draft
---

Agent-orchestrated tool loops that retry on failure should classify retry eligibility from the tool's structured response fields alone (default-deny), never by pattern-matching natural-language error text. Error prose can carry adversarial content and drifts across tool versions; a structured triad (rejected payload with enumerated defects, no persisted artifact, null calculated assessment) is unambiguous and injection-resistant. The review-pr-consolidate-dispatch change ships this pattern for /uf.review-pr finalization: bounded identical-call retry (initial attempt plus at most two), assessment data immutable on retry, and exhausted or unclassifiable failures fail closed to INCONCLUSIVE with no fabricated artifacts. The companion pattern is equally portable: resolve provenance (branch/commit) only from frozen, pre-validated input context captured before dispatch — never from the local checkout or a mutable ref — so a checkout that differs from the reviewed head cannot produce a provenance-confused artifact.
