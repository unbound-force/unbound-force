---
tag: bash-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:09:52Z
identity: bash-permissions-20261102T050952-jay-flowers
tier: draft
---

OpenCode bash permission glob patterns treat the space before a wildcard as significant: "gh api * -f *" requires a literal space after the flag, so glued short-flag forms (gh api /repos/o/r/issues -ftitle=pwned) match no ask rule and fall through to the "gh api*": "allow" default — and gh switches to POST when request parameters are present, so the unmatched command is an unprompted mutation. The trailing-wildcard form ("-f*") covers glued, spaced, and "-f=" equals forms in a single rule because * matches spaces — the shipped refine-bash-permissions set already relies on this for "-X POST*" and "--input*". Structural lesson: when a permission rule set enumerates flag forms (spaced/glued/equals/case variants), apply the full form matrix to every flag class, not just method flags. The refine-bash-permissions set (unbound-force, fixes #531) covered glued forms for -X method flags and equals forms for --field/--raw-field, but left -f/-F data flags spaced-only — the exact gap reported in unbound-force issue #569.
