---
tag: documentation-gates
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:19:45Z
identity: documentation-gates-20261101T221945-jay-flowers
tier: draft
---

Unbound Force documentation gate vs. content-pipeline significance thresholds: adding new plugin tools (MCP-style capability surface shipped to every uf init consumer via the scaffold mirror) fires the AGENTS.md documentation gate even when the divisor-curator blog/tutorial significance thresholds (new agent, new CLI command or subcommand, architectural migration, new hero capability, new user workflow) do not trip — the command workflow can be fully preserved while the capability surface grows. Decision rule from the extend-fetch-pr-review-context change (#861 origin, #873 docs tracker): file the repo documentation issue; decline blog and tutorial with explicit threshold evaluations; consolidate website tracking on existing on-point open issues found by search-first resolution (website#342/#310) rather than filing duplicates; and record the full docs-surface assessment as a follow-up comment on the documentation issue. The gate's user-facing test is capability surface (scaffold-mirror distribution), not workflow change.
