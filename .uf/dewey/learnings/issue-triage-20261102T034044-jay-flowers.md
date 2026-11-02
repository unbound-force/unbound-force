---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:40:44Z
identity: issue-triage-20261102T034044-jay-flowers
tier: draft
---

Enhancement issues that prescribe a tool-call sequence for codification into documentation rot when the tool surface later evolves to intent-level aggregates. Verified on unbound-force #663 (expand dispatch-advisor skill to full dispatch lifecycle, filed 2026-10-05): its step 4 prescribed `dispatch_agent_run` for dispatch-planned runs (per PR #686), but by current HEAD the skill's own Invocation Tool section mandates the `dispatch_review_runs` aggregate and explicitly forbids individual `dispatch_agent_run` calls — implementing the issue as written would write a contradiction into the very skill being expanded and violate determinism rule DE-005 (intent-level orchestration; `dispatch_review_runs` is a shipped precedent). Triage rule: when an issue codifies a tool chain into prose or skills, re-verify each prescribed step against the current tool surface AND the target skill's existing MUST rules before routing to implementation; when an aggregate tool has already collapsed part of the prescribed chain, the issue typically stays VALID (the underlying duplication problem usually still holds and has often grown) but its procedural steps must be reconciled to the current surface at spec time.
