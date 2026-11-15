---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:26:13Z
identity: test-quality-20261110T212613-jay-flowers
tier: draft
---

When a truncation or cap flag is computed from disjunctive conditions — entry-count overflow OR per-entry length cuts — a single test input that triggers both at once cannot detect the removal of either trigger. Isolate each disjunct with a dedicated case: one input with the count inside the cap but an over-length entry, and one with excess count but short entries. Each flag trigger then has a test that fails when only that trigger is removed from the source.
