---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:23:58Z
identity: supply-chain-20261110T212358-jay-flowers
tier: draft
---

A green throwaway-lock OSV scan on a PR head is the completeness proof that an npm `overrides` entry actually remediates an advisory and introduces no new ones — use it as CI-bot corroboration in review instead of re-deriving the affected-package mapping from GHSA prose. Observed on unbound-force PR #1006: the advisory text names `@effect/platform`'s makeCookie() while the override targets `effect` (the monorepo's version-locked package, the exact transitive of @opencode-ai/plugin 1.19.0); the resolved throwaway lock contains only the overridden version, so the passing OSV-Scanner (plugin transitive) check confirms no vulnerable copy remains, nested or hoisted.
