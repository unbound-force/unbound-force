---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:11Z
identity: issue-triage-20261102T025711-jay-flowers
tier: draft
---

Validating a documentation-gap issue requires verifying both sides: the behavior actually shipped (grep commands, plugins, and CHANGELOG) and the user-facing docs are actually silent (grep docs/ for the behavior's vocabulary). A CHANGELOG entry satisfies the changelog obligation but not the user-docs obligation — check docs/ separately. Base-ref resolution and prescribed command-output sections are user-facing operational behavior even when implemented as internal plugin tooling: they determine what local reviews diff against and what output users see. Verified on unbound-force issue #690 (companion to fix-stale-base-ref): resolve_base_ref's three-level fallback (upstream/main → origin/main → main, replacing hardcoded main in /uf.review-council's local path) and the Key Files Changed removal from the /uf.unleash demo shipped in PR #691, yet docs/ contains no base-ref documentation, and the fix-stale-base-ref CHANGELOG entry — unlike neighboring entries citing "Docs: #NNN" — carries no docs-issue reference, corroborating the open gap.
