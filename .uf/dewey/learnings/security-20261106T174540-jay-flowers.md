---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-06T17:45:40Z
identity: security-20261106T174540-jay-flowers
tier: draft
---

Read-only reviewer personas need their bash allowlists audited for write primitives hidden inside allowed commands. `"git diff*": "allow"` permits `git diff --output=<path>` — git writes the diff (which embeds untrusted PR lines) to the target file, silently defeating the persona's `edit: deny`. Narrowing to prefix patterns that keep a trailing `*` (`"git diff --stat*"`) does not close the hole: the `*` still matches a trailing ` --output=` argument. Use exact-match patterns with no trailing wildcard, or add deny rules only after verifying deny-over-allow precedence in the permission matcher and how it treats shell redirection (`git diff > <path>`).
