---
tag: untrusted-data
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:32:07Z
identity: untrusted-data-20261108T183207-jay-flowers
tier: draft
---

Untrusted-data framing in this repository's agent commands is per-ingestion-point, not per-command. Dispatched child reviewers get diff content wrapped in UNTRUSTED delimiters by the review-dispatch tooling; sibling evidence, feedback threads, and the feedback change signal each carry their own bounded-untrusted notes in command prose. When a new route deliberately bypasses an existing boundary — like the Step 3.12 light check in uf.review-pr.md, which dispatches no child agents and calls no dispatch tools — it silently loses the framing that boundary provided. PR 993 feedback item 4 caught exactly this: the light check read the raw compare diff, prior review bodies, and inline comments bare in the parent session while every other ingestion point in the same command family framed identical content as untrusted. Review technique: enumerate every external-content read in changed command prose and verify the route carrying it has its own data-only note; a route that bypasses a tool boundary is a signal the framing was lost, not that it is unnecessary (IQ-11).
