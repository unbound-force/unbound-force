---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:46:32Z
identity: instruction-quality-20261108T184632-jay-flowers
tier: draft
---

Temp-file + --input defends a gh api mutation against shell injection but not against JSON breakage: payloads composed from untrusted text (reviewer finding summaries, comment bodies) still need JSON-safe composition at the payload layer (jq -n --arg, or equivalent escaping). Additionally, every API-mutation step added to command prose needs an explicit failure outcome — especially when a downstream step is spec-illegal until the mutation succeeds (e.g., a FILE reply that must link a just-created issue). An unstated failure path leaves the executing agent improvising around a spec MUST mid-protocol. Sound pattern: bounded retry (once), then stop-and-report without posting the dependent step; keep the cross-run cache note for idempotency, which is a separate concern from in-run failure.
