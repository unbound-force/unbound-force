---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:38:43Z
identity: issue-triage-20261102T043843-jay-flowers
tier: draft
---

Divisor panel agents intermittently return plausible but out-of-schema enum values in triage assessments — e.g. category "improvement" or a domain label like "release-pipeline" instead of the documented taxonomy, or objectivity "high" instead of objective/subjective. Fail-closed validation that silently excludes such responses discards substantive assessments and undercounts the panel (gaze#249 triage, unbound-force#581: five agents ran, three counted in the consensus). Recovery design that preserves a strict final artifact schema without losing data: (1) one corrective re-prompt to the same agent carrying the validation failure, the permitted values, and an instruction to retain the substantive reasoning; (2) record the assessment as failed only when the correction attempt also fails; (3) separate summary counts — available, contacted, responded, schema_valid, corrected, failed — so "responded" is never conflated with "schema-valid"; (4) persist raw invalid responses and normalization details as artifact provenance; (5) never auto-map a domain or subsystem label to a taxonomy value. Count fields must be derived from the dispatch roster at run time, not hard-coded, so panel-size changes do not drift the accounting.
