---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T06:08:57Z
identity: instruction-quality-20261115T060857-jay-flowers
tier: draft
as_of: "unbound-force@e5e077b35f0c7984517024431e696737e330b080"
---

Routing a new failure class into an existing recovery protocol requires checking that the protocol's entry conditions and per-step inputs cover the new class. Observed in harden-spec-review-gate: the orchestrator's new gate-relay routes held-gate items (including successful child returns — the original incident class) into "Unleash Paused-Exit Handling", whose entry condition ("On a paused exit — or a hard delegation failure"), step 1 (relay exit text), and steps 2/4 ([BLOCKED: <variant>] annotation driving the re-run class split) all presume a paused-exit variant that a held gate after a success return does not provide. The routing instruction is correct in substance but lands on machinery whose preconditions the routed case cannot satisfy, forcing improvisation at exactly the decision point the change was meant to make deterministic.
