---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:38:38Z
identity: issue-triage-20261102T043838-jay-flowers
tier: draft
---

Agent panels that require strict enum taxonomies in their responses (e.g., triage category and objectivity enums) intermittently emit plausible-but-out-of-schema values — observed: "improvement" for category, subsystem labels like "release-pipeline" for category, "high" for objectivity. Consolidation that silently excludes such responses under-reports panel participation and discards provenance from otherwise-substantive assessments, biasing consensus toward whichever agents happened to phrase valid enums. The recovery pattern: validate against the documented enum set, issue exactly one corrective re-prompt carrying the validation failure and the permitted values while instructing the agent to retain its substantive reasoning, separate the accounting counts (contacted / responded / schema-valid / corrected / failed), preserve raw invalid responses and normalization details in artifact provenance, and never auto-map subsystem or domain labels onto taxonomy values (release-pipeline is not inherently "enhancement").
