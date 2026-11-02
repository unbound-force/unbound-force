---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:35:15Z
identity: issue-triage-20261102T043515-jay-flowers
tier: draft
---

Cobra RunE plus an unset root SilenceUsage reliably reproduces "usage printed after runtime failure": Execute() prints the error and UsageString whenever RunE returns a non-nil error and SilenceUsage is false. Fast triage verification: grep the repo for SilenceUsage (zero matches confirms the root-cause claim) and confirm the failing path returns its error through RunE (in unbound-force, internal/doctor/doctor.go returns "N check(s) failed" and newDoctorCmd uses RunE, while cmd/unbound-force/main.go:22-26 sets only Use/Short/Version on the root). Fix-design nuance for assessing verification plans: a root-level SilenceUsage:true also suppresses usage for flag-parse errors, so a complete regression plan asserts both boundaries — no "Usage:" in stderr on runtime failure, and "Usage:" preserved for --help and invalid flags — which discriminates a blanket root fix from setting cmd.SilenceUsage inside RunE. Existing test scaffolding (executing newDoctorCmd with SetOut/SetErr buffers over a tool-less t.TempDir that yields failing checks) makes such a regression test follow the repo's io.Writer-injection pattern and fail pre-fix, pass post-fix (TC-006).
