---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:50Z
identity: review-dispatch-20261102T025750-jay-flowers
tier: draft
---

Security-relevant deterministic classification must have a single implementation. In unbound-force, the path → category/security_sensitive/user_facing classification feeds review-dispatch tier routing (lightweight/standard/heavy) and reviewer selection. Duplicating that classifier across plugins (review-dispatch's change-profile computation vs. a re-implemented uf-workflow classify_paths) creates drift risk: the copy that under-classifies security-sensitive changes silently routes security-relevant diffs to lighter review with no audit trail — a weakened gate invisible to CI. The mitigation is the shared-primitives pattern the repo already uses (classifyBranch, matchTaskLine, sanitizeSlug): extract the classifier to .opencode/lib/ as one tested pure function and expose only thin wrappers, so every tier-routing decision traces to a single source of truth. A secondary security gain of tool-extraction itself: inline prose classification is re-reasoned by the LLM each run and can drift after context compression; a tool makes the security-sensitive determination deterministic and auditable (Constitution Principle VI / DE-001).
