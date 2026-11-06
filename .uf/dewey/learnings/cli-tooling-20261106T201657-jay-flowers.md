---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:16:57Z
identity: cli-tooling-20261106T201657-jay-flowers
tier: draft
---

A lib module that runs under plain node (type stripping) resolves imports through node_modules exactly like bundler-run code, but nothing in package.json declares a transitive package it imports — the lockfile pins it only as someone else's dependency. PR 959's instruction-lint.ts imports `yaml` (locked transitively via effect/vite) while .opencode/package.json pins exactly two runtime dependencies; an upstream graph change would break the CLI, the preflight wrapper, and the persona evidence commands with ERR_MODULE_NOT_FOUND, caught only incidentally by a round-trip test. When a governance pin forbids declaring the dependency in the same increment, add an explicit resolution-guard test whose failure message names the undeclared package, so the breakage diagnoses itself.
