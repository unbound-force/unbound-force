---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:15:33Z
identity: issue-triage-20261102T031533-jay-flowers
tier: draft
---

Determinism-extraction issues (Constitution Principle VI / DE-001) that propose a new discovery or listing tool should be consumer-audited before implementation, because an existing tool may already cover the load path. unbound-force issue #670 originally proposed both list_convention_packs and detect_resume_point; after detect_resume_point shipped in #745 as uf_unleash_detect_resume, the owner's status comment narrowed the remaining scope to the command-level pack-loading ladders in uf.review-council.md and uf.review-pr.md and flagged that build_review_prompts already reads .opencode/uf/packs/ from disk when assembling reviewer prompts — so the issue may close without a new tool. Lesson: when triaging or implementing a prose-to-tool extraction, first enumerate the live consumers of the procedural prose and account for existing tool coverage of the same load path; extraction is warranted only where a consumer remains after that audit.
