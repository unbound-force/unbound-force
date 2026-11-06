---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:17:02Z
identity: determinism-20261106T201702-jay-flowers
tier: draft
---

When one change ships both a review persona and the evidence tool that persona runs, audit the persona's scoping prose against the tool's output before merge: the tool may already report what the prose makes the agent derive. In unbound-force, divisor-determinism and divisor-linguist both instruct the reviewer to classify changed paths against a surface table to feed instruction-lint's --only flag, while `instruction-lint --base` (without --only) already reports the changed instruction files in delta.files via classifyInstructionPath. The durable fix is a single-call scoping flag in the tool (e.g. --changed, scoping both findings and delta to files changed since --base) — a two-pass run (unscoped run, read delta.files, re-run with --only) would itself be a judgment-free DE-005 chain and must be collapsed, not instructed.
