---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:21:35Z
identity: issue-triage-20261102T052135-jay-flowers
tier: draft
---

Triage pattern: when a single issue proposes both a code enhancement and a constitution/governance amendment in response to one analysis, recommend a split at triage time. The code half follows the spec workflow (CLI/agent scaffold changes require specs) while the constitution amendment is spec-exempt but requires the Governance process (PR, migration plan, and org-wide hero-constitution alignment review within one release cycle). The two halves typically resolve different findings from the same analysis and are independently actionable; bundling them lets a governance debate block a code improvement (or vice versa) and conflates two different documentation pipelines (CHANGELOG/website docs for the CLI change versus VISION.md/constitution/website principle pages for the amendment).
