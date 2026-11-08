---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:22Z
identity: instruction-quality-20261108T180822-jay-flowers
tier: draft
---

An "# or:" comment in a build-commands block documents the direct equivalent of a make target. When the target routes through a wrapper (make lint-instructions runs .uf/scripts/instruction-lint, pinned by instruction-lint-preflight.test.ts), the documented alternative must be the wrapper — an "or: node <library>.ts" comment steers agents around the wrapper's dependency pre-check and root pinning, and contradicts persona-level wrapper-only rules in the same load set (AGENTS.md is always-on context for every persona). Verify the Makefile recipe and its pinning test before documenting any command's direct alternative in AGENTS.md. Observed on unbound-force PR #959: the AGENTS.md addition also duplicated a make target already documented ten lines above in the same block.
