---
tag: review-dispatch
author: jay-flowers
category: reference
created_at: 2026-11-15T07:25:30Z
identity: review-dispatch-20261115T072530-jay-flowers
tier: draft
---

The dispatch-advisor skill (.opencode/skills/dispatch-advisor/SKILL.md) is the prose contract owner for plan_review_dispatch and dispatch_review_runs consumption. Its Fail Closed section enumerates rejection classes generically ("Any malformed matrix, malformed reviewer manifest, unknown discovered agent, invalid explicit run, Curator run-limit violation, or configured/absolute limit failure returns INCONCLUSIVE before child dispatch"), and its Plan Contract lists entry fields (agent/decision, reason, source, sequence, tier/model/variant, validation errors) without enumerating advisory types. Consequence for spec reviews of review-dispatch changes: new planner validation errors or plan advisories that ride these existing generic mechanisms need no skill or command prose updates — a "no prose changes" non-goal can be verified by quoting these two sections — while changes introducing rejection classes, plan-output fields, or aggregate-warning kinds outside the enumerated shapes do require prose updates (grep .opencode/*.md for the pinned phrases; only the plugin source and its tests pin them today). Verified during the dispatch-tier-resolvability spec review: RD-FR-063's advisory, RD-FR-064's fail-closed validation, and RD-FR-065's warnings/message fold all rode existing mechanisms, so the non-goal held.
