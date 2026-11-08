---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:25:00Z
identity: agent-permissions-20261106T202500-jay-flowers
tier: draft
---

OpenCode bash allow rules are prefix globs matched against the whole command string, and the LAST matching rule wins. A trailing `*` therefore grants every invocation sharing the prefix, including capabilities that contradict the persona's declared posture: with `edit: deny`, `"git diff*": "allow"` admits `git diff --output=<path>` (an arbitrary-file write via a documented git flag), `git difftool --extcmd=<cmd>` (arbitrary execution; difftool shares the prefix), and shell redirection (`git diff > <path>`) — and narrowing the prefix does not help, because `"git diff --stat*"` still matches a trailing `--output=` argument. Hardening has two layers, both required: (1) narrow denies after the allow for flag-based vectors that carry no shell metacharacter (`git diff --output*`, `git difftool*`), and (2) terminal catch-all denies placed after EVERY allow on ANY prefix, one per metacharacter family — `"*>*"` and `"*<*"` (redirection and process substitution), `"*|*"` (pipes), `"*;*"` and `"*&*"` (separators and backgrounding), `"*$(*"` and backticks (command substitution) — so `git status > file` or `git diff | sh` dies on the catch-all even though the allow matched first. Enumerating exact read-only forms instead is brittle against flexible invocations. Read-only reviewer personas whose inputs arrive via the dispatch prompt or manifest tools should prefer `bash: deny` outright (Divisor precedent); when bash git access is genuinely required, scope allows to the read-only commands the persona's own documentation names and finish with the terminal denies.
