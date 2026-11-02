---
tag: opencode-v2-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:30Z
identity: opencode-v2-migration-20261102T005730-jay-flowers
tier: draft
---

Major-version platform migrations can avoid long-lived migration branches by keeping the canonical branch on the old-version config shape — the new major normalizes it in memory without rewriting source files — while new-version entrypoints land incrementally through ordinary short-lived feature branches in dual-entry form (one source exporting both old- and new-style entrypoints, e.g. `{ ...Plugin.define({ id, setup }), server() }`). This keeps every old-version contributor and downstream scaffold consumer working during the dual period, makes switching binaries a per-person choice rather than an event, and reduces cutover rollback to reverting the config-conversion commits. The post-cutover escape hatch is a support branch for the old version, cut from the pre-cutover commit at cutover time (not before), time-boxed (90 days) and restricted to critical/security fixes. Prerequisite before setting the timeline: audit unpinned installer paths — `uf setup` installs latest opencode unpinned, so downstream repos receive the new major immediately and dual-entry compatibility plus scaffold dual-emission must land before any promotion of the new version. Verified from unbound-force#797 decisions D1/D2/D3; complements the per-sibling phase-issue pattern in learning/opencode-v2-migration-20261102T003021-jay-flowers.
