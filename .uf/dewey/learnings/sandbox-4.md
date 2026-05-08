---
tag: sandbox
category: gotcha
created_at: 2026-04-28T19:01:49Z
identity: sandbox-4
tier: draft
---

When adding new guards to the sandbox Start() function (like parsePodmanVersion or isRootlessPodman), ALL existing tests that call Start() will fail because their injected ExecCmd mocks don't handle the new podman commands. The fix pattern is: add cases for the new commands (podman --version, podman info) to either the shared testOpts() helper or each individual test's ExecCmd. Also, tests that exercise platform-specific logic (like macOS UID probe) should use the Platform *PlatformConfig injection field on Options rather than trying to override runtime.GOOS, which is a compile-time constant. Setting opts.Platform = &PlatformConfig{OS: "darwin", UIDMapSupported: false} allows macOS detection logic to be tested on Linux CI runners.
