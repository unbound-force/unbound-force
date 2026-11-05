---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:14:53Z
identity: ci-cd-20261105T191453-jay-flowers
tier: draft
---

The same version pin across two enforcement surfaces (Makefile variable + CI workflow literal) needs a drift test parsing both files and asserting the pins match. Mutual keep-in-sync comments are prose coordination that drifts silently when one surface is bumped.
