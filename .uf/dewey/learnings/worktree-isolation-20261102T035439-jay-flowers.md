---
tag: worktree-isolation
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:39Z
identity: worktree-isolation-20261102T035439-jay-flowers
tier: draft
---

Triage verification of RFC #645 (opencode session relocation via plugin for worktree isolation, unbound-force/unbound-force#645) against the live opencode SQLite database confirmed `session.directory` is text NOT NULL — but revealed stronger coupling than the RFC's plan assumes: `session.project_id` is NOT NULL with an FK to `project(id)` ON DELETE CASCADE, `project.worktree` is NOT NULL, and a `project_directory(project_id, directory, type, strategy)` table exists. An UPDATE of `session.directory` alone leaves the session row inconsistent with its project binding; whether session resume resolves the directory solely from `session.directory` (the RFC's info.ts citation) or through `project_id`/`project.worktree` is the pivotal unknown, elevating the RFC's open question #2 to must-resolve before implementation. Security gaps to carry into the eventual spec: (1) the sketched `UPDATE session SET directory = '<path>' WHERE id = '<sessionID>'` is string-interpolated — the implementation must use parameterized statements (default.md SC-004); (2) the chosen Option 7 plugin writes opencode.db in-process while the server holds the database open — Option 6's verified procedure stops opencode first, so locking/corruption behavior needs explicit analysis (a validated Dewey learning on dewey's graph.db documents a second process opening a SQLite store held by a running server silently degrading); (3) the `name` argument flows into `git worktree add ../worktrees/<name> -b <name>` — needs traversal and shell-metacharacter validation before command construction; (4) the `.opencode_original_dir` anchor file's content becomes `worktree_move_back`'s destination path, so a tampered anchor redirects the session to an arbitrary directory — treat as untrusted input, not trusted state.
