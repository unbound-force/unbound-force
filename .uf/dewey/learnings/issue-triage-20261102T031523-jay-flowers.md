---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:15:23Z
identity: issue-triage-20261102T031523-jay-flowers
tier: draft
---

Status-split issue triage (verified on unbound-force/unbound-force#670, 2026-11-01): when a maintainer comment reports half an issue shipped under a stronger design (detect_resume_point → uf_unleash_detect_resume in #745, artifact-first resumability from .uf/artifacts/dispatch/ superseding the tasks.md comment-marker checks), verify the remaining scope against the live repository before accepting it — both the scope claim and the consumer list in the comment can be stale. Method that worked: repo-wide grep for the proposed tool name (list_convention_packs: zero hits in either plugin copy → remaining work is real) and for the prose it would replace. Verified consumer state diverged from both the issue body and the comment: the full pack-loading ladder (with embedded `test -d .opencode/uf/packs && echo "PACKS=yes"` bash) survives only in uf.review-pr.md:305-324; uf.review-council.md's ladder is already gone (only review-voice/severity references remain, superseded by build_review_prompts' dispatch-side pack loading), while unlisted lighter enumeration prose persists in uf.address-feedback.md:253, uf.cobalt-crush.md:50, and uf.agent-brief.md:266-289. The delegation path already renders pack directives tool-side (uf_delegate_command `pack` parameter, uf.unleash.md:1037-1040). Net architectural fact: dispatch-side loading covers dispatched review children, but direct command paths (uf.review-pr's own ladder, cobalt-crush, agent-brief) still re-reason pack enumeration from prose — so "close without a new tool" requires removing those ladders, not just citing build_review_prompts.
