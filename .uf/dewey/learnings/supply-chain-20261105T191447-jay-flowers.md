---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:14:47Z
identity: supply-chain-20261105T191447-jay-flowers
tier: draft
---

Two compensating controls (pins-vs-installed doctor check + throwaway-lock CI scan) cover different dimensions and neither closes the override gap: neither asserts that security-motivated overrides (toml/uuid bumps) actually constrained the installed transitive versions. When deleting a lock, verify every drift-control that replaced it (direct pins, package.json overrides) has a post-install verification path.
