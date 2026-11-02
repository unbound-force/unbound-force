---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:09Z
identity: issue-triage-20261102T001509-jay-flowers
tier: draft
---

Triage pattern for migration phase issues: when a phase issue bundles immediate cutover work with time-deferred post-window cleanup (e.g., archive a support branch after a 90-day window; remove legacy code paths once no consumers remain), recommend splitting the deferred items into a separately tracked issue whose acceptance criteria are the trigger conditions themselves (window elapsed, zero-consumer verification method named). Bundled deferred scope either forces the phase issue to stay open months after its primary work completes or risks orphaning the cleanup when the issue closes; the split keeps each trigger explicitly gated and trackable. Observed triaging unbound-force #807 (Phase 5 cutover of the OpenCode v2 migration): the issue is otherwise gate-checked and rollback-covered, but its "Post-window" bullet defers v1 server() removal behind conditions with no named verification mechanism.
