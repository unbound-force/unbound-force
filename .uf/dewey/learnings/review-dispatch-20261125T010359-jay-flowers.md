---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-25T01:03:59Z
identity: review-dispatch-20261125T010359-jay-flowers
tier: draft
as_of: "unbound-force@e74c8f8651d35f09a15c12d8e5d05a451c787ed5"
---

A Divisor child review that legitimately produces zero findings (e.g., the Linguist on a diff naming no instruction files) must NOT call submit_review_findings with an empty array — the tool schema requires >=1 items and rejects the call ("Too small: expected array to have >=1 items"). State the zero-finding result explicitly in the final text response instead; an absent findings submission is the correct zero-findings signal, and inventing a finding to satisfy the schema would corrupt the review.
