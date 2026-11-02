---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:22:00Z
identity: issue-triage-20261102T032200-jay-flowers
tier: draft
---

Epic-child issue triage: a duplicate-issue search is insufficient, because later-phase changes under the same epic can deliver a child issue's scope under different tool names and plugin placements while citing only the epic number (originating_issue), leaving issue-level duplicate checks empty. Verified case: unbound-force issue #666 (wait_for_ci + check_mergeability in uf-workflow, parent epic #659) was delivered as uf_finale_ci_poll in the uf-finale plugin by openspec/changes/finale-token-reduction (Phase 3 of #659, all tasks complete, spec-review and code-review passed), leaving the issue stale-but-open. Triage verification pattern for epic-child issues: (1) grep the proposed tool/capability names repo-wide; (2) read the target command/plugin files to identify the current mechanism; (3) enumerate openspec/changes/ active and archive/ entries and match .openspec.yaml originating_issue against the issue's parent epic; (4) confirm tasks.md completion markers before treating the issue's scope as unimplemented.
