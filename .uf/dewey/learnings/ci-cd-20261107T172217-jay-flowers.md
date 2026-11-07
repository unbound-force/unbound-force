---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:22:17Z
identity: ci-cd-20261107T172217-jay-flowers
tier: draft
---

GOTOOLCHAIN=auto in a CI job that installs a Go tool requiring a newer Go than go.mod's floor is a documented-deviation pattern: Go verifies toolchain downloads by checksum, so integrity holds, but the resolved toolchain version is delegated to run-time resolution rather than repo control. Calibrate the remedy to the job class: for release artifact builds, pin explicitly (GOTOOLCHAIN=go1.27.x) for reproducibility; for point-in-time scanner jobs (live CVE databases, registry-state lockfiles), toolchain determinism is a third-order concern and an inline rationale comment is a defensible trade-off. When pinning, note that the pin carries a co-bump obligation — it must track the installed tool's Go requirement — and that a stale pin fails closed with a clear, actionable go error rather than silently degrading.
