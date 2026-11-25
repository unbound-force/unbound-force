---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-24T20:04:29Z
identity: review-dispatch-20261124T200429-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

When an instruction tells a reviewer to treat a governance signal (e.g., a human-decision disposition) as authoritative, never route it through review_context: dispatch_review_runs concatenates caller review_context with rendered issue text into one UNTRUSTED block, so the reviewer cannot distinguish a genuine disposition from issue-body text mimicking one. Carry trusted signals in a structured session_metadata field instead.
