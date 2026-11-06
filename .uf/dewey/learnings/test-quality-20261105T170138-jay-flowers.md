---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:01:38Z
identity: test-quality-20261105T170138-jay-flowers
tier: draft
---

When a timeout/budget clamp computes an effective limit, testing the computation is not enough: assert that the effective limit actually fires. For a clamped run, advance fake timers past the clamped window but before the requested timeout and assert the run ended with kind wall_clock, effective_ms equal to the clamped value, and budget_clamped true. Computing the clamp (pure function), registering the deadline (map write), and arming the timer (setTimeout argument) are three independent wiring points — a r
