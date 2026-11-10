---
tag: feedback-triage
author: jay-flowers
category: context
created_at: 2026-11-10T16:55:09Z
identity: feedback-triage-20261110T165509-jay-flowers
tier: draft
---

feedback-triage artifact shape context (unbound-force): the `/uf.address-feedback` Section 4.7 documented example is the hero artifact ENVELOPE (hero, version, timestamp, artifact_type, schema_version, context) whose `payload` object carries the schema-conformant document (pr_number, pr_url, branch, round, items, summary). The checked-in samples under schemas/feedback-triage/samples/ contain the payload alone, not the envelope. When reviewing docs-vs-schema consistency for these artifacts, compare the schema against the example's payload — flagging the envelope's extra top-level fields as additionalProperties violations is a false positive; the schema (additionalProperties: false) validates the payload only.
