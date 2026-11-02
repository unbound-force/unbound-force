---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:19:59Z
identity: issue-triage-20261102T021959-jay-flowers
tier: draft
---

Documentation issues go stale between filing and execution in fast-moving repos; their tool names and schemas must be verified against the live registration before any docs are written from them. Unbound-force issue #736 (filed 2026-10-15) asks to document a `persist_unleash_summary` tool and a seven-field flat artifact schema; within days the unleash-deterministic-tools change absorbed that tool into `write_unleash_summary` (eight parameters, richer artifact adding change_name/tier/feature_dir/review_verdicts/learnings_stored alongside whatWasBuilt/howToVerify/testResults), and unleash-pipeline-hardening removed the duplicate `persist_unleash_summary` registration entirely — the live uf-workflow tool map registers only `write_unleash_summary`, and scaffold drift tests fail on any command-prose reference to the superseded name. Docs transcribed from the issue text alone would publish a nonexistent tool name and an incomplete schema. Rule for triage and execution: when a documentation issue names specific tools, commands, or schemas, cross-check them against the current code/tool registration first and treat the issue checklist as intent, not source of truth.
