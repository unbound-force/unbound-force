---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:20:13Z
identity: supply-chain-20261105T142013-jay-flowers
tier: draft
---

Lockless npm migrations must audit every install call site against the design's declared mitigation stack, not just the primary path. In unbound-force PR 957 (stop-shipping-plugin-lock), design.md D2 retained --ignore-scripts as the lockless-world lifecycle-script mitigation; the uf init activation path and the uf doctor remediation string carried it, but the Makefile plugin-test recipe — the change's widest transitive surface (full vitest 5 dev tree) — replaced lock-verified `npm ci` with a lockless install omitting the flag, and no design decision justified the divergence. Symptom pattern: per-path task specs (tasks.md 1.3 vs 1.4) that quietly differ in flag sets, each individually matching its task text while the implementation diverges from the design's posture. Check: when a change removes a lockfile or changes install posture, diff the flag sets of every command that previously invoked the verified-install mechanism (Makefiles, CI workflows, docs, remediation strings) against the design's declared mitigation stack, and require a recorded design decision for any path that diverges.
