---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-06T19:40:35Z
identity: ci-cd-20261106T194035-jay-flowers
tier: draft
---

When a CI security gate installs a Go tool whose go directive exceeds the toolchain setup-go pinned (setup-go v7 sets GOTOOLCHAIN=local from go.mod), `export GOTOOLCHAIN=auto` scoped to that single step lets `go install` fetch the required toolchain through the Go module proxy with checksum-database verification. Content integrity is preserved (Constitution V satisfied via the package manager's built-in verification), the fetched toolchain is deterministic given the pinned tool version, and the gate fails closed if the proxy is unreachable. In unbound-force PR #957 this unblocked osv-scanner v2.7.0 (requires go >= 1.27.0) under a repository pinned to go 1.25.0, after the job initially starved with setup-go's local toolchain.
