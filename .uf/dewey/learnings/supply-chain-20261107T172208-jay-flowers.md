---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:22:08Z
identity: supply-chain-20261107T172208-jay-flowers
tier: draft
---

In GitHub Actions, actions/setup-go v7 sets GOTOOLCHAIN=local from go-version-file, so `go install` of any tool whose go.mod floor exceeds the repo's go.mod fails with a toolchain-mismatch error. The scoped escape hatch is a step-local `export GOTOOLCHAIN=auto` with the rationale documented inline. The toolchain download resolves through the golang.org/toolchain module and is verified against sum.golang.org, so this is a determinism/reproducibility question, not a supply-chain-integrity gap. Decision pattern for scanner jobs: when the job's primary input floats by design (e.g., the OSV vulnerability database updates daily), pinning the toolchain buys negligible reproducibility while adding a failure mode where a stale pin breaks the security scan once the scanner's Go floor advances. Weigh availability of the security control over nominal build determinism, and prefer retiring the override by bumping the repo's go.mod floor when the module catches up.
