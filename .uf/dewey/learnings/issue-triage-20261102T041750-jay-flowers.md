---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:17:50Z
identity: issue-triage-20261102T041750-jay-flowers
tier: draft
---

Triage heuristic for unbound-force repositories: when an issue bundles integration/configuration work with documentation deliverables, recommend a split at triage time. Docs-only PRs are exempt from the /uf.review-council gate (AGENTS.md review-council exemptions) and qualify for quick-tier routing ('trivial fixes, chores, docs'), while CI/harness configuration work requires full council review and is typically coupled to increment sequencing. Bundling the two forces documentation onto the slower review path and attaches unrelated sequencing constraints to it. Observed while triaging unbound-force/unbound-force#584, which bundles FullSend validation_loop.script wiring (AC#6) with pre-commit hook documentation (AC#7).
