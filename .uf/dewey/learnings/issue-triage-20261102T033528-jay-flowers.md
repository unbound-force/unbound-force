---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:28Z
identity: issue-triage-20261102T033528-jay-flowers
tier: draft
---

Skill-vs-tool triage test for instruction-deduplication issues: when a proposal extracts duplicated agent-instruction prose, classify it as an intent-level tool (not a skill) when every step of the absorbed sequence — state fetch, existence checks, mapping tables, temp-file + API posting, cleanup — is deterministic once judgment products exist (composed review body, human gate decision); determinism pack DE-005 governs. Three design properties carry over: (1) making the human gate a required parameter (e.g., gate_confirmed) converts a prose-instructed MANDATORY GATE plus its session-resume guard into a structural gate that a session resumed after context compression must re-supply, so the gate survives prompt drift (DE-003: behavioral guards are interim mitigations, not the mechanism of record); (2) temp-file cleanup on all paths (success and failure) prevents orphaned files carrying review bodies; (3) fold narrowly scoped sibling lookup tools (e.g., a standalone map_verdict) into the intent-level tool as internal constants so the agent is not re-chained through lookup calls.
