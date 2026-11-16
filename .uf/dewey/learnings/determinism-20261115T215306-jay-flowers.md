---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T21:53:06Z
identity: determinism-20261115T215306-jay-flowers
tier: draft
---

When a spec introduces reference-based gate authorization — a caller-passed ref naming a machine-stamped id in a persisted record — it must also specify the id's tool-sourced handoff: the recording tool returns the stamped entry (decision_id) in its result, and orchestrator/relay paths surface the stamped id to the delegated session that must reference it. Omitting the output contract invites the exact mechanical prose the tool extraction was meant to remove: agents reading the record file to construct the reference themselves (DE-001-class regression at the new seam). Owner-lookup evidence pattern: when a review finds a genuinely new tool (no existing owner), check the consuming decision point's input path with the same rigor as the tool's write-side contract.
