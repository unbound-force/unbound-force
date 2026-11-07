---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:10:27Z
identity: supply-chain-20261107T161027-jay-flowers
tier: draft
---

A root dependency pin bump alone does not remediate a CVE when another direct dependency pins its own nested exact copy of the vulnerable package. In unbound-force PR #957, bumping the root zod pin to 4.6.6 (GHSA-mrhx-c7xc-m3w5) would have left a nested vulnerable zod 4.1.8 in place, because @opencode-ai/plugin 1.19.0 declares its own exact zod 4.1.8 dependency; the fix required a matching npm `overrides` entry ("zod": "4.6.6") to force the nested copy onto the fixed version. When remediating a transitive CVE, always check whether the vulnerable package appears as a nested exact pin under a direct dependency, and mirror the fixed version in overrides.
