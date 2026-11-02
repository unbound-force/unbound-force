---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:18:40Z
identity: issue-triage-20261102T021840-jay-flowers
tier: draft
---

Triage pattern for documentation-gate issues in unbound-force/unbound-force: a behavior-only change to command markdown (no Go/TypeScript code) is still user-facing and discharges TWO documentation gates — the AGENTS.md current-repo gate (an issue tracking the repo's own docs/ pages) and the org constitution's Cross-Repo Documentation rule (a parallel issue in unbound-force/website). A valid docs-tracking issue names the triggering OpenSpec change directory, cites the affected pages with the stale behavior they currently document, and is filed before PR merge. Verification recipe: (1) the named change directory exists and its tasks.md references the issue number, (2) the cited pages contain the claimed stale text at the claimed locations, (3) the pages do not yet mention the new behavior — that absence is the gap being tracked, (4) the website-repo counterpart exists separately and is complementary, not a duplicate. Verified on unleash-direct-review-dispatch (Increment 1 of #716): unbound-force#728 ↔ tasks.md task 6.2, docs/usage.md:294 "Run the full pipeline autonomously", docs/heroes.md:82 "Autonomous pipeline execution via /uf.unleash" (issue said "~line 82" — exact), website#307 ↔ tasks.md task 6.3.
