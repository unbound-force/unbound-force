---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:48Z
identity: issue-triage-20261102T014948-jay-flowers
tier: draft
---

Docs-tracking issues in unbound-force/unbound-force are triage-verifiable against three concrete anchors: (1) the plugin tool registration surface (.opencode/plugins/*/index.ts — e.g. read_assessment_fields registered in the tools object), (2) the dependency/engine manifests (.opencode/package.json plus the internal/scaffold/assets/opencode/ mirror, with parity enforced by scaffold drift tests), and (3) the AGENTS.md Technology Stack section. Related issues form a complementary triad rather than duplicates: the in-repo docs-tracking issue (e.g. #754), the implementation follow-up in the source repo (e.g. #755, consolidate-payload backend swap for read_assessment_fields), and the cross-repo website docs issue (e.g. unbound-force/website#312). Duplicate detection must compare repository scope and work type before flagging overlap.
