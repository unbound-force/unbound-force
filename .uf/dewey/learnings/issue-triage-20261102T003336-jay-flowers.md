---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:33:36Z
identity: issue-triage-20261102T003336-jay-flowers
tier: draft
---

Triage duplicate-screening for org-wide migration umbrellas: OpenCode v2 umbrella #797 Phase 3 spawns four parallel sibling-repo trackers with near-identical titles (#802 dewey, #803 gaze, #804 replicator, #805 website — all "Phase 3 — <repo>: OpenCode v2 verification pass"). Keyword-based duplicate candidates must be resolved by comparing the target repository named in the body and the umbrella relationship: sibling phase trackers are parallel work items scoped to different repos, and the umbrella (#797) is the parent/tracker — none are duplicates of each other or of a phase tracker. Complements learning/opencode-v2-migration-20261102T003021-jay-flowers, which covers the operability pattern of these phase issues (same-committed-tree compatibility, upstream-bug ownership).
