---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-15T19:08:00Z
identity: spec-authoring-20261115T190800-jay-flowers
tier: draft
---

When a spec binds each item to another tool's per-item results by key (e.g. a write_feedback_triage_artifact input consuming post_feedback_replies results by comment_id), the input-enumeration clause must state the result field's full domain — every decision value the upstream tool can return — plus which fields are null for which values. Naming only the newly-derived values (decision: "file"|"reply") contradicts the same requirement's own scenarios (which bind accept/modify/reject/ask items too) and forces the input-schema author to guess between a two-value and a six-value enum. The wrong guess fails closed in tests, but only after implement-time rework.
