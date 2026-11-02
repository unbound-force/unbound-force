---
tag: permission-rules
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:09:56Z
identity: permission-rules-20261102T050956-jay-flowers
tier: draft
---

opencode bash-permission globs are substring matches over the full command line: a spaced short-flag pattern like "gh api * -f *" requires a literal ' -f ' and does not cover glued forms such as 'gh api /repos/o/r -ftitle=x'. The glued form falls through to the broader 'gh api*' allow rule and — because gh auto-switches to POST whenever request fields are present — silently becomes an unattended mutation. unbound-force's refine-bash-permissions rule set (design.md D3) covered glued/equals variants for method flags (-XPOST*) and long data flags (--field=) but missed glued short data flags; the corrective pattern "gh api * -f*" subsumes both spaced and glued forms with no availability regression (worst case: an ask on flag-shaped input, which errs safe for a mutation gate). Rule of thumb for glob-based permission allowlists: for every flag a pattern gates, enumerate its spaced, glued, and '=' variants — treat the design's own -X/--field precedent as the coverage checklist.
