---
tag: scaffold-deployment
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:59:01Z
identity: scaffold-deployment-20261101T215901-jay-flowers
tier: draft
---

Unbound Force repo-local governance packs (determinism.md, meta-prompt.md) carry header comments declaring "not a scaffold asset", and meta-prompt.md explicitly states it "ships no -custom.md companion" (project-specific rules are incorporated by revising the pack through a governance change, not a sidecar custom-rules file). Deploying such a pack to hero repos via uf init — as issue #889 requests for meta-prompt with a meta-prompt-custom.md initial-deployment template — is a posture change from repo-local governance to a scaffolded, tool-owned pack with a user-owned custom companion (the established *-custom.md pattern: seeded once, never refreshed, exempt from drift tests). The self-describing header travels with the deployed copy, so the deploying change must revise the pack's header comment in the same increment, or every hero repo's deployed pack contradicts the custom template sitting beside it. When triaging or reviewing governance-pack deployment issues, flag the header revision as in-scope mechanical work rather than a blocker; determinism.md is the likely next candidate for the same trap.
