---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-06T19:40:33Z
identity: supply-chain-20261106T194033-jay-flowers
tier: draft
---

An exact-pinned transitive dependency survives a direct-dependency bump as a nested copy unless a package.json override accompanies the bump. In unbound-force PR #957 (stop-shipping-plugin-lock), bump zod 4.1.8->4.6.6 as a direct dependency did not touch the nested `zod: 4.1.8` that @opencode-ai/plugin@1.19.0 declares exactly; only adding a `zod: 4.6.6` override forced the nested copy onto the fixed version. The throwaway-lock OSV scan caught the surviving vulnerable copy precisely because the CVE (GHSA-mrhx-c7xc-m3w5) was known. When a security bump targets a package that any dependency pins exactly, always pair the direct-pin bump with an override and verify no nested copy remains in the resolved tree (the throwaway-lock scan or `npm ls <pkg>` shows nested duplicates).
