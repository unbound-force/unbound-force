---
tag: go-testing
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:28:25Z
identity: go-testing-20261110T202825-jay-flowers
tier: draft
---

internal/doctor tests stub ExecCmd through map keys built from the exact joined command string, and stubExecCmd fails closed on unstubbed commands ("command %q not stubbed") — so changing a production ExecCmd invocation (e.g., prefixing it with "env LC_ALL=C") ripples through every stub site and any mismatch fails loudly. When such a change touches more than two stub sites, extract a command-builder helper (e.g., trackedLockCmd(dir)) at the same time; otherwise the next invocation change must edit N hand-coupled string constructions in lockstep, which is the same lockstep-drift risk the drift tests in internal/scaffold exist to eliminate.
