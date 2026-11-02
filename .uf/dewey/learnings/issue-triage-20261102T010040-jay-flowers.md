---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:00:40Z
identity: issue-triage-20261102T010040-jay-flowers
tier: draft
---

Quantitative figures in issues belonging to a long migration chain drift as the codebase evolves, and triage should cross-check them against the repository's current self-description rather than accepting them as ground truth. Two instances from the OpenCode v2 migration chain in unbound-force/unbound-force: (1) spike issue #798 requires verifying "16 agents load, 32 commands work" while AGENTS.md's structure listing states "17 active agents (model-matrix-classifier hidden)" and "Slash commands (33 files)" — an off-by-one on both counts that may be intentional (hidden classifier excluded; one special-cased command) or plain drift, and is not resolvable from triage context alone; (2) umbrella #797 itself records that predecessor #714's plugin pin figure ("1.4.10") is stale against the verified pin (1.19.0). Neither discrepancy blocks triage — the underlying intent (all agents and commands functioning under v2; a working plugin pin) is unambiguous — but the spike's findings should record the authoritative counts so the follow-up spec verifies against real numbers, and triage commentary should flag the delta so the author can confirm intent.
