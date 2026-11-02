---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:47Z
identity: issue-triage-20261102T023547-jay-flowers
tier: draft
---

Docs-gate issues in the unbound-force meta repo triage cleanly on three verifiable signals: (1) the 'docs' label is present, per the AGENTS.md documentation gate that requires a docs issue for user-facing changes before PR merge; (2) the issue references an implementing PR on a convention-correct branch (opsx/<name> for OpenSpec, NNN-<name> for Speckit); (3) the named tool or behavior matches a registered contract — for agent-facing tools, the session's live tool registry plus the repo's openspec/changes/<name>/ spec deltas are the verification sources (issue #706's checkpoint/get_progress/clear_progress matched the FR-001..FR-003 deltas and the compiled/uf-workflow validated learning). Duplicate triage must compare the named feature, not the shared component: issues #739, #740, #741, and #757 all touch the uf-workflow plugin area yet none track documentation of the progress-state tools, so related-area candidates are not duplicates.
