---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:30:22Z
identity: review-dispatch-20261110T203022-jay-flowers
tier: draft
---

When a bounded tool-result array feeds a persistence path (for example consolidate_dispatch proposals stored via dewey_store_learning), enforce the result byte budget by dropping whole trailing entries and flagging the drop count (proposals_truncated), never by per-field character truncation. A per-field cut silently corrupts records a parent persists verbatim — issue #969 found 11 of 12 council learnings stored mid-word from a 512-character preview cut. Measure the budget with one shared byte-size idiom (compact JSON UTF-8 length) across every bounded result builder so all retrieval modes count the same bytes, and cover the whole-or-absent guarantee with an end-to-end harvest test through the real submit tool and the artifact readback.
