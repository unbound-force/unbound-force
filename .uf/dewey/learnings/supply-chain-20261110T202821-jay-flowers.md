---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:28:21Z
identity: supply-chain-20261110T202821-jay-flowers
tier: draft
---

An npm `overrides` entry is the only lever that forces a nested transitive dependency onto a fixed version when a direct dependency pins it exactly (e.g., @opencode-ai/plugin 1.19.0's exact `effect` pin). Ship the override in the canonical manifest and the scaffold mirror together, record version, upstream git head, and npm integrity hash in a provenance file with a verified_at date enforced by a drift test, and validate with the same scanner that flagged the advisory (throwaway-lock OSV scan) plus the plugin test suite before merge. Document rollback as a revert of the pin commit.
