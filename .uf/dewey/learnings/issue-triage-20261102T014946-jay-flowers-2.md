---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:46Z
identity: issue-triage-20261102T014946-jay-flowers-2
tier: draft
---

Docs-tracking issues that follow a merged user-facing change in unbound-force repos arrive in companion pairs: an in-repo issue (e.g. unbound-force/unbound-force#754) plus a cross-repo unbound-force/website issue (e.g. website#312) created minutes apart, each carrying near-identical surface summaries scoped to its own repo's documentation. They are companions serving the two halves of the documentation gate (in-repo docs issue + the constitution's Cross-Repo Documentation rule), not duplicates — duplicate searches scoped to one repo will miss the pair, and cross-repo searches will surface both. Triage corroboration shortcuts: AGENTS.md Technology Stack confirms version/engine-floor claims (e.g. @opencode-ai/plugin 1.19.0, Node >= 22.22.2 via ini@7.0.0), the cited plugin tool's own input schema confirms bounds claims (read_assessment_fields: 128/64/4096 chars, 16x512 split entries), and the spec delta's requirement headings confirm FR ranges.
