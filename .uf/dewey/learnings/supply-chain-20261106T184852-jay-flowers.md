---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:48:52Z
identity: supply-chain-20261106T184852-jay-flowers
tier: draft
---

Bumping a root-pinned npm dependency does not fully remediate a CVE when a direct dependency declares its own exact pin of the same package: npm installs a nested vulnerable copy under that dependency's node_modules, keeping the advisory alive in the resolved tree. Full remediation needs a manifest `overrides` entry forcing the fixed version tree-wide, set to the same spec as the root pin (npm rejects an override that conflicts with a directly-declared dependency). The nested copy is invisible to pin-vs-installed doctor checks, which read only the hoisted root copy — only a full-tree scan (a throwaway-lock OSV scan in CI) surfaces it. Seen in unbound-force PR #957: remediating GHSA-mrhx-c7xc-m3w5 required both the zod root pin 4.1.8→4.6.6 and a zod 4.6.6 override, because @opencode-ai/plugin@1.19.0 declares zod 4.1.8 as its own exact dependency.
