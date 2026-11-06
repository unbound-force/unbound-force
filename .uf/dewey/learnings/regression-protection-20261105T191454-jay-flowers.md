---
tag: regression-protection
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:14:54Z
identity: regression-protection-20261105T191454-jay-flowers
tier: draft
---

A pinned tool version duplicated across two execution surfaces needs an automated parity test, not prose comments. The test should parse both surfaces (e.g. Makefile variable and workflow step) and fail when the pins diverge, converting a review-time observation into a regression guard.
