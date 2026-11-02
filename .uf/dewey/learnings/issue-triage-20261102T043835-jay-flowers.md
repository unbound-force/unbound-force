---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:38:35Z
identity: issue-triage-20261102T043835-jay-flowers
tier: draft
---

LLM-agent output contracts with strict enums need bounded recovery, not silent exclusion. Incident (gaze#249, filed as unbound-force/unbound-force#581): 2 of 5 Divisor triage agents returned plausible but out-of-taxonomy values (divisor-architect category "improvement"; divisor-sre category "release-pipeline" with objectivity "high") and both substantive assessments were dropped from consolidation, which then recorded only three consulted agents. Fix pattern: validate against the documented enum; on failure re-prompt the same agent once with the validation error, the permitted values, and an instruction to retain the substantive reasoning; record the assessment as failed only if the correction also fails; keep separate counts (available, contacted, responded, schema-valid, corrected, failed); persist raw invalid responses and correction details as artifact provenance; and never heuristically map domain or subsystem labels onto taxonomy values (release-pipeline is not enhancement) — corrections must originate from the agent so the audit trail shows who chose the final value. Implementation lane in uf.triage-issue: assessment validation is tool-owned (read_assessment_fields; the command template prescribes no validation prose), so the re-prompt pass belongs in the dispatch/validation tooling with fixture tests covering valid, correctable, and irrecoverable responses; the corrective prompt's permitted-values list must derive from the canonical schema enum rather than restated prose (enumeration discipline — restated lists drift); and the command template's embedded scaffold mirror under internal/scaffold/assets/opencode/commands/ must be synced in the same change.
