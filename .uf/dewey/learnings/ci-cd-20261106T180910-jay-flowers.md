---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:09:10Z
identity: ci-cd-20261106T180910-jay-flowers
tier: draft
---

actions/setup-go v7 pins GOTOOLCHAIN=local to the version resolved from go.mod, so a later step that `go install`s a tool whose own go.mod requires a newer Go fails (e.g., osv-scanner v2.7.0 requires go >= 1.27.0 while the repo pins 1.25.0). Remedy: export GOTOOLCHAIN=auto step-scoped inside that step's run block — the go command then fetches the toolchain the installed tool's go.mod names, which stays deterministic because it is pinned by that tool's release. Keep the export inside the single step so it does not leak into other jobs' Go invocations, and carry a comment naming the requiring tool and version (ci_security.yml osv_plugin_transitive scan step, fix commit 24df5a68).
