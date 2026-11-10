---
tag: injection-safety
author: jay-flowers
category: pattern
created_at: 2026-11-09T20:27:08Z
identity: injection-safety-20261109T202708-jay-flowers
tier: draft
---

When extracting shell choreography from command prose into a plugin tool (DE-001/DE-005), two distinct injection surfaces must be handled separately: (1) the tool's own process invocations — solved by argument-array executors (execFileSync with argv, no shell); and (2) the values the tool RETURNS, which consuming command prose may later render into shell templates. The checkout_pr_head pattern (PR #1002, issue #965) handles both: argv arrays for its own git/gh calls, plus a conservative charset gate (A-Za-z0-9._/- for refnames; plain-name or userinfo-free https URL for remotes, rejecting `.`/`..`) on every value admitted to the success payload. Key detail: JSON.stringify does NOT neutralize `$` or backtick — both are legal git refname characters that stay active inside double-quoted shell strings — so JSON-escaping an interpolated refname is not a substitute for a charset check. Order matters too: validate the branch name before using it in config keys and before returning it, and keep every pre-validation path fail-closed (STOP) so unvalidated values can never reach a rendered template.
