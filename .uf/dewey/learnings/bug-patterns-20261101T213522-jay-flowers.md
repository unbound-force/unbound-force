---
tag: bug-patterns
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:35:22Z
identity: bug-patterns-20261101T213522-jay-flowers
tier: draft
---

gh CLI JSON field expectations baked into spec-time tool contracts drift from installed gh versions and break tools at call time. unbound-force issue #895: uf_finale_ci_poll failed with \"Unknown JSON field: conclusion\" because the finale-token-reduction spec contract defines failedChecks as {name, conclusion, url} and the plugin requests that field from gh pr checks --json, while the installed gh CLI (2.91.0) exposes bucket/state — which the host-side fetch_pr_checks tool parses correctly, leaving two divergent schemas for one external command in the same ecosystem. Same failure class as uf-fix-ci's nonexistent baseRepository field (see learning spec-review-20261029T060640-jay-flowers). Durable practices: verify every requested --json field against the installed gh at spec time; centralize the external-command schema across consumers (one shared field mapping or a version-tolerant parser with a deprecation path) instead of patching consumers in isolation; when the tool's output contract itself must change, thread the correction through spec FRs, command prose, and embedded scaffold mirrors, and fail closed on unrecognized values.
