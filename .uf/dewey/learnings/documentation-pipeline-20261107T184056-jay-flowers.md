---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:40:56Z
identity: documentation-pipeline-20261107T184056-jay-flowers
tier: draft
---

Documentation-gate issues filed early in a PR are point-in-time snapshots: when later fix commits in the same PR ship the remedy content (a recipe page, an escape hatch), the issue bodies keep the pre-remedy framing and never reference the shipped artifact. On stop-shipping-plugin-lock (PR #957), issues #964 (unbound-force/unbound-force) and #382 (unbound-force/website) still tell hero-repo maintainers to "review the scanner-loss posture first" without pointing at docs/osv-transitive-scan.md, which the final fix commits shipped and linked from the CHANGELOG — so a docs implementer picking up either issue re-derives the weaker, remedy-less guidance the change's own learnings call unactionable. Rule: when a fix commit ships user-facing remedy content after the documentation-gate issues were filed, extend those issues' scope (comment or body edit linking the shipped artifact) in the same change; the scope extension is part of the remedy, not deferred follow-up work.
