---
tag: determinism
author: jay-flowers
category: decision
created_at: 2026-11-16T11:04:16Z
identity: determinism-20261116T110416-jay-flowers
tier: draft
as_of: "unbound-force@1d0e5cef2b10545a779afe8aa2c925fb2ab22f9e"
---

DE-005 boundary adjudication (PR #1147, /uf.address-feedback 4.6→4.7): a second tool call is not a judgment-free chain when its arguments are primarily judgment products from an earlier phase (the triage records) with the prior tool's structured result as one optional input, and the second call also runs on a path where the first never executes (the 4.5 skip-posting branch still writes the artifact). DE-005's trigger requires each next call's arguments to be fully determined by the previous call's structured result; a sink call that aggregates human-gate decisions plus an optional prior result is an intent-level call in its own right. The judgment-free segment that did collapse: write→clear_progress moved inside write_feedback_triage_artifact (design D6, explicitly reasoned). Marker for future reviews: collapse when data flows result→arguments; keep separate when the later call is a workflow sink consuming gate decisions.
