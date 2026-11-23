---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T19:19:57Z
identity: determinism-20261115T191957-jay-flowers
tier: draft
---

When an extraction spec mandates that a tool bind two input lists by a key (e.g., feedback-triage items to post_feedback_replies outcomes by comment_id), the binding key must appear in BOTH field enumerations. The feedback-triage-artifact-tool spec's FTA-FR-001 enumerated the per-item result fields (decision, issue_number, issue_url) without comment_id while mandating comment_id key match — an implementer following the enumeration alone cannot perform the binding in-tool and falls back to agent-side pairing, regressing the extraction the change exists to make. The upstream ReplyItemResult already carried comment_id; the fix was one word in the enumeration. General rule: a tool contract that joins inputs needs the join key in every joined list's field enumeration, and input-only keys must be documented as stripped at render time when the output schema forbids them (additionalProperties: false).
