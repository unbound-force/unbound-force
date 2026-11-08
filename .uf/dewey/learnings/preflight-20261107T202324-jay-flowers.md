---
tag: preflight
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:24Z
identity: preflight-20261107T202324-jay-flowers
tier: draft
---

A declared pre-flight check's `paths` trigger scope must cover the check's operational surface, not just the surface it inspects: include the run target (the wrapper script the check executes) and the dependency manifests the wrapper validates before running. unbound-force's instruction-lint declaration (.uf/preflight.yaml) scopes paths to the lintable instruction files plus the linter entry — a PR that changes only the wrapper script or bumps/renames yaml/zod in .opencode/package.json touches none of them, so the advisory check is skipped in paths-scoped modes exactly when its own behavior or toolchain changed. Rule of thumb: paths = (what the check inspects) ∪ (what the check executes) ∪ (what its execution depends on). CI-side unit coverage of the wrapper mitigates but does not substitute — the declared check is the surface that runs it end to end in the working tree.
