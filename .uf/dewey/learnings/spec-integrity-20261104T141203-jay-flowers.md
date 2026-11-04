---
tag: spec-integrity
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:12:03Z
identity: spec-integrity-20261104T141203-jay-flowers
tier: draft
---

When a review loop replaces an algorithmic rule mid-change, the revision must supersede every artifact carrying the earlier rule — originating task text, test names, and scenario boundary wording — or mark them superseded in place. In unbound-force PR 949 (review-dispatch-consumer-bounds), the loop replaced the proposal-dedup rule (overlap coefficient >= 0.5 / 4-token floor → Jaccard >= 0.3 / 6-token floor / negation guard) but left the originating task stating the old rule while citing the requirement that carries the new one, a test named "fewer than four tokens", and a scenario GIVEN worded as overlap < 0.5 — which is not a sufficient keep-condition under Jaccard 0.3 (equal-size sets sharing 49% score 0.325 and merge). Audit pattern: after a review-loop rule change, grep the change directory for every carrier of the old rule's distinctive constants — threshold numbers, token floors, similarity-measure names — and supersede each one.
