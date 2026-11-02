---
tag: issue-triage
author: jay-flowers
category: decision
created_at: 2026-11-02T04:10:06Z
identity: issue-triage-20261102T041006-jay-flowers
tier: draft
---

Triage boundary check for agent-guidance proposals: instruction text added to agent-facing files (AGENTS.md, commands, agents, skills) counts as mechanical procedural prose — and must be extracted to an executable tool per DE-001 — only when it prescribes a fixed procedure with deterministic output (parsing, existence checks, git choreography, enumeration loops, template rendering). Judgment-class guidance such as "scan the target test file for existing table-driven patterns covering the same assertion category and extend them rather than reimplementing" cannot be tool-executed (the extend-vs-standalone decision is semantic) and legitimately remains prose. When triaging such issues, verify the claimed guidance gap against both AGENTS.md sections and the loaded convention packs (go.md TC-006 mandates table-driven style but says nothing about extending existing tables), and re-verify cited code patterns locally — retro-generated issues carry stale line numbers (unbound-force issue #609 cited guardrailBlock at ~6605; the current tree places it at 10196 of an 11,305-line scaffold_test.go).
