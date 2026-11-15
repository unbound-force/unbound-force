---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:00:07Z
identity: ci-cd-20261107T180007-jay-flowers
tier: draft
---

GOTOOLCHAIN=auto in a CI job unpins the Go toolchain that actions/setup-go derives from go.mod. Seen in unbound-force's ci_security.yml osv_plugin_transitive job (PR #957): setup-go v7 pins GOTOOLCHAIN=local (go 1.25.0 from go.mod), but osv-scanner v2.7.0 requires go >= 1.27.0, so the job exports GOTOOLCHAIN=auto and the effective toolchain becomes a property of the installed module rather than the repository. Integrity is preserved (module and toolchain downloads are verified against sum.golang.org when fetched through the default GOPROXY), but determinism and review visibility are not: the next scanner pin bump can silently change the toolchain with no reviewed diff to a toolchain line. Prefer an explicit GOTOOLCHAIN=go<version> pin (or a go.mod bump) so toolchain changes land as reviewable diffs; when auto is unavoidable, record the required-minimum rationale beside the pin and re-verify on every scanner bump.
