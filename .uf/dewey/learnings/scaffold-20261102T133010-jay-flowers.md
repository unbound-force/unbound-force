---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-02T13:30:10Z
identity: scaffold-20261102T133010-jay-flowers
tier: draft
---

When adding an early-exit mode to a scaffold engine's Run(), design the mode's failure accounting together with its dispositions, diffing against the incumbent soft-failure precedent already in the same function. Case (init-override-stubs-gitignore-sync review, 2026-11-02): the new OverridesOnly early return prints its sub-tool dispositions via printSummary but unconditionally returns an empty Result and nil error, while the same Run() carries the D12 precedent — pluginResult sets Status "partial" and increments FailedSubTools for a failed activation while deliberately keeping the CLI exit code zero. The new mode's I/O failures (action "failed", err nil, detail rendered only on the summary line) therefore leave Status empty, FailedSubTools 0, and exit code 0 — machine-invisible to CI and scripting even though a human reading the summary sees them. Review technique: for every new Run() branch or early exit, enumerate how each disposition it can produce maps onto the Result fields (Status, FailedSubTools) and the error return, and require either coverage or an explicit documented decision — mirroring how D12 documented plugin activation's soft-failure contract. The companion wiring rule: when a safety ordering is expressed as call sequence ("seed X after Y"), check whether the code gates on Y's disposition; sequence alone does not enforce the chain.
