---
tag: cli-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-08T19:49:39Z
identity: cli-tooling-20261108T194939-jay-flowers
tier: draft
---

A CLI wrapper that pre-checks dependencies and runtime capability to deliver a clean exit-code contract (exit 2 = could-not-run with a remediation message; exit 1 = findings) is load-bearing resilience surface, not convenience. Instruction text that invokes the runtime directly (`node lib/tool.ts`) re-introduces every crash class the wrapper eliminates — missing node_modules, missing packages, sub-capability runtime — each exiting with the code the contract assigns to findings and dropping the remediation message, so a fresh checkout (gitignored node_modules) fails ambiguously on first run. Before writing a tool invocation into a persona, command, or skill, check `.uf/scripts/` and the Makefile for an existing wrapper and route through it; a node-direct allow is justified only by the recorded absence of a wrapper (as for tool-inventory.ts). Frontmatter bash allows and prose commands form one contract — swap both in the same change — and guard tests must glob the persona class rather than pin one exemplar file, or the next persona ships divergent silently (PR 959: divisor-determinism.md line 90 went node-direct in the same PR that hardened and guarded divisor-linguist.md's wrapper-only doctrine).
