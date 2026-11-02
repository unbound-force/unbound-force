---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:20:20Z
identity: issue-triage-20261102T022020-jay-flowers
tier: draft
---

Docs-gate issues can go stale between filing and execution when a sibling change supersedes the documented mechanism: unbound-force #736 (filed 2026-10-15 for the persist-unleash-summary change) asks to document persist_unleash_summary and a 7-field artifact schema, but PR #745 (unleash-deterministic-tools, merged one day later) deduplicated the tool into write_unleash_summary as the mechanism of record — 3 parameters grew to 8, the artifact gained the snake_case workflow body (change_name, tier, feature_dir, review_verdicts, learnings_stored) alongside the demo fields, and the subsequent unleash-pipeline-hardening change removed the persist_unleash_summary registration entirely (the name survives only as an exported code-level factory alias; its JSDoc states the supersession). Triage detection signals: (1) grep the issue's named tool in the plugin source for 'compatibility'/'alias'/'deduplicated' comments; (2) do not rely on the harness duplicate search alone — it matches issue subjects, not covered surfaces, so it reported no duplicates for #736 while missing open sibling #739, whose scope table already carries the current producer/consumer contract; a knowledge-graph search on the artifact path and tool names surfaces such adjacent siblings. Disposition per repo precedent: keep the issue VALID (the docs gap is real — user-facing docs pages grep to zero mentions — and gate issues are consolidation points, not duplicates), but re-scope the checklist to the mechanism-of-record tool's live schema before execution and land shared sections with the sibling issue.
