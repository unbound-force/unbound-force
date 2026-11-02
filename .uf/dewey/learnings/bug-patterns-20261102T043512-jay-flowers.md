---
tag: bug-patterns
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:35:12Z
identity: bug-patterns-20261102T043512-jay-flowers
tier: draft
---

Cobra prints the full command usage block after any RunE runtime error unless SilenceUsage is set (default false). For diagnostic/health-check CLIs such as `uf doctor`, where exit code 1 signals environmental findings (e.g., "1 check(s) failed" from internal/doctor Run) rather than an invocation error, this default conflates environment failures with syntax errors and buries the actual failed check below usage text in long reports. Verified in unbound-force issue #577: cmd/unbound-force/main.go defines the root command without SilenceUsage, and internal/doctor/doctor.go returns "%d check(s) failed" as a RunE error, triggering usage output. The standard fix pattern: set SilenceUsage at RunE entry (after flag parsing) or on the root command, keeping usage for --help and flag-parse errors; regression tests should assert stderr lacks "Usage:" on the runtime-failure path while invalid-flag paths retain it.
