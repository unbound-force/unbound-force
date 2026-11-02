---
tag: session-relocation
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:46Z
identity: session-relocation-20261102T035446-jay-flowers
tier: draft
---

Triage of unbound-force/unbound-force#645 (RFC: session directory relocation via plugin for worktree isolation): when no public API exists to change an opencode session's project directory, the only viable mechanism is writing to opencode's private SQLite (session.directory — read fresh from the DB on every resume, per the issue's schema inspection). This couples the plugin to an undocumented internal schema: host-app upgrades can break relocation silently. Any implementation spec for this pattern must require: (1) runtime schema verification before the UPDATE, failing closed with an actionable error naming the mismatch; (2) a durable record of the original directory outside the mutated row (anchor file or session metadata) so worktree_move_back and manual recovery survive a half-completed move; (3) DB path resolution at runtime via the host's own CLI (`opencode db path`) instead of hardcoding ~/.local/share/opencode/opencode.db, since XDG_DATA_HOME/OPENCODE_CONFIG_DIR vary; (4) return-instructions restart (Strategy A) over kexec/exec process replacement (Strategy B), which races tool-result persistence. The issue's open questions about project_directory/git_worktree registration and MCP server directory binding are load-bearing — resolve them by prototyping before design freeze. Adjacent prior art in this repo (sandbox persistent-workdir changes) solves the same "OpenCode starts in the wrong directory" class of problem at container spawn time via --workdir/WORKSPACE; it does not cover live session relocation.
