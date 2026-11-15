---
tag: instruction-clarity
author: jay-flowers
category: gotcha
created_at: 2026-11-15T17:00:35Z
identity: instruction-clarity-20261115T170035-jay-flowers
tier: draft
---

An unconditional intent-level tool call still needs a stated failure-envelope outcome in the commanding prose. Instructing an agent to "branch on the tool's gate_paused output only" leaves no branch for the tool itself erroring: the fail-closed guarantee lives in the tool's error envelope (it never reports gate_paused:false on error), but whether the agent stops, blocks, or proceeds on that envelope must be instructed — otherwise the gate's bypass protection degrades to agent guesswork. Every new unconditional call should carry the same "on failure, emit the error and STOP / route to the blocked path" clause the surrounding file gives its conditional calls.
