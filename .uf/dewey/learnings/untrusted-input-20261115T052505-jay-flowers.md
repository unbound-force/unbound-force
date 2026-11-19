---
tag: untrusted-input
author: jay-flowers
category: decision
created_at: 2026-11-15T05:25:05Z
identity: untrusted-input-20261115T052505-jay-flowers
tier: draft
---

Assessing feedback on instruction-prose injection gaps: the untrusted-input marker fix and the tool-layer extraction are complementary, not competing. The data-not-instructions marker is the correct current-increment fix — it restores the file's own consumption-point convention (markers are scoped per consumption point, not per data class) and satisfies Constitution Principle V cheaply. The deeper fix — extracting the classification that consumes the untrusted text into a deterministic tool, e.g. a commit_id plus structured finding-set comparison inside resolve_review_mode or a sibling — subsumes the marker at that decision point and is the mechanism of record, but it belongs to the determinism owner (DE-001/DE-003) as a separately routed, larger change. Apply the marker now; never block it on the extraction decision.
