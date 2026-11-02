---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:15:16Z
identity: issue-triage-20261102T021516-jay-flowers
tier: draft
---

Documentation-gate issues filed as the recorded artifact of a spec task's gate assessment are deterministically triageable through a three-link verification chain: (1) provenance — the change's tasks.md completion note names the filed issue numbers (unbound-force unleash-deterministic-tools task 8.2 records "Filed both targets 2026-10-16: unbound-force/unbound-force#739 ... unbound-force/website#309"); (2) contract match — the issue's cited tool contracts match the repo's registered tool descriptions (write_unleash_summary's .uf/artifacts/unleash-summary/<branch>.json camelCase body vs. uf_finale_pr_context's documented consumption; uf_unleash_detect_resume's artifact-first dispatch-artifact derivation with artifacts authoritative over checkpoint state); (3) gap grep — the per-page scope table verifies with a single docs-tree grep: docs/usage.md and docs/architecture.md mention /uf.unleash but return zero matches for the Increment 2 tool names (gap persists), while docs/heroes.md's single high-level mention confirms its "no change" row. The chain also settles duplicate and cross-repo questions in one pass: sibling increment issues (#728, Increment 1) are adjacent scope rather than duplicates, and the org constitution's website-repo obligation is discharged by the separately filed website#309.
