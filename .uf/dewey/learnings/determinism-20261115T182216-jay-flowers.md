---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T18:22:16Z
identity: determinism-20261115T182216-jay-flowers
tier: draft
---

Extraction specs that consume another tool's structured result must name the source tool's actual result fields, not the target artifact's field names. In the feedback-triage-artifact-tool spec review (2026-11), FTA-FR-001 and design D1 named `filed_issue_number`/`filed_issue_url` as `post_feedback_replies` per-item result fields, but `ReplyItemResult` (.opencode/lib/feedback-replies.ts) carries `issue_number`/`issue_url`; the rename mapping lived only in the command prose the change deleted. Before approving a prose→tool migration's input contract, grep the source tool's result interface and verify every enumerated input field exists; specify any rename as tool-internal mapping.
