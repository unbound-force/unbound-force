---
tag: bug-patterns
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:35:10Z
identity: bug-patterns-20261102T043510-jay-flowers
tier: draft
---

Cobra CLI bug pattern (verified during triage of unbound-force issue #577): when a command's RunE returns a runtime error and neither the root command nor the subcommand sets SilenceUsage, Cobra prints "Error: ..." followed by the full command usage — implying the user misinvoked the command when the failure is actually a runtime/diagnostic one (e.g., a failed environment check in `uf doctor`). Root cause confirmed in cmd/unbound-force/main.go: the root cobra.Command sets only Use/Short/Version, and doctor.Run's error propagates through RunE to root.Execute(). Fix subtlety worth checking at review time: setting SilenceUsage statically on the root command also suppresses usage for flag-parse errors; the pattern that suppresses usage only for runtime failures is setting cmd.SilenceUsage = true dynamically inside RunE (after flag parsing succeeds). A compliant regression test asserts stderr contains no "Usage:" on runtime failure while `--help` and invalid-flag paths retain usage/help output. All Unbound Force hero CLIs (uf, gaze, dewey, replicator) are Cobra-based — when a similar "usage printed after runtime failure" report appears, check the root command for the missing SilenceUsage and prefer the dynamic in-RunE form.
