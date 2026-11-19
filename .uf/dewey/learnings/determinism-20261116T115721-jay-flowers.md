---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T11:57:21Z
identity: determinism-20261116T115721-jay-flowers
tier: draft
as_of: "unbound-force@474f5aee31d7e0bc4583635cccfeec6030277b33"
---

A tool parameter the agent fills by selecting from earlier tool outputs is DE-001 mechanical prose in disguise. The tell is a bracketed fill rule inside a call template: it names a selection source ("the most recent per-task delegation envelope of THIS phase's own delegations"), enumerates omission conditions, and adds anti-reach-back guards — a fixed procedure over recorded state with zero judgment, and one that depends on in-context memory exactly where session-resume guards declare it unreliable. When the spec names the derivation source ("MUST be derived from resolveInheritedDeadline on the current session"), the consuming tool can call that source itself: derive the value inside the tool, delete the agent-supplied parameter, and the fill rule, the omission conditions, and the guards all disappear with it. Check the delegation executor first — if it already computes the value for its result envelope, the derivation exists in tool code and only the parameter plumbing is wrong. Watch for the spec-divergence symptom too: prose omission conditions keyed on envelope availability (worktree-spawned tasks, already-complete-on-entry) silently drop a mandated report where the spec's only condition is the absence of a budgeted ancestor.
