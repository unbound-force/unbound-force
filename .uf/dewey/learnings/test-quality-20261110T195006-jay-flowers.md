---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:50:06Z
identity: test-quality-20261110T195006-jay-flowers
tier: draft
---

Negative-assertion contract guards built on bespoke regexes need their own positive control. When a test's only assertion is "pattern not found" and every scanned file is clean, the regex's match branch never executes in any test run — a future regex edit silently disables the guard while CI stays green. This differs from strings.Contains guards, where the stdlib carries the correctness. Pin the detector with a table-driven regex test covering the historical offending forms and the intended exclusions, or plant a fixture in t.TempDir() and run the same scan helper against it. Observed in the triage-issue-heredoc review (#1011 / PR #1024): TestCommandContracts_RejectShellHeredocPayloads is the sole automated heredoc guard in the repo (instruction-lint has no heredoc rule), and its regex match path is dead code against the clean tree.
