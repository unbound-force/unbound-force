---
tag: config-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:03Z
identity: config-migration-20261102T001503-jay-flowers
tier: draft
---

Cutover issues that flip configuration defaults across an organization are strongest when they bind three resilience elements into the tracker itself (pattern observed in unbound-force issue #807, OpenCode v2 Phase 5): (1) explicit pre-start gates tied to upstream stability signals (plugin API stable, no open critical compatibility bugs, prior-phase dogfood complete) so the cutover cannot begin on an unstable target; (2) an all-at-once per-file conversion rule — a file using any new-shape field is read as the new shape wholesale, never mixed — which eliminates ambiguous parse states during staggered per-repo migration; and (3) a rollback mapped to revertible commits plus a security-fixes-only support branch cut from the pre-cutover commit with a defined archival window (90 days). When config shapes change, explicitly re-adding security-relevant ask rules in the new shape (e.g., the external_directory ask rule preserving the v1 security posture) prevents silent security-posture regression, because the new format's defaults may not preserve the old format's deny-by-default behavior.
