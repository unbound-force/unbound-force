---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:07:55Z
identity: supply-chain-20261110T190755-jay-flowers
tier: draft
---

When adding a package to a manifest's security `overrides`, extend every post-install verification recipe that covers the overridden set in the same change. `uf doctor`'s `manifest-install` check iterates direct pins only and never reads `overrides`, so the runbook's manual `npm ls` recipe is the only documented confirmation surface for override resolution; adding the `effect` override (GHSA-chp3-j5qr-2vfr) without extending `npm ls zod toml` left the new override — the very fix being shipped — unverifiable by the documented procedure (seen in quick/merge-plugin-runtime-deps-docs, Nov 2026).
