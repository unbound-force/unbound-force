---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:14:34Z
identity: documentation-gate-20261102T001434-jay-flowers
tier: draft
---

Documentation-gate issues for workflow-safety changes are strongest when their dispositions capture the destructive-operation guards, not just file names. unbound-force/unbound-force#810 (tracking the rebase-upstream-before-push OpenSpec change) is a well-formed example: it records that /uf.finale's new sync-with-base step pushes with git_push_force_with_lease only (never --force) after a rebase, aborts rebase conflicts via git_rebase_abort and stops, and degrades base-state check failures (e.g. offline fetch) to a warning rather than a hard block — then maps each artifact affected (CHANGELOG.md, docs/usage.md, docs/architecture.md), records explicit no-op determinations with rationale for AGENTS.md and README.md, and references the cross-repo website tracking issue (unbound-force/website#320). Triage tip: when a docs-gate issue states dispositions were "updated in the implementing PR", the issue is a verification record, not a work order — verify the claims against repository state (changelog entry present, doc step counts match, website issue exists, e.g. via the knowledge graph) instead of assuming pending documentation work.
