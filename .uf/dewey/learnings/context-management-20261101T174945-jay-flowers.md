---
tag: context-management
author: jay-flowers
category: gotcha
created_at: 2026-11-01T17:49:45Z
identity: context-management-20261101T174945-jay-flowers
tier: draft
---

During context-limit compression churn in long agent pipelines, tool results narrated in un-injectable live-tail messages can be confabulated into compress summaries and then mistaken for durable truth — twice in one /uf.unleash run on quick/accept-variantless-parent-model (2026-11-01): a fabricated APPROVE consolidation (correlation 2a3556dd — no artifact, no dispatch store entry, db_search showed the UUID only in my own compress inputs) and a fabricated capture/make-check round (diff path fd606da3 — no file on disk; invented test counts). Guards that caught both: after capture_review_diff/dispatch_review_runs/consolidate_dispatch, verify the returned diff_path/artifact path EXISTS on disk (ls) before treating the result as real; treat opencode.db (db_search) as ground truth for whether a tool call actually happened (query the correlation_id/diff UUID globally); after any context churn, re-verify pivotal results against durable state (files, checkpoint records) rather than trusting compressed summaries. Also append a correction to any progress note that recorded a phantom result as fact.
