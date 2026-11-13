---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-11-10T18:50:17Z
identity: documentation-20261110T185017-jay-flowers
tier: draft
---

Manifest security-override bumps shipped through the scaffold mirror are user-facing for hero-repo adopters and require their own docs-labeled tracking issue — one per bump, even when a prior bump's issue is completed and closed. Verified on unbound-force/unbound-force: issue #977 ("Document the review-plugin security bump (zod 4.6.6, toml 5.0.1) for hero-repo adopters") was "Filed per the documentation gate for user-facing changes" because "uf init ships the updated manifest, and existing deployments pick the new pins up on their next uf init + install cycle"; the effect 4.0.0-beta.103 override (GHSA-chp3-j5qr-2vfr, PR #1024) had none despite following the same shape, with #977 already closed. A CHANGELOG entry does not satisfy this gate. Curator duplicate-check recipe: search the advisory's GHSA ID across all issue states (precedent issues cite GHSA IDs verbatim), search override terms ("effect override", "effect 4.0.0"), and confirm the newest open docs-labeled issue predates the change under review — also rule out the adjacent scopes that look like coverage but are not: the lockless-rollout docs issue (#964 / website#382) and scan-infrastructure follow-ups (#991) track the scanning posture, not per-advisory bumps.
