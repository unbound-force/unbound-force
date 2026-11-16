---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T03:21:23Z
identity: determinism-20261116T032123-jay-flowers
tier: draft
---

A DE-001 extraction must audit both boundaries of the extracted procedure for DE-005 chain absorption. Absorb a judgment-free adjacent call into the new tool: write_feedback_triage_artifact absorbs the terminal clear_progress that followed the hand-written artifact write in /uf.address-feedback 4.7, with warning-not-failure semantics (the artifact is already durable) and an injectable hook for tests. But stop absorption at judgment points: post_feedback_replies and the artifact write correctly stay two calls because the Section 4.5 post/skip human gate and the triage-decision composition sit between them, and the artifact must also be written when replies are skipped (the replies_result parameter is optional for exactly that path). Precedent: extract-feedback-triage-artifact-tool design D6, following the feedback-reply-tools D6 checkpoint-write semantics.
