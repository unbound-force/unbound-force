---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:29:23Z
identity: issue-triage-20261102T002923-jay-flowers
tier: draft
---

Phase-tracked child issues that open with an explicit "Umbrella: #N" line are scoped subdivisions of a migration program, not duplicates of the umbrella issue. Triage duplicate checks must treat the umbrella declaration as a scoping signal: the child tracks one phase (e.g., a dogfood bake period) with its own scope and done-when criteria, while the umbrella owns cross-phase orchestration and gate decisions. Flagging such a child as a duplicate of its parent fragments phase tracking and discards the on-child record the phase depends on (e.g., gate-criteria status recorded on the phase issue as its exit evidence).
