---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:26:08Z
identity: review-dispatch-20261110T212608-jay-flowers
tier: draft
---

Bounded tool-result previews that feed records a parent persists verbatim must cap by dropping whole entries with an overflow flag, never by truncating individual fields: a cut body silently loses the record's conclusion, and a per-entry truncated:true marker does not stop consumers from storing the damaged record. Keep full-length records in the durable artifact and point consumers there. PR #1004 fixed the PR #955 finding this way: consolidate_dispatch's result proposals array is now a capped preview of complete entries (at most 10, proposals_truncated marks drops) while the artifact's payload.proposals holds the full-length records that dewey_store_learning consumes.
