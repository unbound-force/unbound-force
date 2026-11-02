---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:15:20Z
identity: issue-triage-20261102T021520-jay-flowers
tier: draft
---

Docs-gate issues that name specific tools, artifact paths, and a page-by-page scope table can be triaged mechanically instead of judgmentally. Checklist that converted the VALID call for unbound-force#739 into verifiable evidence: (1) each named tool (write_unleash_summary, uf_finale_pr_context, uf_unleash_detect_resume) matched its registered MCP tool description, including the persisted artifact path (.uf/artifacts/unleash-summary/<sanitized-branch>.json), the consumer, and the camelCase body shape; (2) grep of the target docs pages confirmed the described content is absent while the target sections (usage.md /uf.unleash section, architecture.md commands tree) exist; (3) the referenced change directory (openspec/changes/unleash-deterministic-tools/) exists and its spec requirements and design decisions match the issue's two deltas; (4) the no-change rows (docs/heroes.md single pipeline-level mention, AGENTS.md generic plugin listing, README no description change) were spot-checked against actual page content. Naming tools and paths makes documentation issues falsifiable — prefer that over vague "update the docs" requests when filing.
