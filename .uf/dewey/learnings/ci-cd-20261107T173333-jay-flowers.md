---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:33:33Z
identity: ci-cd-20261107T173333-jay-flowers
tier: draft
---

A compensating CI control added in the changing repository does not propagate to scaffolded targets: scaffolds deploy manifests and sources, not workflow jobs. When a change removes a security-observability input (e.g., a tracked dependency lock) from every target and compensates with a repo-local CI job, the rollout guidance that directs targets through the migration MUST carry a copyable recipe — the exact commands plus a pointer to the canonical job, including its non-obvious pins (exact npm version for lock generation, GOTOOLCHAIN workarounds, scanner version) — otherwise directives like "review your scanner-loss posture first" are unactionable, and tool-driven nudges (doctor warnings pointing at `git rm --cached`) apply migration pressure the guidance has not paired with the replacement capability. Verified on PR #957 (stop-shipping-plugin-lock): the `osv_plugin_transitive` job exists only in the meta repo's ci_security.yml while hero repos follow a recipe-less rollout note.
