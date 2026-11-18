---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T11:39:25Z
identity: determinism-20261116T113925-jay-flowers
tier: draft
as_of: "unbound-force@474f5aee31d7e0bc4583635cccfeec6030277b33"
---

A bounded retry that consumes a structured failure envelope is a DE-005 chain, not single-call guidance. When a change adds retryable classifications and resume hints to a delegation envelope, the once-bounded re-issue that consumes them (branch on timeout + retryable, relay the resume fields, anchor on checkbox state when the hint is null, stop on second timeout / non-retryable / non-timeout) contains no judgment between the two calls — the retry's arguments are fully determined by the first call's structured result — so the retry must ship inside the intent-level tool in the same change. Verified on PR #1157: the per-task executor already held the original prompt, the assembled resume hint, and the retryable classification, yet the retry routing shipped as 16 lines of /uf.unleash command prose, and spec PTB-FR-003's sentence ("Command prose MUST relay files_modified and tasks_completed into the retry prompt") mandated the prose mechanism. Authoring-time smell: a spec delta that assigns a mechanical relay to command prose. The requirement should mandate the executor-owned behavior and preserve its substance (never blank retries; hint always relayed) by construction, with the retry provenance recorded on the result envelope.
