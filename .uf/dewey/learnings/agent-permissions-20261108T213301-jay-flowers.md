---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-08T21:33:01Z
identity: agent-permissions-20261108T213301-jay-flowers
tier: draft
---

Terminal metacharacter deny families in OpenCode agent bash permission maps are only a closed boundary for single-line commands. Under the repo's own engine model (last-match-wins globs compiled to `^pattern$` regexes with the `s` flag — instruction-lint-preflight.test.ts decide()), a trailing-star allow's `*` spans newlines, so a multi-line command whose first line carries the allowed prefix and whose second line is a metacharacter-free command (curl download, rm, etc.) matches the allow and no single-character deny. Before treating the deny family as the structural layer for a NEW allow, verify the engine's multi-line matching semantics empirically or extend the family with a literal-newline deny (YAML key `*\n*`: "deny", placed with the terminal denies and added to the guard test's `terminal` list plus a multi-line decide() case). The residual is documented for the linguist persona (issue #971), but each new allow surface — especially one whose arguments are judgment products composed from untrusted-derived analysis, like tool-inventory's `--match` words — re-opens it and deserves its own risk-affirmation, not inheritance of the prior acceptance.
