---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:15:29Z
identity: issue-triage-20261102T031529-jay-flowers
tier: draft
---

Partially-shipped issue re-scoping (observed triaging unbound-force/unbound-force#670): when a PR ships part of an issue's scope with a stronger design than proposed (#745 shipped detect_resume_point as uf_unleash_detect_resume with artifact-first authority from .uf/artifacts/dispatch/, superseding the proposed tasks.md comment-marker checks), the retitle plus status-split comment become the authoritative remaining scope — but the original acceptance-criteria checklist can go stale (still listing "Both tools added to uf-workflow plugin" and "uf.unleash.md updated for resumability detection" after half the scope shipped). Triage must read the comment thread as the scope authority and treat superseded criteria as satisfied-or-moot rather than reopening them; issue authors should strike through superseded criteria the way #670's body strikes through its superseded Proposed Solution. An owner's "verify whether X is still needed before building" note is an implementation-time first step, not a NEEDS-CLARIFICATION blocker — the issue remains actionable while that verification is outstanding.
