---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T17:45:34Z
identity: agent-permissions-20261106T174534-jay-flowers
tier: draft
---

OpenCode agent bash permission allow rules are prefix globs matched against the command string: a trailing `*` grants every invocation sharing the prefix, including capabilities that contradict the persona's declared posture. In a persona with `edit: deny`, `"git diff*": "allow"` admits `git diff --output=<path>` (writes an arbitrary file — documented git option), shell-redirection writes (`git diff > <path>`), and `git difftool --extcmd=<cmd>` (executes an arbitrary external command; `difftool` shares the `git diff` prefix), all silently bypassing the edit denial. Read-only reviewer personas whose diffs arrive via the dispatch prompt or the read_diff_chunk manifest tool should use `bash: deny` (eight-of-nine Divisor persona precedent); if bash git access is genuinely required, allowlist fully-specified read-only invocations and add explicit deny rules for `*--output*` and `git difftool*`, after verifying deny-over-allow precedence in the permission evaluator. Prefix globs on multi-capability command families are never safe for read-only personas; note that even narrowed prefixes like `git diff --stat*` remain redirection-leaky (`git diff --stat --output=x` still matches), which is why exact-form allowlists plus deny rules are the minimum hardening.
