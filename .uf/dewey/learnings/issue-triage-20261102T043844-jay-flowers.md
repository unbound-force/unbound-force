---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:38:44Z
identity: issue-triage-20261102T043844-jay-flowers
tier: draft
---

LLM panel-agent output contracts fail at the enum boundary, not the extraction boundary: plausible near-synonyms ("improvement" for enhancement, "release-pipeline" as a domain label, "high" for an objective/subjective field) pass length-based field validation and only collide with enum-constrained artifact schemas, forcing a silent drop-or-invalid-artifact choice that loses substantive assessments and provenance (unbound-force#581, gaze#249 triage: five agents ran, three counted). Recovery pattern: enforce enums at the extraction layer where values are still attributable to a run, re-prompt the same agent once with the validation failure and the permitted values while instructing it to retain its substantive reasoning, record failed corrections without blocking consolidation from remaining assessments, and separate the summary counts (responded / schema-valid / corrected / failed) so partial validity is representable in the artifact. Shipped precedent for the re-prompt surface: redispatch_failed_runs already recovers failed dispatch reviewers through persisted prompts; schema-invalid responses from successful runs are the analogous case.
