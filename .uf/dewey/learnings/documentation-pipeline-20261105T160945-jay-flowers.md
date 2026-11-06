---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:09:45Z
identity: documentation-pipeline-20261105T160945-jay-flowers
tier: draft
---

When a change removes a detective security control and records the loss as an accepted risk, verify the risk outcome propagates to every user-facing documentation surface, not just the internal design doc: the CHANGELOG entry and its rollout instructions, and the documentation-gate issue scope. Verified on unbound-force PR 957 (stop-shipping-plugin-lock): design.md's Risks section documented the OSV-Scanner/Dependabot lock-based scanning loss, but the CHANGELOG rollout note ('existing hero repositories can drop their committed lock') and the change's docs-gate task 4.3 issue scope (install command, ignored lock, rollout note) both omitted the scanning consequence — so operators following the rollout instruction would silently lose transitive vulnerability coverage. Pattern: an accepted risk recorded only in spec-internal artifacts has no lifecycle after the change archives; the compensating control OR the explicit acceptance must also appear wherever affected users are told to act.
