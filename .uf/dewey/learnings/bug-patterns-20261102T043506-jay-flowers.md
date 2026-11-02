---
tag: bug-patterns
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:35:06Z
identity: bug-patterns-20261102T043506-jay-flowers
tier: draft
---

Cobra prints the full command usage block after any error returned from a command's RunE unless SilenceUsage is set to true on the root or executing command. Diagnostic commands that intentionally exit non-zero on environment check failures therefore imply an invocation error that never occurred. Verified in unbound-force issue #577: `uf doctor` (internal/doctor/doctor.go returns `fmt.Errorf("%d check(s) failed", ...)`, wired through RunE in cmd/unbound-force/main.go) printed "Error: 1 check(s) failed" followed by the doctor usage and flag list because the root command sets only Use/Short/Version — a repo-wide grep found no SilenceUsage/SilenceErrors anywhere. Fix pattern: set SilenceUsage: true at the root so every subcommand treats runtime failures uniformly; --help output is unaffected, but flag-parse errors also stop printing usage unless re-enabled via FlagErrorFunc. Regression test: drive the root command against a failing check and assert stderr contains no "Usage:" while the returned error stays non-nil.
