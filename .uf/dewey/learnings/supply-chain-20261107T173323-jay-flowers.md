---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:33:23Z
identity: supply-chain-20261107T173323-jay-flowers
tier: draft
---

When a change removes a security control from an artifact that downstream repositories consume and compensates only in the origin repo's CI, the compensating control needs a propagation path: scaffold it, ship a copyable recipe, or explicitly decline with a documented trade-off. Rollout guidance phrased as "review your posture first" is unactionable — no completion criteria, no control shipped. When propagating a CI recipe, copy the tested job verbatim so its security properties survive (full 40-char SHA action pins, least-privilege permissions, --ignore-scripts, exact-pinned tool versions); a paraphrased recipe regresses to mutable tags. Precedent: stop-shipping-plugin-lock (PR #957) — the throwaway-lock OSV scan in the meta repo's ci_security.yml caught five HIGH advisories (zod, toml) in the scaffolded plugin graph, but hero repos following the CHANGELOG rollout dropped their committed lock with no recipe and lost transitive CVE scanning.
