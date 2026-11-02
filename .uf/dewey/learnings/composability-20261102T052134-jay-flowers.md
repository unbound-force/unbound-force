---
tag: composability
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:21:34Z
identity: composability-20261102T052134-jay-flowers
tier: draft
---

Unbound Force composability claims (Constitution Principle II) operate at two distinct coupling layers that triage and review must assess separately. Layer 1, runtime-call coupling: absent by design — personas degrade gracefully when peers are missing (LookPath guards for sub-tools, warning-not-failure sub-tool init per Principle II, optional language packs, Grep fallback when Dewey is absent). Layer 2, shared-contract coupling: intentional and load-bearing — convention packs (e.g., severity.md with its MUST-load directive for review personas) are dual-consumed by Cobalt-Crush for writing and Divisor personas for reviewing, and alwaysDeployedPacks in internal/scaffold/scaffold.go co-deploys the core packs so writer and reviewer share one rule set. Because reviews are measured against the same rules the code was written to, "independently installable" must be scoped to runtime independence, not zero coupling. When triaging installability claims, verify the uf init flag surface directly at issue-handling time — it evolves (e.g., --divisor persona carve-out; a --packs-only flag appears in the multi-agent-init proposal) — rather than inferring install granularity from runtime degradation behavior.
