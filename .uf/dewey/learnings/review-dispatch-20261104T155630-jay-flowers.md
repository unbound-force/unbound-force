---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:56:30Z
identity: review-dispatch-20261104T155630-jay-flowers
tier: draft
---

Selector-surface coverage enumeration: when a change ships a tool whose contract fans out over an enumerated set (section selectors, error codes, modes), enumerate every value the production code branches on and grep the new test suite for a success-path request of each; values with zero matches are the coverage gaps even when every spec-enumerated scenario passes. Applied to read_dispatch_artifact on PR #949 (review-dispatch-consumer-bounds): all three RD-FR-057 scenarios passed while two of five section selector values had no success-path coverage — advisories was never requested, and findings was requested only past the end, an assertion the equal-length runs array would also satisfy. The consumer contract (RD-FR-056) reads those exact sections live, so the gap was load-bearing. Extends the error-code enumeration pattern (learning/review-dispatch-20261031T025137-jay-flowers) from error discriminants to selector discriminants; empty-path assertions must not be counted as coverage of a selector's success path.
