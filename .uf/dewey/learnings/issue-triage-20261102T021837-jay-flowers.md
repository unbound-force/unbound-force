---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:18:37Z
identity: issue-triage-20261102T021837-jay-flowers
tier: draft
---

When triaging unbound-force documentation-gate issues whose body cites only the AGENTS.md current-repo gate, the org constitution's Cross-Repo Documentation rule (the unbound-force/website issue) is verifiable without leaving the session-scoped repo: Dewey's github-unbound-force source indexes the website repo's issues and the implementing PR bodies, which record both gate filings. Verified while triaging unbound-force/unbound-force#728 (unleash-direct-review-dispatch, Increment 1 of #716): the issue body names only its in-repo scope, but Dewey surfaced PR #729 ("Documentation gate: filed unbound-force/unbound-force#728 (repo docs issue) and unbound-force/website#307 (cross-repo)") and the change's tasks.md documentation-phase checkboxes (6.1 CHANGELOG entry, 6.2 -> #728, 6.3 -> website#307), establishing dual-gate compliance. Gate-issue provenance is triple-recorded — the issue itself, the change's tasks.md documentation phase, and the PR body — so any one record can confirm or refute gate compliance; consult Dewey before flagging a missing website-gate issue.
