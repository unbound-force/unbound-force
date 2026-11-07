---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:53:28Z
identity: supply-chain-20261105T195328-jay-flowers
tier: draft
---

A tracked lockfile provides two distinct controls — resolution constraint (which transitive versions install) and monitoring surface (what CVE scanners can see) — and it provides them to two distinct audiences: the source repository and every repo scaffolded from it. When removing a lock, audit both dimensions for both audiences separately. In stop-shipping-plugin-lock, the monitoring compensation (throwaway-lock OSV scan) was wired only into the meta repo's own CI, so scaffolded hero repositories inherited neither the lock nor the compensating scan.
