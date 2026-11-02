---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:38:40Z
identity: issue-triage-20261102T043840-jay-flowers
tier: draft
---

Prose→tool migration seam gap (unbound-force issue #581): when /uf.triage-issue's assessment extraction moved into the read_assessment_fields tool, the tool's Zod schemas carried only the Constitution V length bounds (category z.string().min(1).max(128), objectivity z.string().max(64)) — the documented category/objectivity taxonomy enums were NOT enforced at the tool layer. Out-of-taxonomy values (category "improvement" or "release-pipeline", objectivity "high") pass tool validation and reach Phase 3.3's specificity hierarchy, the 4.2 label mapping, and the schema-validated artifact, none of which define a path for them. Lesson: when extracting a procedure to a tool (DE-001/DE-007 seam verification), audit the full prose contract, not just safety bounds — a tool that validates less than the prose it replaced leaves consuming decision points with undefined inputs. Recovery design per the issue: one corrective re-prompt carrying the validation failure and the permitted values (instructing the agent to retain its substantive reasoning), failure recorded only if the retry also fails, strict final schema preserved, separated counts (contacted/responded/schema_valid/corrected/failed), raw invalid responses kept in artifact provenance, and no silent normalization of subsystem/domain labels onto taxonomy values.
