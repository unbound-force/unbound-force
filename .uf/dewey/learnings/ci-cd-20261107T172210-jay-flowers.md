---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:22:10Z
identity: ci-cd-20261107T172210-jay-flowers
tier: draft
---

In GitHub Actions jobs that `go install` a version-pinned tool whose go.mod floor exceeds the repo's module floor, GOTOOLCHAIN=auto resolves to the pinned tool's declared floor (e.g., osv-scanner@v2.7.0 requiring go >= 1.27.0 over a local go 1.25.0), not the latest Go release — the effective toolchain is already fixed by the tool pin and drifts only when that pin is deliberately bumped. An explicit GOTOOLCHAIN=<floor> pin (or a job-scoped setup-go go-version) therefore encodes the same result under repo control at near-zero cost: the only maintenance is bumping the pin alongside the tool pin, a missed bump fails loudly at the toolchain check, and the pattern matches the paired-pin keep-in-sync-comment convention already used for exact npm pins like npx -y npm@11.6.2. Auto buys no freshness (it selects the declared floor, not the latest patched toolchain) and costs reproducibility of the scan environment in security gates.
