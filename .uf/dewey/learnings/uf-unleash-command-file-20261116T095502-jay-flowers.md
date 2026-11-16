---
tag: uf-unleash-command-file
author: jay-flowers
category: file-specific
created_at: 2026-11-16T09:55:02Z
identity: uf-unleash-command-file-20261116T095502-jay-flowers
tier: draft
---

File-specific notes for this checkout (2026-11-16): .opencode/commands/uf.unleash.md carries ~21.3k self tokens, about 2.6x its 8000-token IL-030 budget, a condition that predates the per-task-delegation-budgets change (20.5k at merge-base 9c42f8f7) — every change to it should minimize added prose and record the measured instruction-text delta in CHANGELOG per DE-004 rather than attempting an in-scope extraction. The cmd/unbound-force TestRunInit_FreshDir scaffold file-count pin ([84-88] vs 89 observed) is caused by a concurrent opencode host writing a gitignored .opencode/package-lock.json that the scaffold embed.FS picks up; the failure is environmental and tracks the lock file's presence, so remediation is removing the stray lock (never node_modules) and re-running, and the same race can trip make plugin-test's no-lock-shipped assertion. The phase-checkpoint budget line in uf.unleash.md is now tool-rendered via the checkpoint tool's phase_checkpoint_line (exported pure renderPhaseCheckpointLine in plugins/uf-workflow/index.ts, response-only — get_progress never carries it), and the Go content gate in internal/scaffold/scaffold_test.go pins that contract with stale-absence checks that fail if the old prose template or timeout.kind enumeration routing returns. Tags: opsx/per-task-delegation-budgets, 2026-11-16, file-specific.
