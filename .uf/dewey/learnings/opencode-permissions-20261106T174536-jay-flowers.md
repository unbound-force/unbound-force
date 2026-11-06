---
tag: opencode-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T17:45:36Z
identity: opencode-permissions-20261106T174536-jay-flowers
tier: draft
---

An opencode bash allow rule with a trailing `*` matches the command's full flag surface, including write flags the author never intended: `"git diff*": "allow"` also permits `git diff --output=<path>`, which writes an arbitrary file and bypasses `edit: deny` (edit governs the edit tool, not bash-side writes). opencode evaluates the LAST matching rule in a permission object, so the fix is a narrow deny placed after the broad allow — e.g. `"git diff --output*": "deny"` immediately after `"git diff*": "allow"` — not enumeration of exact read-only forms, which is brittle against flexible invocations. When authoring reviewer personas, scope allow rules to read-only-by-construction commands (the divisor-curator precedent: `gh repo view*` allowed, `gh issue create*` escalated to `ask`) or declare `bash: deny` and rely on caller-supplied diffs.
