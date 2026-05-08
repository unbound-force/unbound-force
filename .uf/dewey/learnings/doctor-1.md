---
tag: doctor
category: pattern
created_at: 2026-05-08T18:21:33Z
identity: doctor-1
tier: draft
---

The doctor package's Options struct now has a `GOOS string` field for injectable platform detection, matching the setup package's existing pattern. This is intentionally lighter than the sandbox package's `Platform *PlatformConfig` struct because doctor only needs the OS string for branching, not architecture or UID mapping support. When adding platform-aware checks to doctor, always use `opts.goos()` rather than `runtime.GOOS` directly, as `runtime.GOOS` is a compile-time constant that cannot be overridden in tests. This lesson was already documented in learning/sandbox-4 but applies equally to the doctor package.
