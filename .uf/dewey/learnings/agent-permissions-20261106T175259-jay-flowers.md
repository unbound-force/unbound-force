---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T17:52:59Z
identity: agent-permissions-20261106T175259-jay-flowers
tier: draft
---

OpenCode bash permission patterns are prefix globs over the full command string: the repo's trailing-star idiom (`"git log*": "allow"`) matches a command with any arguments, and therefore also appended flags and same-prefix subcommands. `"git diff*": "allow"` consequently permits `git diff --output=<file>` (writes a file, bypassing edit:deny) and `git difftool --extcmd=<cmd>` (executes an arbitrary command). Star-suffixed "narrow" forms like `git diff --name-only*` remain extendable — an appended `--output` still matches. For deny-by-default personas that read untrusted PR prose: allow trailing-star forms only on commands with no write-capable flags or same-prefix subcommands (e.g. `git status*`); use exact star-free entries for commands like `git diff`; and update any instruction prose naming fallback commands in the same commit so the persona's own instructions never trip its permission wall.
