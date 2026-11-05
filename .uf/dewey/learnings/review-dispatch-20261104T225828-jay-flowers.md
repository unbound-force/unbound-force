---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T22:58:28Z
identity: review-dispatch-20261104T225828-jay-flowers
tier: draft
---

Version skew between a producer's newly added capability and an older deployed copy of the producer is detectable at the consumer with a two-signal strand rule instead of silent misreporting. Pattern (unbound-force PR #949, RD-FR-056): the consumer compares the compact tool-result preview (produced by the loaded plugin, capped at 10 entries) against the persisted artifact's full section. Preview non-empty while the artifact section is empty means the loaded plugin predates durable proposals — the correct be
