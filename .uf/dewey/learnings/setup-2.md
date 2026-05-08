---
tag: setup
category: gotcha
created_at: 2026-05-08T18:21:38Z
identity: setup-2
tier: draft
---

When adding new steps to `uf setup`, the step count label pattern `[N/M]` is hardcoded as string literals in every `fmt.Fprintf` call. Adding 3 new steps (Podman, DevPod, DevPod provider) required updating 13 existing labels from `[N/13]` to `[N/16]` plus all existing tests that assert these labels. Additionally, existing integration tests that call `Run()` with injected `ExecCmd` need fallback cases for the new tool commands (podman, devpod, devpod provider list) to prevent cascading test failures. This is the same pattern documented in learning/sandbox-4 about new guards causing existing test failures.
