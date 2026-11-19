---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T00:37:36Z
identity: test-quality-20261115T003736-jay-flowers
tier: draft
---

When a bug fix introduces a pure selector over an enum map (e.g., VERDICT_STATE_MAP with 3 verdict branches), test the immediate bug's verdict and also add parameterized fixtures for the other enum values and the empty-input edge case. In this PR, only REQUEST_CHANGES was exercised while APPROVE/COMMENT mappings had no fixture, leaving 2/3 of the contract unverified via line coverage.
