---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:36:57Z
identity: issue-triage-20261102T023657-jay-flowers
tier: draft
---

Same-command-family docs issues in unbound-force triage as distinct, not duplicate, when they reference different originating changes: a keyword search on "/uf.finale" surfaces #710 (uf-finale plugin + uf-workflow git tools, finale-token-reduction, closes #659), #739 (/uf.unleash deterministic-tool migration + unleash summary, Increment 2 of umbrella #716), and #810 (uf_finale_base_state sync-with-base gate, rebase-upstream-before-push) — each maps to a different OpenSpec change, which the repo's one-docs-issue-per-user-facing-change documentation gate makes expected rather than duplicative. Discriminate on the referenced change (branch, "Closes #N", "Increment N of #M") and the named tool surface in the body, not the command name in the title. Scope-update comments that append later changes' behavior deltas keep a single docs issue authoritative across related fixes and are an objectivity signal, not drift: #710's fix-tui-git-output-leak comment (envelope-only tool output, uf_finale_stage returning {fileCount, files}, silent uf_finale_ci_poll) is independently corroborated by the finale-token-reduction tasks.md supersession annotation ("print full output to TUI ... superseded by fix-tui-git-output-leak: envelope-only output") and by the cross-repo website tracking issue unbound-force/website#306 — three artifacts recording the same delta confirms the comment reflects shipped behavior.
