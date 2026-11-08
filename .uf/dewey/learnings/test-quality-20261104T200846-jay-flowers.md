---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-04T20:08:46Z
identity: test-quality-20261104T200846-jay-flowers
tier: draft
---

When a change's invariant is "artifact X must never be created" and the same change gitignores X, every existing cleanliness check becomes blind to violations: the ignore line hides the artifact from git status, and CI never sees it. PR #957 (stop-shipping-plugin-lock) enforced its no-lock invariant on the uf init path via exact-command unit-test fakes, but the Makefile plugin-test path — the only place real npm runs — had no absence assertion, so a dropped --no-package-lock flag would write a lock that no   [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
