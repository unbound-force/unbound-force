---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:30:31Z
identity: instruction-quality-20261110T203031-jay-flowers
tier: draft
---

A design doc's claim that neighboring files "carry no old-contract language and are not edited" is a claim to verify, not trust. When a diff changes a tool's output contract (for example, removing a per-field truncation from a preview array), grep the whole instruction load set — commands, agents, skills, packs — for the superseded wording, including files the design says are unaffected, and classify every remaining match. In the RD-FR-058 (full-lesson-proposals) review, the surviving "512 characters" mentions were the unrelated dispatch_review_runs error-summary cap (uf.review-council.md:288, uf.review-pr.md:512, uf.address-feedback.md:464, uf.triage-issue.md:274) and uf.unleash.md:1031's skill_absent_reason bound — not stale proposal-preview claims — so the change's consumer-prose coverage was complete.
