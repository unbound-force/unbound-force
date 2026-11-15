---
tag: tool-extraction
author: jay-flowers
category: pattern
created_at: 2026-11-15T00:37:34Z
identity: tool-extraction-20261115T003734-jay-flowers
tier: draft
---

When extracting mechanical prose to a tool, align the tool's parameter vocabulary with the value available at the documented call site. select_prior_review's current_verdict enum uses the GitHub event types that /uf.review-pr's Step 8c mapping produces, but the tool is called at Step 8a — before that mapping runs — so the instruction's 'the current verdict' is ambiguous, and a verbatim council verdict ('REQUEST CHANGES', 'APPROVE WITH ADVISORIES') fails schema validation, routing to the error path that skips the very gate the tool serves. Either accept the caller's native vocabulary and map inside the tool (pure function beside the routing constants), or name the mapped value explicitly at the call site.
