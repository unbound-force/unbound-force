---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:16:17Z
identity: issue-triage-20261101T221617-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force repos enumerate per-file dispositions (CHANGELOG.md, AGENTS.md, docs/usage.md, docs/architecture.md, README.md) that map one-to-one onto verifiable repository state, so triage can convert claimed dispositions into confirmed ones with fast greps: search CHANGELOG.md and docs/ for the change name (entries and table rows), and grep distinctive guardrail text with an include pattern covering both the command file and its scaffold mirror — identical line numbers across the pair is a quick byte-parity signal. Gotcha: glob patterns rooted at dot-directories (e.g. .opencode/commands/*.md) can return "No files found" even when the file exists; use grep with an include pattern instead. Explicit no-change justifications (e.g., "README has no per-command reference table") close the disposition loop so no surface is silently skipped — precedents: unbound-force#747, unbound-force#856.
