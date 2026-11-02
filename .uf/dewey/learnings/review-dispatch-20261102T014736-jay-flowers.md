---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:47:36Z
identity: review-dispatch-20261102T014736-jay-flowers
tier: draft
---

Carrier migration does not move the trust boundary. When swapping a tool's structured-field source from parsed child prose to payload-carried fields (review-dispatch #755, the design D4 task-5.4 follow-up), the Constitution V untrusted-input validation must migrate with the data: child-agent output remains untrusted regardless of whether it arrives as persisted text or as payload fields. The implementation should validate at the boundary where child output enters the store (submission time) and either retain read-side bounds checking in read_assessment_fields or make write-side validation the single enforced gate — no path may carry unvalidated child text into the payload unbounded. The bounds contract to preserve verbatim: category non-empty at most 128 chars; objectivity at most 64; reasoning at most 4096; split recommendations at most 16 strings of at most 512 chars each; every value type-checked; invalid values skipped as unavailable for that run. The issue's explicit "keep the same bounds and unavailability semantics" commitment is the correct anchor; the spec for the follow-up should pin down which side (write or read) owns enforcement so the two do not silently diverge.
