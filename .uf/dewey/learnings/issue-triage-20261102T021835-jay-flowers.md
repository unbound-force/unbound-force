---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:18:35Z
identity: issue-triage-20261102T021835-jay-flowers
tier: draft
---

Unbound Force documentation-gate triage: a user-facing change discharges two distinct documentation obligations — the repo-local AGENTS.md documentation gate (an issue in the current repo tracking repo-local pages such as docs/usage.md and docs/heroes.md) and the org constitution's Cross-Repo Documentation rule (an issue in unbound-force/website). The unleash-direct-review-dispatch change demonstrates the companion pattern: unbound-force/unbound-force#728 and unbound-force/website#307 carry the same title in different repos and are companions, not duplicates — the website issue tracks the public site, the current-repo issue tracks repo-local docs pages. When triaging a docs-gate issue, verify against the gate text itself: (1) the change is user-facing and non-exempt (command/behavior changes count; spec artifacts are exempt), (2) the issue is filed pre-merge in the correct repo, (3) the named pages actually carry the stale content (exact-line verification — issue #728's 'line ~82' claim matched docs/heroes.md line 82 exactly), and (4) the originating change directory exists with a complete artifact set. A same-titled cross-repo sibling is evidence of complete gate discharge, not duplication.
