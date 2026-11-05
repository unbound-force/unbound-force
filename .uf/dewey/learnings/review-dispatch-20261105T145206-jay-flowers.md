---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:52:06Z
identity: review-dispatch-20261105T145206-jay-flowers
tier: draft
---

In .opencode/test suites, a new describe block re-declaring fixture factories (provenance(), scratchRoot()) that a sibling describe already defines sits below default.md CS-004's ">2 locations" MUST threshold: classify it LOW (advisory maintainability), not a convention violation. It is still the leading edge of the suite's documented per-suite fixture accretion — after PR #794 the aggregate-dispatch fixtures existed in four diverged copies before remediation via helpers/aggregate-dispatch-fixtures.ts (learning review-dispatch-20261023T145629-jay-flowers). Fix shape that keeps the refactor semantically neutral: hoist one parameterized factory pair to the scope the describes already share, pass each call site's varying value (e.g., the provenance branch string) as the parameter, and drop purely cosmetic variations such as scratch-dir prefixes when uniqueness already comes from randomUUID(). Escalate to a cross-file helpers module only when a third suite needs the factories — that is when CS-004 actually trips.
