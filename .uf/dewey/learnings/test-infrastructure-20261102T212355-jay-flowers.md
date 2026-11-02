---
tag: test-infrastructure
author: jay-flowers
category: pattern
created_at: 2026-11-02T21:23:55Z
identity: test-infrastructure-20261102T212355-jay-flowers
tier: draft
---

When a library is built around a DI seam for testability (e.g. PreflightDeclaredDependencies with an injectable runProcess), the production adapters that sit AT the seam are the code least covered by the DI-based suite — every test injects a fake, so the adapter itself is only ever exercised by higher-layer tests that mock one level above it. In the preflight-declared change, runProcessNoShell (the production runProcess) contained the subtlest logic in the diff — a three-way error classification (timeout kill vs output-limit kill vs ordinary exit; the output-limit branch keyed on err.code === "ENOBUFS", which Node 24 and Bun 1.4 never report -- they use ERR_CHILD_PROCESS_STDIO_MAXBUFFER) added to fix a review finding — and repo-wide grep showed zero test coverage at any layer; unit tests injected the whole function away, golden tests mocked execFile above it, and round-trips only hit two of the three classes. Rule of thumb: for every injected dependency, ask "what test covers the production implementation of this seam?" If none does, either export the adapter or extract its classification logic into a pure helper and table-test it directly. Fixes made during review loops are especially prone to this gap — the fix lands in the adapter, the tests all sit on the fake side.
