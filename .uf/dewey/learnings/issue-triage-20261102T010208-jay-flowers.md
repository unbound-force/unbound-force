---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:02:08Z
identity: issue-triage-20261102T010208-jay-flowers
tier: draft
---

Documentation-sync issue validation (unbound-force #789 triage): a docs-tracking issue for a landed user-facing change validates as VALID and objective when (1) its premise is corroborated by shipped artifacts — severity.md's "Human-directed fixes vs autonomous auto-fixes" clarifying note and the validated compiled article "compiled/spec-review > The Per-Finding Consent Gate" (issue #777, commit eeb7d09) confirm the per-finding spec-review decision gate landed, with autonomous auto-fixes bounded to LOW/MEDIUM and consented per-finding edits allowed at any severity; (2) it satisfies the repository's documentation gate (a current-repo docs issue is required for user-facing changes before PR merge); and (3) duplicate candidates are keyword-only matches with disjoint scope — #799 (OpenCode v2 migration strategic spec) shares "spec" terminology but not the decision-gate scope. Enumerated user-visible deltas (decision-option lists, exit-reason strings, severity ordering) provide verifiable completion criteria. Point-in-time spec text describing the superseded bulk consent gate is a historical record, not an update target; the issue correctly scopes its tasks to user-facing docs.
