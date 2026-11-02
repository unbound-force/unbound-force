---
tag: cobra
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:35:04Z
identity: cobra-20261102T043504-jay-flowers
tier: draft
---

Cobra prints the full command usage after any RunE error unless SilenceUsage is set on the command or root, so runtime failures render as if the user mistyped the invocation (unbound-force#577: `uf doctor` exits 1 on failed environment checks, then prints "Error: N check(s) failed" followed by Usage). Verified in code: the root command in cmd/unbound-force/main.go set only Use/Short/Version with zero SilenceUsage occurrences repo-wide, while internal/doctor.Run returned the failure-count error through RunE to root.Execute(). Fix pitfall: setting SilenceUsage statically on the root also suppresses usage for unknown flags and in-RunE flag-value validation (e.g., doctor's --format check), violating the requirement that invalid syntax retain usage. Prefer setting cmd.SilenceUsage = true inside RunE after flag validation, with regression tests asserting runtime failures print no "Usage:" while --help and invalid-flag paths keep it.
