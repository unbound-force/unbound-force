---
tag: documentation-planning
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:24Z
identity: documentation-planning-20261102T013524-jay-flowers
tier: draft
---

Spec documentation-impact task lists anchored to recalled passages go silently stale when a change alters CLI flag semantics or file-ownership behavior — passage recall misses sibling surfaces that describe the same behavior. In the unbound-force meta repo (init-override-stubs-gitignore-sync spec review, 2026-11-01), Phase 6 named the four configuration.md "never seeded" passages and a flag-table row, but missed docs/cli-reference.md's --force row ("Overwrite all existing files", falsified by the new never-reset per-developer override class), the cli-reference ownership intro's two-class model, and docs/architecture.md's Managed files listing ("appended", stale once the managed block gained upgrade-in-place semantics). Review technique: at spec review time, grep docs/ for the changed behavior's vocabulary — flag names and ownership verbs ("appended", "overwritten", "skipped", "updated", "deployed", "seeded") — and verify every match is either covered by a docs task or verified unaffected; anchor doc tasks to surfaced lines, not recalled ones.
