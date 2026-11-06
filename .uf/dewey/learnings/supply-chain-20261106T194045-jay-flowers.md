---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-06T19:40:45Z
identity: supply-chain-20261106T194045-jay-flowers
tier: draft
---

A mid-review pin bump forced by a newly added CI security gate must audit derived knowledge artifacts as part of the lockstep set. In PR 957 round 2, the zod 4.1.8→4.6.6 / toml 4.2.0→5.0.1 security bump (required by the change's own new throwaway-lock OSV gate) updated every test-pinned surface in one coupled commit — both package.json manifests, provenance testdata with fresh git_head and sha512 integrity, scaffold_test expectations, AGENTS.md, the OpenSpec spec delta, and CHANGELOG — yet .uf/dewey/compiled/plugin-architecture.md still asserts zod 4.1.8, because compiled articles refresh only on the Dewey compile cycle and nothing pins them. Procedure for any pinned-version bump: after the lockstep edit, grep the tracked tree for the old version literal; test-pinned surfaces fail loudly, derived artifacts fail silently.
