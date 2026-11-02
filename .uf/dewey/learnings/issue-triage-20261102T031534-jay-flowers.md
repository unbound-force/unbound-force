---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:15:34Z
identity: issue-triage-20261102T031534-jay-flowers
tier: draft
---

Triage verification of partially-shipped tool-extraction issues pays for itself: grep each named prose consumer individually before accepting the issue's duplication claims. unbound-force issue #670 claimed the convention-pack loading ladder was duplicated in uf.review-pr.md and uf.review-council.md, but by triage time only uf.review-pr.md §3.9 still carried the ladder — the council path had already centralized pack loading into build_review_prompts (openspec design D4: directory discovery of .opencode/uf/packs/), leaving only documentation pointers in uf.review-council.md. When a sibling PR ships part of an issue's scope with a stronger design (#745 shipped detect_resume_point as uf_unleash_detect_resume with artifact-first authority), treat the maintainer's status comment and retitle as the operative scope — the original body's acceptance checkboxes go stale — and verify each named consumer individually, because the remaining work is often narrower than the issue states.
