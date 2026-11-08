---
tag: regression-protection
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:04:54Z
identity: regression-protection-20261107T190454-jay-flowers
tier: draft
---

When a parity drift guard is deliberately scoped to a subset of the pins a copyable surface embeds, record the exclusion at the guard itself, not only in the copied artifact. Precedent: TestOsvRecipeDoc_PinsMatchCanonicalJob (internal/scaffold/scaffold_test.go) sweeps the npm-runner, GOTOOLCHAIN, and osv-scanner version pins in docs/osv-transitive-scan.md against ci_security.yml, but the recipe YAML also embeds three SHA-pinned action references outside the sweep. The exclusion is justified — the doc's gotchas instruct copyists to resolve their own trusted SHAs rather than trusting the page's copies — but that rationale lives in the doc while the test's own comment says 'Hero repos copy the recipe verbatim', telling slightly different stories about what actually gets copied. A one-line comment in the test naming the excluded pin classes and the reason keeps the guard's boundary auditable at the enforcement point, so a future extender knows the gap is conscious rather than overlooked. Cheaper than extending the sweep when the copied artifact already disclaims the value.
