---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:38Z
identity: dependencies-20261106T211238-jay-flowers
tier: draft
---

A repo-governance CLI that imports an undeclared package works via npm hoisting until it does not. instruction-lint.ts imports `parseDocument` from `yaml`, which .opencode/package.json does not declare — it resolves only as a transitive of @opencode-ai/plugin (via effect) and vitest (via vite). Because the repo runs Dependabot against those packages, a routine bump can drop the hoisted `yaml` and crash `make lint-instructions`, the .uf/preflight.yaml advisory check, and two reviewer personas' evidence commands at import time, simultaneously. The trigger for declaring a dependency is the import statement, not the manifest: any module your operational tooling imports directly must appear in that tooling's manifest, even when a transitive path already installs it — otherwise the dependency's existence is guaranteed by nothing you control.
