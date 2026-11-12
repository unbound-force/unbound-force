---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:30:26Z
identity: review-dispatch-20261110T203026-jay-flowers
tier: draft
---

An implementation mechanism that changes an observable field of a tool result is contract, not internals, and needs its own clause in the same spec delta: when a byte-budget enforcement adds a new emitted key (errors_truncated) or caps a previously verbatim field (compact finalization errors to 5 x 256 characters), record the cap and the key in the requirement text and extend the delta's amendment-header scope to name the superseded clause. RD-FR-058 (full-lesson-proposals) capped the compact result's errors so its 40,960-byte budget holds on error-heavy paths, but the delta's amendment scope said 'proposal clauses' only — once archived, the spec would describe an uncapped errors array and a key order without errors_truncated. Before archiving a delta, grep the implementation diff for newly emitted keys and changed bounds and mirror each one in the requirement text.
