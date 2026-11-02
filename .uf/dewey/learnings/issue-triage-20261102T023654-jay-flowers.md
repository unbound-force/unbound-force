---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:36:54Z
identity: issue-triage-20261102T023654-jay-flowers
tier: draft
---

Open documentation-tracking issues in unbound-force act as living behavior records: when a later change alters behavior that an open docs issue was filed to document (fix-tui-git-output-leak switching uf-workflow/uf-finale tool handlers to JSON-envelope-only output, uf_finale_stage returning {fileCount, files}, and silent uf_finale_ci_poll), the later change posts a scope-update comment to the existing issue (unbound-force/unbound-force#710, tracking the finale-token-reduction extraction) instead of filing a competing docs issue, and the originating change's tasks.md records the supersession inline ("superseded by fix-tui-git-output-leak: envelope-only output"). A triager therefore has two corroborating surfaces confirming the scope update is legitimate consolidation rather than scope creep. Sibling docs issues covering different increments of the same tool surface (#739 Increment 2 of #716, #810 uf_finale_base_state) are follow-ups, not duplicates, and the website twin (unbound-force/website#306) tracks cross-repo docs separately per the established documentation-gate pattern.
