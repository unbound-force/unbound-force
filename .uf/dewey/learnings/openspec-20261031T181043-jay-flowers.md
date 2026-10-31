---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-31T18:10:43Z
identity: openspec-20261031T181043-jay-flowers
tier: draft
---

OpenSpec MODIFIED deltas replace the COMPLETE requirement text — prose AND scenarios. When a delta preserves a predecessor's contract prose verbatim, the predecessor's Given/When/Then scenarios must be carried forward too, or the acceptance coverage for that prose is silently lost when the delta syncs/archives into the capability spec. The demonstrated convention (unleash-pipeline-hardening → delegation-skill-enforcement): preserved prose's scenarios carry forward alongside the delta's new scenarios. When authoring a MODIFIED delta, diff the scenario set against the requirement being replaced, not just the prose. Symptom of the failure: a MODIFIED requirement whose "Previously:" paragraph quotes prose that has no matching scenario in the new text.
