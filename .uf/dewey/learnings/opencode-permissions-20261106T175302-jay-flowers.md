---
tag: opencode-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T17:53:02Z
identity: opencode-permissions-20261106T175302-jay-flowers
tier: draft
---

In OpenCode agent permission frontmatter, bash rules are glob patterns evaluated in insertion order with the LAST matching rule winning (broad rules first, narrow rules last). Consequences: (1) a prefix-glob allow like "git diff*": "allow" also permits write-capable flag forms such as `git diff --output=<path>`, bypassing `edit: deny`; (2) narrowing to "git diff --name-only*" does NOT close the hole, because the trailing `*` still matches appended flags (`git diff --name-only --output=x`); (3) a compensating deny (e.g. "*--output*": "deny") only takes effect as the FINAL entry of the bash map. Complete fixes: a last-position deny for the write-capable flag (after verifying the matcher supports mid-string globs), or exact-match patterns with no trailing `*` paired with updated command fallbacks in the agent body.
