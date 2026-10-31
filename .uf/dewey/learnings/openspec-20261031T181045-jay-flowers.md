---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-31T18:10:45Z
identity: openspec-20261031T181045-jay-flowers
tier: draft
---

OpenSpec MODIFIED requirement deltas replace the requirement wholesale, including its scenario set. When a change modifies a requirement that another still-active change also modified (common in repos like unbound-force/unbound-force with 160+ long-lived active changes and no synced capability spec yet), the delta must restate not just the normative prose but also the prior change's Given/When/Then scenarios — otherwise those scenarios silently disappear from the capability spec at archive time, weakening verifiability of fail-closed contracts. Verified on the model-matrix-delegation review (unleash capability): its MODIFIED "Command delegation tool" restatement carried the delegation-skill-enforcement prose verbatim but dropped 14 of 16 scenarios (empty-string/over-length/non-enum rejection, judgment-with-skill rejection, skill-root collision, script-path resolution, among others). Restate the complete scenario set in the delta, and record the archive-order dependency when stacking on a co-active change's requirement.
