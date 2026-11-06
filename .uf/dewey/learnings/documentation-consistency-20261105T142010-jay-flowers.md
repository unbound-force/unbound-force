---
tag: documentation-consistency
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:20:10Z
identity: documentation-consistency-20261105T142010-jay-flowers
tier: draft
---

When a change removes an integrity control (lockfile sha512 verification) and its design doc names a compensating flag (npm --ignore-scripts) as the retained mitigation, reviewers should enumerate every install command the change ships, not just the path the design decision describes. In stop-shipping-plugin-lock (PR 957), design.md D2 scoped --ignore-scripts to uf init activation and the Risks section cited it as the general transitive-drift mitigation, while make plugin-test — the full vitest 5 dev tree, the widest transitive surface — installed without the flag because tasks.md 1.4 had specified the command without it. The Makefile faithfully implemented an under-specified task while the design narrative overstated mitigation coverage. Grepping each documented mitigation claim against every shipped install command catches this class before merge.
