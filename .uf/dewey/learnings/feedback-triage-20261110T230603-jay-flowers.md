---
tag: feedback-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:06:03Z
identity: feedback-triage-20261110T230603-jay-flowers
tier: draft
---

feedback-triage schema-compat nuance verified during the feedback-reply-tools-docs review: the documented claim "a reader pinned to 1.0.0 rejects 1.1.0 artifacts" holds for every artifact this repo's producer emits, because the Phase 4.7 template always includes filed_issue_number/filed_issue_url (null when unused) and summary.filed/replied (0 when unused), and v1.0.0.schema.json sets additionalProperties:false on both items and summary — so those properties alone fail 1.0.0 validation regardless of the decision enum. A minimal 1.1.0 artifact that omits the optional new fields and uses only legacy decisions would still validate under 1.0.0; the decision-enum rejection (file/reply) triggers only when those decisions appear. When documenting reader incompatibility, distinguish "rejects the new decisions" (the README's precise wording) from "rejects the artifacts" (true only under the producer's always-emit convention).
