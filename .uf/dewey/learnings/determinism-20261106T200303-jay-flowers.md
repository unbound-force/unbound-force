---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:03:03Z
identity: determinism-20261106T200303-jay-flowers
tier: draft
---

A `?=` Make variable and a workflow YAML literal that must carry the same pinned tool version form a twin pin that prose comments cannot keep in sync. Enforce the pairing mechanically — a test that reads both files and asserts the version literals match — matching the determinism pack's rule that mechanical invariants execute in tools rather than 'keep in sync' comments. Without the check, a one-sided bump silently diverges the CI-scanned dependency tree from the locally tested one.
