---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:38:37Z
identity: issue-triage-20261102T043837-jay-flowers
tier: draft
---

Panel-output enum validation that silently excludes invalid responses skews consensus and loses provenance: /uf.triage-issue's read_assessment_fields marks out-of-taxonomy assessment fields (e.g. category "improvement", objectivity "high") unavailable per persona, so substantive assessments drop from consolidation and the artifact undercounts consulted agents (observed 3 of 5 in the gaze#249 triage; unbound-force#581). Resilient panel design needs a recovery tier between schema-valid and failed — one corrective re-prompt carrying the validation failure and the permitted values, raw invalid responses preserved in provenance, and summary counts separating contacted/responded/schema-valid/corrected/failed — while never silently mapping domain labels (e.g. "release-pipeline") onto taxonomy values. Structural caveat: the category/objectivity enums are restated in the command response-contract table, the assessment-field tool validation, and schemas/issue-triage (CI-tested negative-enum samples), so any recovery feature must keep the permitted-value lists single-sourced or mechanically synced.
